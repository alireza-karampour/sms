package nats

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type StreamName = string
type Subject = string

type SimplePublisher struct {
	*nats.Conn
	jetstream.JetStream
	Streams map[StreamName]jetstream.Stream
}

func Connect(addr string) (*SimplePublisher, error) {
	nc, err := nats.Connect(fmt.Sprintf("nats://%s", addr))
	if err != nil {
		return nil, err
	}
	jsi, err := jetstream.New(nc)
	if err != nil {
		return nil, err
	}

	return &SimplePublisher{
		Conn:      nc,
		JetStream: jsi,
		Streams:   make(map[StreamName]jetstream.Stream),
	}, nil
}

func (sp *SimplePublisher) BindStreams(ctx context.Context, streams []jetstream.StreamConfig) error {
	for _, str := range streams {
		jss, err := sp.CreateOrUpdateStream(ctx, str)
		if err != nil {
			return err
		}
		sp.Streams[str.Name] = jss
	}
	return nil
}
