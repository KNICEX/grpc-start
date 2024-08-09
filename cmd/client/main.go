package main

import (
	"github.com/KNICEX/grpc-start/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var grpcClient pb.EchoClient

func main() {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	grpcClient = pb.NewEchoClient(conn)

}
