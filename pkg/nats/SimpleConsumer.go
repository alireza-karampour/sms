package nats

import (
	"context"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type StreamConsumersConfig struct {
	Stream    jetstream.StreamConfig
	Consumers []jetstream.ConsumerConfig
}

type StreamConsumers struct {
	Stream    jetstream.Stream
	Consumers []jetstream.Consumer
}

func (s *StreamConsumers) AddConsumer(consumer jetstream.Consumer) {
	sync.OnceFunc(func() {
		s.Consumers = make([]jetstream.Consumer, 0, 1)
	})()
	s.Consumers = append(s.Consumers, consumer)
}

type SimpleConsumer struct {
	*Base
	Consumers map[string]*StreamConsumers
}

func NewSimpleConsumer(nc *nats.Conn) (*SimpleConsumer, error) {
	b, err := NewBase(nc)
	if err != nil {
		return nil, err
	}

	sc := &SimpleConsumer{
		Base:      b,
		Consumers: make(map[string]*StreamConsumers),
	}
	return sc, nil
}

func (sc *SimpleConsumer) BindConsumers(ctx context.Context, streams ...StreamConsumersConfig) error {
	for _, conf := range streams {
		strName := conf.Stream.Name
		err := sc.BindStreams(ctx, conf.Stream)
		if err != nil {
			return err
		}

		for _, consumerConf := range conf.Consumers {
			cons, err := sc.CreateOrUpdateConsumer(ctx, strName, consumerConf)
			if err != nil {
				return err
			}
			consumers, ok := sc.Consumers[strName]
			if !ok {
				sc.Consumers[strName] = new(StreamConsumers)
				consumers = sc.Consumers[strName]
			}
			consumers.Stream = sc.Streams[strName]
			consumers.AddConsumer(cons)
		}
	}
	return nil
}
