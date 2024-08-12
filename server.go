package grpc_start

import (
	"context"
	"github.com/KNICEX/grpc-start/register"
	"google.golang.org/grpc"
	"net"
	"time"
)

type Server struct {
	name string
	*grpc.Server
	r register.Register
}

func NewServer(name string, opts ...ServerOption) *Server {
	res := &Server{
		name:   name,
		Server: grpc.NewServer(),
	}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

type ServerOption func(server *Server)

func ServerWithRegistry(r register.Register) ServerOption {
	return func(server *Server) {
		server.r = r
	}
}

func (s *Server) Start(addr string) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	if s.r != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		err = s.r.Register(ctx, register.ServiceInstance{
			ServiceName: s.name,
			Address:     l.Addr().String(),
		})
		if err != nil {
			return err
		}
	}
	return s.Serve(l)
}
