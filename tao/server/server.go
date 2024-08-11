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
	err = s.redisClient.HSet(ctx, in.Id, "Otype", in.Otype).Err()
	if err != nil {
		return nil, err
	}
	for _, kv := range in.Data {
		err = s.redisClient.HSet(ctx, in.Id, kv.Key, kv.Value).Err()
		if err != nil {
			return nil, err
		}
	}
	return &pb.GenericOkResponse{}, nil
}

func (s *Server) AssocAdd(ctx context.Context, in *pb.AssocAddRequest) (*pb.GenericOkResponse, error) {
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
	assoc = &Association{
		Id1:       in.Id2,
		Id2:       in.Id1,
		Atype:     in.Atype,
		Timestamp: time.Now(),
	}
	_, err = s.pgDB.Model(assoc).Insert()
	if err != nil {
		return nil, err
	}
	assocSetName := fmt.Sprintf("%s-%s", in.Id1, in.Atype)
	member := redis.Z{
		Score:  float64(assoc.Timestamp.Nanosecond()),
		Member: in.Id2,
	}
	err = s.redisClient.ZAdd(ctx, assocSetName, member).Err()
	if err != nil {
		return nil, err
	}
	// Reverse
	assocSetName = fmt.Sprintf("%s-%s", in.Id2, assoc.Atype)
	member = redis.Z{
		Score:  float64(assoc.Timestamp.Nanosecond()),
		Member: in.Id1,
	}
	err = s.redisClient.ZAdd(ctx, assocSetName, member).Err()
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
	var models []Object
	query := s.pgDB.Model(&models).
		Where("otype = ?", in.Otype).
		Limit(int(in.Limit))

	if in.Data != nil {
		for _, kv := range in.Data.Data {
			query = query.Where(fmt.Sprintf("data->>'%s'='%s'", kv.Key, kv.Value))
		}
	}

	err := query.Select()
	if err != nil {
		return nil, err
	}

	assocGetResp := &pb.AssocGetResponse{
		Objects: make([]*pb.Object, len(models)),
	}
	for i, model := range models {
		assocGetResp.Objects[i] = &pb.Object{
			Id:    model.Id,
			Items: make([]*pb.KeyValuePair, 0),
		}
		for j, v := range models[i].Data {
			assocGetResp.Objects[i].Items = append(assocGetResp.Objects[i].Items, &pb.KeyValuePair{
				Key:   j,
				Value: v.(string),
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
	for _, req := range in.Req {
		_, err := s.AssocAdd(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return &pb.GenericOkResponse{}, nil
}
