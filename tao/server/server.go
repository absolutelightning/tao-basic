package main

import (
	"context"
	"fmt"
	pb "github.com/absolutelightning/tao-basic/tao/proto"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"slices"
	"time"
)

type Object struct {
	Id    string
	Otype string
	Data  map[string]interface{}
}

type Association struct {
	Id1       string
	Atype     string
	Id2       string
	Timestamp time.Time
}

func (s *Server) ObjectAdd(ctx context.Context, in *pb.ObjectAddRequest) (*pb.GenericOkResponse, error) {
	// Inserting in postgres
	obj := &Object{
		Id:    in.Id,
		Otype: in.Otype,
	}
	if in.Id == "" {
		uid, _ := uuid.NewUUID()
		in.Id = uid.String()
	}
	obj.Data = make(map[string]interface{})
	for _, kv := range in.Data {
		obj.Data[kv.Key] = kv.Value
	}
	_, err := s.pgDB.Model(obj).Insert()
	if err != nil {
		return nil, err
	}

	// Use Redis pipeline for batching commands
	pipe := s.redisClient.Pipeline()

	// Set Otype in Redis
	pipe.HSet(ctx, in.Id, "Otype", in.Otype)

	// Set the data fields in Redis
	for _, kv := range in.Data {
		pipe.HSet(ctx, in.Id, kv.Key, kv.Value)
	}

	// Execute the pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.GenericOkResponse{}, nil
}

func (s *Server) AssocAdd(ctx context.Context, in *pb.AssocAddRequest) (*pb.GenericOkResponse, error) {
	// Insert the association into PostgreSQL
	assoc := &Association{
		Id1:       in.Id1,
		Id2:       in.Id2,
		Atype:     in.Atype,
		Timestamp: time.Now(),
	}
	_, err := s.pgDB.Model(assoc).Insert()
	if err != nil {
		return nil, err
	}

	// Reverse relation - CAN BE EXECUTED IN BACKGROUND
	reverseAssoc := &Association{
		Id1:       in.Id2,
		Id2:       in.Id1,
		Atype:     in.Atype,
		Timestamp: time.Now(),
	}
	_, err = s.pgDB.Model(reverseAssoc).Insert()
	if err != nil {
		return nil, err
	}

	// Start Redis pipeline
	pipe := s.redisClient.Pipeline()

	// Add association to Redis sorted set
	assocSetName := fmt.Sprintf("%s-%s", in.Id1, in.Atype)
	member := redis.Z{
		Score:  float64(assoc.Timestamp.UnixNano()), // Use UnixNano for better timestamp precision
		Member: in.Id2,
	}
	pipe.ZAdd(ctx, assocSetName, member)

	// Add reverse association to Redis sorted set
	reverseAssocSetName := fmt.Sprintf("%s-%s", in.Id2, in.Atype)
	reverseMember := redis.Z{
		Score:  float64(reverseAssoc.Timestamp.UnixNano()), // Use UnixNano for better timestamp precision
		Member: in.Id1,
	}
	pipe.ZAdd(ctx, reverseAssocSetName, reverseMember)

	// Execute the pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.GenericOkResponse{}, nil
}

func (s *Server) AssocGet(ctx context.Context, in *pb.AssocGetRequest) (*pb.AssocGetResponse, error) {
	assocSetName := fmt.Sprintf("%s-%s", in.Id1, in.Atype)
	var res []string
	opts := &redis.ZRangeBy{
		Min: "-inf",
		Max: "+inf",
	}
	if in.Low != nil && in.High != nil {
		opts = &redis.ZRangeBy{
			Min: *in.Low,
			Max: *in.High,
		}
	}
	results, err := s.redisClient.ZRangeByScore(ctx, assocSetName, opts).Result()
	if err != nil {
		return nil, err
	}
	res = results
	assocIds := make([]string, len(res))
	for _, v := range res {
		if slices.Contains(in.Id2, v) {
			assocIds = append(assocIds, v)
		}
	}
	pipe := s.redisClient.Pipeline()
	cmds := make([]*redis.MapStringStringCmd, len(assocIds))
	for i, v := range assocIds {
		cmds[i] = pipe.HGetAll(ctx, v)
	}
	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}
	resMap := make(map[string]map[string]string, len(assocIds))
	for i, cmd := range cmds {
		if err = cmd.Err(); err != nil {
			return nil, err
		}
		resMap[assocIds[i]] = cmd.Val()
	}

	assocGetResp := &pb.AssocGetResponse{
		Objects: make([]*pb.Object, len(assocIds)),
	}
	itr := 0
	for k, v := range resMap {
		assocGetResp.Objects[itr] = &pb.Object{
			Id:    k,
			Items: make([]*pb.KeyValuePair, 0),
		}
		for key, val := range v {
			assocGetResp.Objects[itr].Items = append(assocGetResp.Objects[itr].Items, &pb.KeyValuePair{
				Key:   key,
				Value: val,
			})
		}
		itr++
	}
	return assocGetResp, nil
}

func (s *Server) ObjectGet(ctx context.Context, in *pb.ObjectGetRequest) (*pb.AssocGetResponse, error) {

	var setKeys []string
	allOtype, err := s.redisClient.SMembers(ctx, "Otype:"+in.Otype).Result()
	if err != nil {
		return nil, err
	}
	if in.Data != nil {
		for _, kv := range in.Data.Data {
			setKey := fmt.Sprintf("%s:%s", kv.Key, kv.Value)
			setKeys = append(setKeys, setKey)
		}
	}

	allSetKeys := make([]string, 0)
	for _, v := range setKeys {
		allSetKeys = append(allSetKeys, v)
	}
	for _, v := range allOtype {
		allSetKeys = append(allSetKeys, v)
	}

	ids, err := s.redisClient.SInter(ctx, allSetKeys...).Result()

	if err != nil {
		return nil, err
	}

	allData := make([]map[string]string, len(ids))
	for _, id := range ids {
		data, hgeterr := s.redisClient.HGetAll(ctx, id).Result()
		if hgeterr != nil {
			return nil, hgeterr
		}
		allData = append(allData, data)
	}

	assocGetResp := &pb.AssocGetResponse{
		Objects: make([]*pb.Object, len(ids)),
	}
	for i, model := range ids {
		assocGetResp.Objects[i] = &pb.Object{
			Id:    model,
			Items: make([]*pb.KeyValuePair, 0),
		}
		for j, v := range allData[i] {
			assocGetResp.Objects[i].Items = append(assocGetResp.Objects[i].Items, &pb.KeyValuePair{
				Key:   j,
				Value: v,
			})
		}
	}
	return assocGetResp, nil
}

func (s *Server) AssocRange(ctx context.Context, in *pb.AssocRangeRequest) (*pb.AssocGetResponse, error) {
	assocSetName := fmt.Sprintf("%s-%s", in.Id1, in.Atype)
	res, err := s.redisClient.ZRange(ctx, assocSetName, in.Pos, in.Pos+in.Limit).Result()
	if err != nil {
		return nil, err
	}
	assocIds := res
	pipe := s.redisClient.Pipeline()
	cmds := make([]*redis.MapStringStringCmd, len(assocIds))
	for i, v := range res {
		cmds[i] = pipe.HGetAll(ctx, v)
	}
	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}
	results := make(map[string]map[string]string, len(assocIds))
	for i, cmd := range cmds {
		if err = cmd.Err(); err != nil {
			return nil, err
		}
		results[assocIds[i]] = cmd.Val()
	}

	assocGetResp := &pb.AssocGetResponse{
		Objects: make([]*pb.Object, len(assocIds)),
	}
	itr := 0
	for k, v := range results {
		assocGetResp.Objects[itr] = &pb.Object{
			Id:    k,
			Items: make([]*pb.KeyValuePair, 0),
		}
		for key, val := range v {
			assocGetResp.Objects[itr].Items = append(assocGetResp.Objects[itr].Items, &pb.KeyValuePair{
				Key:   key,
				Value: val,
			})
		}
		itr++
	}
	return assocGetResp, nil
}

func (s *Server) BulkAssocAdd(ctx context.Context, in *pb.BulkAssocAddRequest) (*pb.GenericOkResponse, error) {
	// Insert the association into PostgreSQL
	assocs := make([]Association, 0)
	for _, req := range in.Req {
		assocs = append(assocs, Association{
			Id1:       req.Id1,
			Id2:       req.Id2,
			Atype:     req.Atype,
			Timestamp: time.Now(),
		})
	}
	_, err := s.pgDB.Model(&assocs).Insert()
	if err != nil {
		return nil, err
	}

	// Reverse relation - CAN BE EXECUTED IN BACKGROUND
	reverseAssocs := make([]Association, 0)
	for _, req := range in.Req {
		reverseAssocs = append(reverseAssocs, Association{
			Id1:       req.Id2,
			Id2:       req.Id1,
			Atype:     req.Atype,
			Timestamp: time.Now(),
		})
	}
	_, err = s.pgDB.Model(&reverseAssocs).Insert()
	if err != nil {
		return nil, err
	}

	// Start Redis pipeline
	pipe := s.redisClient.Pipeline()

	// Add association to Redis sorted set
	assocData := make(map[string][]redis.Z)
	for _, assoc := range assocs {
		assocSetName := fmt.Sprintf("%s-%s", assoc.Id1, assoc.Atype)
		member := redis.Z{
			Score:  float64(assoc.Timestamp.UnixNano()), // Use UnixNano for better timestamp precision
			Member: assoc.Id2,
		}
		if _, ok := assocData[assocSetName]; !ok {
			assocData[assocSetName] = make([]redis.Z, 0)
		}
		assocData[assocSetName] = append(assocData[assocSetName], member)
	}
	for k, v := range assocData {
		pipe.ZAdd(ctx, k, v...)
	}

	// Add reverse association to Redis sorted set
	reverseAssocData := make(map[string][]redis.Z)
	for _, assoc := range reverseAssocs {
		assocSetName := fmt.Sprintf("%s-%s", assoc.Id1, assoc.Atype)
		member := redis.Z{
			Score:  float64(assoc.Timestamp.UnixNano()), // Use UnixNano for better timestamp precision
			Member: assoc.Id2,
		}
		if _, ok := assocData[assocSetName]; !ok {
			reverseAssocData[assocSetName] = make([]redis.Z, 0)
		}
		reverseAssocData[assocSetName] = append(assocData[assocSetName], member)
	}
	for k, v := range reverseAssocData {
		pipe.ZAdd(ctx, k, v...)
	}

	// Execute the pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.GenericOkResponse{}, nil
}

func (s *Server) BulkObjectAdd(ctx context.Context, in *pb.BulkObjectAddRequest) (*pb.GenericOkResponse, error) {
	objects := make([]Object, 0)
	for _, req := range in.Req {
		obj := Object{
			Id:    req.Id,
			Otype: req.Otype,
		}
		obj.Data = make(map[string]interface{})
		for _, kv := range req.Data {
			obj.Data[kv.Key] = kv.Value
		}
		objects = append(objects, obj)
	}
	_, err := s.pgDB.Model(&objects).Insert()
	if err != nil {
		return nil, err
	}

	// Use Redis pipeline for batching commands
	pipe := s.redisClient.Pipeline()

	for _, req := range in.Req {
		// Set Otype in Redis
		pipe.HSet(ctx, req.Id, "Otype", req.Otype)

		// Set the data fields in Redis
		for _, kv := range req.Data {
			pipe.HSet(ctx, req.Id, kv.Key, kv.Value)
			setKey := fmt.Sprintf("%s:%s", kv.Key, kv.Value)
			pipe.SAdd(ctx, setKey, req.Id)
		}
		pipe.SAdd(ctx, "Otype:"+req.Otype, req.Id)
	}

	// Execute the pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.GenericOkResponse{}, nil
}
