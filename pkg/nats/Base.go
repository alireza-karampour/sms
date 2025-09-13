package nats

import (
	"context"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type Base struct {
	*nats.Conn
	jetstream.JetStream
	Streams map[StreamName]jetstream.Stream
}

func NewBase(nc *nats.Conn) (*Base, error) {
	jsi, err := jetstream.New(nc)
	if err != nil {
		return nil, err
	}
	return &Base{
		Conn:      nc,
		JetStream: jsi,
		Streams:   make(map[StreamName]jetstream.Stream),
	}, nil
}

func (b *Base) BindStreams(ctx context.Context, streams ...jetstream.StreamConfig) error {
	for _, str := range streams {
		jss, err := b.CreateOrUpdateStream(ctx, str)
		if err != nil {
			return err
		}
		b.Streams[str.Name] = jss
	}
	return nil
}
