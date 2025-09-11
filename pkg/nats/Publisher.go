package nats

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type StreamName = string
type Subject = string

type Publisher struct {
	*Base
}

func NewSimplePublisher(nc *nats.Conn) (*Publisher, error) {
	b, err := NewBase(nc)
	if err != nil {
		return nil, err
	}

	return &Publisher{
		Base: b,
	}, nil
}

func (sp *Publisher) BindStreams(ctx context.Context, streams ...jetstream.StreamConfig) error {
	for _, str := range streams {
		jss, err := sp.CreateOrUpdateStream(ctx, str)
		if err != nil {
			return err
		}
		sp.Streams[str.Name] = jss
	}
	return nil
}
