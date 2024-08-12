package register

import "context"

type Register interface {
	Register(ctx context.Context, ins ServiceInstance) error
	Unregister(ctx context.Context, ins ServiceInstance) error

	ListService(ctx context.Context, serviceName string) ([]ServiceInstance, error)

	Subscribe(serviceName string) (<-chan Event, error)
	Close() error
}

type ServiceInstance struct {
	ServiceName string
	Address     string
}

type EventType int

const (
	EventUnknown EventType = iota
	EventTypeAdd
	EventTypeDelete
)

type Event struct {
	Type     EventType
	Instance ServiceInstance
}
