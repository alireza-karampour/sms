package nats

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type StreamName = string
type Subject = string

type SimplePublisher struct {
	*Base
}

func NewSimplePublisher(nc *nats.Conn) (*SimplePublisher, error) {
	b, err := NewBase(nc)
	if err != nil {
		return nil, err
	}

	return &SimplePublisher{
		Base: b,
	}, nil
}

func (sp *SimplePublisher) BindStreams(ctx context.Context, streams ...jetstream.StreamConfig) error {
	for _, str := range streams {
		jss, err := sp.CreateOrUpdateStream(ctx, str)
		if err != nil {
			return err
		}
		sp.Streams[str.Name] = jss
	}
	return nil
}
