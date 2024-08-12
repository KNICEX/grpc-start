package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/KNICEX/grpc-start/register"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"log"
	"sync"
	"time"
)

var typesMap = map[mvccpb.Event_EventType]register.EventType{
	mvccpb.PUT:    register.EventTypeAdd,
	mvccpb.DELETE: register.EventTypeDelete,
}

type Registry struct {
	client *clientv3.Client
	sess   *concurrency.Session

	watchCancel []func()
	mu          sync.Mutex
}

func testConn(client *clientv3.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	_, err := client.Status(ctx, client.Endpoints()[0])
	return err

}

func NewRegistry(client *clientv3.Client) (*Registry, error) {
	if err := testConn(client); err != nil {
		return nil, err
	}
	sess, err := concurrency.NewSession(client)
	if err != nil {
		return nil, err
	}
	return &Registry{
		client: client,
		sess:   sess,
	}, nil
}

func (r *Registry) Register(ctx context.Context, ins register.ServiceInstance) error {
	instanceKey := fmt.Sprintf("/micro/%s/%s", ins.ServiceName, ins.Address)
	val, err := json.Marshal(ins)
	if err != nil {
		return err
	}
	// WithLease 会在注册的时候绑定一个租约，如果租约到期，注册的服务会被删除
	_, err = r.client.Put(ctx, instanceKey, string(val), clientv3.WithLease(r.sess.Lease()))
	if err != nil {
		return err
	}

	aliveCh, err := r.client.KeepAlive(context.Background(), r.sess.Lease())
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case _, ok := <-aliveCh:
				log.Println("keep alive:", ok)
				if !ok {
					log.Println("keep alive channel closed")
					return
				}
			}
		}
	}()
	return nil
}

func (r *Registry) Unregister(ctx context.Context, ins register.ServiceInstance) error {
	instanceKey := fmt.Sprintf("/micro/%s/%s", ins.ServiceName, ins.Address)
	_, err := r.client.Delete(ctx, instanceKey)
	return err
}

func (r *Registry) ListService(ctx context.Context, serviceName string) ([]register.ServiceInstance, error) {
	serviceKey := fmt.Sprintf("/micro/%s", serviceName)
	resp, err := r.client.Get(ctx, serviceKey, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	instances := make([]register.ServiceInstance, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var ins register.ServiceInstance
		err = json.Unmarshal(kv.Value, &ins)
		if err != nil {
			return nil, err
		}
		instances = append(instances, ins)
	}
	return instances, nil
}

func (r *Registry) Subscribe(serviceName string) (<-chan register.Event, error) {
	serviceKey := fmt.Sprintf("/micro/%s", serviceName)
	ctx, cancel := context.WithCancel(context.Background())

	r.mu.Lock()
	r.watchCancel = append(r.watchCancel, cancel)
	r.mu.Unlock()

	watchCh := r.client.Watch(ctx, serviceKey, clientv3.WithPrefix())
	res := make(chan register.Event)
	go func() {
		for {
			select {
			case resp := <-watchCh:
				if resp.Canceled {
					return
				}
				if resp.Err() != nil {
					log.Println("etcd registry watch error:", resp.Err())
					continue
				}
				for _, event := range resp.Events {
					var ins register.ServiceInstance
					err := json.Unmarshal(event.Kv.Value, &ins)
					if err != nil {
						continue
					}

					select {
					case res <- register.Event{
						Type:     typesMap[event.Type],
						Instance: ins,
					}:
					case <-ctx.Done():
						return
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return res, nil
}

func (r *Registry) Close() error {
	r.mu.Lock()
	for _, cancel := range r.watchCancel {
		cancel()
	}
	r.mu.Unlock()
	return nil
}
