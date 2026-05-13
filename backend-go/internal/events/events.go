package events

import "context"

type Event struct {
	Type     string
	EntityID string
	Payload  map[string]any
}

type Publisher interface {
	Publish(context.Context, Event) error
}

type NoopPublisher struct{}

func (NoopPublisher) Publish(context.Context, Event) error {
	return nil
}
