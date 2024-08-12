package main

import (
	grpc_start "github.com/KNICEX/grpc-start"
	"github.com/KNICEX/grpc-start/pb"
	"github.com/KNICEX/grpc-start/register/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"time"
)

var grpcClient pb.EchoClient

func newClient() pb.EchoClient {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:12379"},
		DialTimeout: time.Second * 3,
	})
	if err != nil {
		panic(err)
	}
	r, err := etcd.NewRegistry(etcdClient)
	if err != nil {
		panic(err)
	}
	rs := grpc_start.NewResolverBuilder(r, time.Second*5)
	conn, err := grpc.NewClient("registry:///test", grpc.WithResolvers(rs), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	grpcClient = pb.NewEchoClient(conn)
	return grpcClient
}

func main() {

}
