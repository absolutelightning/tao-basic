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
	assocSetName := fmt.Sprintf("%s-%s", in.Id1, assoc.Atype)
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
	low := float32(0.0)
	if in.Low != nil {
		low = *in.Low
	}
	high := float32(0.0)
	if in.High != nil {
		low = *in.High
	}
	assocSetName := fmt.Sprintf("%s-%s", in.Id1, in.Atype)
	res, err := s.redisClient.ZRange(ctx, assocSetName, int64(low), int64(high)).Result()
	if err != nil {
		return nil, err
	}
	assocIds := make([]string, len(res))
	for _, v := range res {
		if slices.Contains(in.Id2, v) {
			assocIds = append(assocIds, v)
		}
	}
	assocGetResp := &pb.AssocGetResponse{
		Objects: make([]*pb.Object, len(assocIds)),
	}
	for i, id := range assocIds {
		data := s.redisClient.HGetAll(ctx, id).Val()
		assocGetResp.Objects[i] = &pb.Object{
			Id:    id,
			Items: make([]*pb.KeyValuePair, len(data)),
		}
		for j, v := range data {
			assocGetResp.Objects[i].Items = append(assocGetResp.Objects[i].Items, &pb.KeyValuePair{
				Key:   j,
				Value: v,
			})
		}
	}
	return assocGetResp, nil
}

func (s *Server) ObjectGet(ctx context.Context, in *pb.ObjectGetRequest) (*pb.AssocGetResponse, error) {
	var models []Object
	s.pgDB.Model(&models).
		Where("otype = ?", in.Otype).
		Limit(int(in.Limit)).
		Select()

	assocGetResp := &pb.AssocGetResponse{
		Objects: make([]*pb.Object, len(models)),
	}
	for i, model := range models {
		data := s.redisClient.HGetAll(ctx, model.Id).Val()
		assocGetResp.Objects[i] = &pb.Object{
			Id:    model.Id,
			Items: make([]*pb.KeyValuePair, 0),
		}
		for j, v := range data {
			assocGetResp.Objects[i].Items = append(assocGetResp.Objects[i].Items, &pb.KeyValuePair{
				Key:   j,
				Value: v,
			})
		}
	}
	return assocGetResp, nil
}
