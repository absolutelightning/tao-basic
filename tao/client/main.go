package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"

	pb "github.com/absolutelightning/tao-basic/tao/proto"
)

var serverAddr string = "localhost:7051"

func main() {
	ops := grpc.WithDefaultCallOptions(
		grpc.MaxCallSendMsgSize(1000000000000),
		grpc.MaxCallSendMsgSize(1000000000000))
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()), ops)

	if err != nil {
		log.Fatalf("Failed to dial server: %v", err)
	}

	c := pb.NewTaoServiceClient(conn)

	kvData := make([]*pb.KeyValuePair, 1)
	kvData[0] = &pb.KeyValuePair{}
	kvData[0].Key = "batman"
	kvData[0].Value = "bruce wayne"

	_, err = c.ObjectAdd(context.Background(), &pb.ObjectAddRequest{
		Id:    "1",
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
		Id:    "2",
		Otype: "org",
		Data:  kvData,
	})
	if err != nil {
		panic(err)
	}

	_, err = c.AssocAdd(context.Background(), &pb.AssocAddRequest{
		Id1:   "1",
		Id2:   "2",
		Atype: "ws-org",
	})
	if err != nil {
		panic(err)
	}

	kvData = make([]*pb.KeyValuePair, 1)
	kvData[0] = &pb.KeyValuePair{}
	kvData[0].Key = "superman"
	kvData[0].Value = "clark kent"
	fmt.Println(c.ObjectGet(context.Background(), &pb.ObjectGetRequest{
		Otype: "org",
		Data: &pb.ObjectGetDataRequest{
			Data: kvData,
		},
	}))

	kvData = make([]*pb.KeyValuePair, 1)
	kvData[0] = &pb.KeyValuePair{}
	kvData[0].Key = "batman"
	kvData[0].Value = "bruce wayne"
	fmt.Println(c.ObjectGet(context.Background(), &pb.ObjectGetRequest{
		Otype: "ws",
		Data: &pb.ObjectGetDataRequest{
			Data: kvData,
		},
	}))

	fmt.Println(c.AssocRange(context.Background(), &pb.AssocRangeRequest{
		Id1:   "1",
		Atype: "ws-org",
		Pos:   0,
		Limit: -1,
	}))

	fmt.Println(c.AssocRange(context.Background(), &pb.AssocRangeRequest{
		Id1:   "2",
		Atype: "ws-org",
		Pos:   0,
		Limit: -1,
	}))

	defer conn.Close()
}
