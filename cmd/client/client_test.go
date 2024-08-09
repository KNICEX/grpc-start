package main

import (
	"context"
	"github.com/KNICEX/grpc-start/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"testing"
)

func newClient() pb.EchoClient {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	grpcClient = pb.NewEchoClient(conn)
	return grpcClient
}

func TestServerStreamingEcho(t *testing.T) {
	client := newClient()
	sc, err := client.ServerStreamingEcho(context.Background(), &pb.EchoReq{
		Message: "hello",
	})
	if err != nil {
		t.Fatal(err)
	}

	for {
		resp, err := sc.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Log(err)
			continue
		}
		t.Log("receive message:", resp.Message)
	}

}

func TestClientStreamingEcho(t *testing.T) {
	client := newClient()
	cc, err := client.ClientStreamingEcho(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		if err := cc.Send(&pb.EchoReq{Message: "hello"}); err != nil {
			t.Fatal(err)
		}
	}
	resp, err := cc.CloseAndRecv()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("receive message:", resp.Message)
}

func TestBidirectionalStreamingEcho(t *testing.T) {
	client := newClient()
	bc, err := client.BidirectionalStreamingEcho(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		if err := bc.Send(&pb.EchoReq{Message: "hello"}); err != nil {
			t.Fatal(err)
		}
		resp, err := bc.Recv()
		if err != nil {
			t.Fatal(err)
		}
		t.Log("receive message:", resp.Message)
	}
	if err = bc.CloseSend(); err != nil {
		t.Fatal(err)
	}

}
