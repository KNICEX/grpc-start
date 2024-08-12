package main

import (
	"context"
	"github.com/KNICEX/grpc-start/pb"
	"github.com/stretchr/testify/require"
	"io"
	"testing"
)

func TestServerStreamingEcho(t *testing.T) {
	client := newClient()
	sc, err := client.ServerStreamingEcho(context.Background(), &pb.EchoReq{
		Message: "hello",
	})
	require.NoError(t, err)

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
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		if err := cc.Send(&pb.EchoReq{Message: "hello"}); err != nil {
			t.Fatal(err)
		}
	}
	resp, err := cc.CloseAndRecv()
	require.NoError(t, err)
	t.Log("receive message:", resp.Message)
}

func TestBidirectionalStreamingEcho(t *testing.T) {
	client := newClient()
	bc, err := client.BidirectionalStreamingEcho(context.Background())
	require.NoError(t, err)

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

func TestBatch(t *testing.T) {
	t.Run("TestServerStreamingEcho", TestServerStreamingEcho)
	t.Run("TestClientStreamingEcho", TestClientStreamingEcho)
	t.Run("TestBidirectionalStreamingEcho", TestBidirectionalStreamingEcho)
}
