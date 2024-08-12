package grpc_start

import (
	"context"
	"github.com/KNICEX/grpc-start/register"
	"google.golang.org/grpc/resolver"
	"log"
	"time"
)

type grpcResolverBuilder struct {
	r       register.Register
	timeout time.Duration
}

func NewResolverBuilder(r register.Register, timeout time.Duration) resolver.Builder {
	return &grpcResolverBuilder{
		r:       r,
		timeout: timeout,
	}
}

func (g *grpcResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {

	res := &grpcResolver{
		target:  target,
		cc:      cc,
		r:       g.r,
		timeout: g.timeout,
		close:   make(chan struct{}),
	}
	res.resolve()
	return res, res.watch()
}

func (g *grpcResolverBuilder) Scheme() string {
	return "registry"
}

type grpcResolver struct {
	target resolver.Target
	cc     resolver.ClientConn
	r      register.Register

	state   resolver.State
	timeout time.Duration
	close   chan struct{}
}

func (g *grpcResolver) resolve() resolver.State {
	ctx, cancel := context.WithTimeout(context.Background(), g.timeout)
	defer cancel()
	instances, err := g.r.ListService(ctx, g.target.Endpoint())
	if err != nil {
		g.cc.ReportError(err)
		return resolver.State{}
	}
	addresses := make([]resolver.Address, len(instances))
	for i, addr := range instances {
		addresses[i] = resolver.Address{
			Addr: addr.Address,
		}
	}

	g.state = resolver.State{
		Addresses: addresses,
	}
	err = g.cc.UpdateState(g.state)
	if err != nil {
		g.cc.ReportError(err)
	}
	return g.state
}

func (g *grpcResolver) ResolveNow(options resolver.ResolveNowOptions) {
	g.state = g.resolve()
}

func (g *grpcResolver) watch() error {
	eventCh, err := g.r.Subscribe(g.target.Endpoint())
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case e := <-eventCh:
				//switch e.Type {
				//case register.EventTypeAdd:
				//
				//	g.state.Addresses = append(g.state.Addresses, resolver.Address{
				//		Addr: e.Instance.Address,
				//	})
				//case register.EventTypeDelete:
				//	for i, addr := range g.state.Addresses {
				//		if addr.Addr == e.Instance.Address {
				//			g.state.Addresses = append(g.state.Addresses[:i], g.state.Addresses[i+1:]...)
				//			break
				//		}
				//	}
				//default:
				//	g.cc.ReportError(fmt.Errorf("unknown event type: %v", e.Type))
				//}

				log.Println("event:", e)
				g.resolve()

			case <-g.close:
				return
			}
		}
	}()
	return nil
}
func (g *grpcResolver) Close() {
	close(g.close)
}
