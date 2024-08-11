package main

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"

	pb "github.com/absolutelightning/tao-basic/tao/proto"
)

var serverAddr string = "localhost:7051"

func main() {
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("Failed to dial server: %v", err)
	}

	c := pb.NewTaoServiceClient(conn)

	kvData := make([]*pb.KeyValuePair, 1)
	kvData[0] = &pb.KeyValuePair{}
	kvData[0].Key = "batman"
	kvData[0].Value = "bruce wayne"

	_, err = c.ObjectAdd(context.Background(), &pb.ObjectAddRequest{
		Id:    "3",
		Otype: "ws",
		Data:  kvData,
	})
	if err != nil {
		panic(err)
	}

	kvData = make([]*pb.KeyValuePair, 1)
	kvData[0] = &pb.KeyValuePair{}
	kvData[0].Key = "superman"
	kvData[0].Value = "clark kent"
	_, err = c.ObjectAdd(context.Background(), &pb.ObjectAddRequest{
		Id:    "4",
		Otype: "org",
		Data:  kvData,
	})
	if err != nil {
		panic(err)
	}

	_, err = c.AssocAdd(context.Background(), &pb.AssocAddRequest{
		Id1:   "3",
		Id2:   "4",
		Atype: "ws-org",
	})
	if err != nil {
		panic(err)
	}

	defer conn.Close()
}
