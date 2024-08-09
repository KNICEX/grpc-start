package main

import (
	"fmt"
	"github.com/KNICEX/grpc-start/pb"
	"google.golang.org/grpc"
	"io"
	"net"
	"strconv"
	"sync"
)

type Service struct {
	pb.UnimplementedEchoServer
}

func (s Service) ServerStreamingEcho(req *pb.EchoReq, g grpc.ServerStreamingServer[pb.EchoResp]) error {
	msg := req.Message
	for i := 0; i < 10; i++ {
		if err := g.Send(&pb.EchoResp{Message: msg + strconv.Itoa(i)}); err != nil {
			return err
		}
	}
	return nil
}

func (s Service) ClientStreamingEcho(g grpc.ClientStreamingServer[pb.EchoReq, pb.EchoResp]) error {
	i := 0
	for {
		req, err := g.Recv()
		if err == io.EOF {
			// 客户端发送完毕
			return g.SendAndClose(&pb.EchoResp{Message: "ClientStreamingEcho"})
		}
		if err != nil {
			return err
		}
		i++
		fmt.Println("ClientStreamingEcho", req.Message)
	}
}

func (s Service) BidirectionalStreamingEcho(g grpc.BidiStreamingServer[pb.EchoReq, pb.EchoResp]) error {
	msgChan := make(chan string)
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			req, err := g.Recv()
			if err == io.EOF {
				close(msgChan)
				return
			}
			if err != nil {
				close(msgChan)
				return
			}
			msgChan <- req.Message
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for msg := range msgChan {
			err := g.Send(&pb.EchoResp{Message: msg + " reply"})
			if err != nil {
				fmt.Println("send error ", err)
			}
		}
	}()
	wg.Wait()
	return nil
}

func main() {
	l, err := net.Listen("tcp", ":50051")
	if err != nil {
		panic(err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterEchoServer(grpcServer, Service{})
	fmt.Println("Server is running on port :50051")

	if err = grpcServer.Serve(l); err != nil {
		panic(err)
	}
}
