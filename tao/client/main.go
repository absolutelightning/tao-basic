package main

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"

	pb "github.com/absolutelightning/tao-basic/tao/proto"
)

var serverAddr string = "localhost:50051"

func main() {
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("Failed to dial server: %v", err)
	}

	c := pb.NewTaoServiceClient(conn)

	kvData := make([]*pb.KeyValuePair, 1)
	kvData[0].Key = "batman"
	kvData[1].Value = "bruce wayne"

	c.ObjectAdd(context.Background(), &pb.ObjectAddRequest{
		Id:    "1",
		Otype: "ws",
		Data:  kvData,
	})

	kvData = make([]*pb.KeyValuePair, 1)
	kvData[0].Key = "superman"
	kvData[1].Value = "clark kent"
	c.ObjectAdd(context.Background(), &pb.ObjectAddRequest{
		Id:    "2",
		Otype: "org",
		Data:  kvData,
	})

	c.AssocAdd(context.Background(), &pb.AssocAddRequest{
		Id1:   "1",
		Id2:   "2",
		Atype: "ws-org",
	})

	defer conn.Close()
}
