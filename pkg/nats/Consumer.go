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

type Consumer struct {
	*Base
	Consumers map[string]*StreamConsumers
	ctxs      []jetstream.ConsumeContext
}

func NewConsumer(nc *nats.Conn) (*Consumer, error) {
	b, err := NewBase(nc)
	if err != nil {
		return nil, err
	}

	c := &Consumer{
		Base:      b,
		Consumers: make(map[string]*StreamConsumers),
		ctxs:      make([]jetstream.ConsumeContext, 0, 1),
	}
	return c, nil
}

func (c *Consumer) BindConsumers(ctx context.Context, streams ...*StreamConsumersConfig) error {
	for _, conf := range streams {
		strName := conf.Stream.Name
		err := c.BindStreams(ctx, conf.Stream)
		if err != nil {
			return err
		}

		for _, consumerConf := range conf.Consumers {
			cons, err := c.CreateOrUpdateConsumer(ctx, strName, consumerConf)
			if err != nil {
				return err
			}
			consumers, ok := c.Consumers[strName]
			if !ok {
				c.Consumers[strName] = new(StreamConsumers)
				consumers = c.Consumers[strName]
			}
			consumers.Stream = c.Streams[strName]
			consumers.AddConsumer(cons)
		}
	}
	return nil
}

func (c *Consumer) StartConsumers(ctx context.Context, consumeHandler func(msg jetstream.Msg), opts ...jetstream.PullConsumeOpt) error {
	for _, consumers := range c.Consumers {
		for _, consumer := range consumers.Consumers {
			ctx, err := consumer.Consume(consumeHandler, opts...)
			if err != nil {
				return err
			}
			c.ctxs = append(c.ctxs, ctx)
		}
	}
	return nil
}
