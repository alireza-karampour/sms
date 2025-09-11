package workers

import (
	"context"

	"github.com/alireza-karampour/sms/pkg/nats"
	"github.com/nats-io/nats.go/jetstream"
)

type Sms struct {
	*nats.SimpleConsumer
}

func NewSms(ctx context.Context, natsAddress string) (*Sms, error) {
	nc, err := nats.Connect(natsAddress)
	if err != nil {
		return nil, err
	}

	sc, err := nats.NewSimpleConsumer(nc)
	if err != nil {
		return nil, err
	}

	worker := &Sms{
		SimpleConsumer: sc,
	}

	err = worker.bindConsumer(ctx)
	if err != nil {
		return nil, err
	}

	return worker, nil
}

func (s *Sms) bindConsumer(ctx context.Context) error {
	normalSms := &nats.StreamConsumersConfig{
		Stream: jetstream.StreamConfig{
			Name:        "Sms",
			Description: "work queue for handling sms with normal priority",
			Subjects: []string{
				"sms.send.request",
				"sms.send.status",
				"sms.send.error",
			},
			Retention: jetstream.WorkQueuePolicy,
			Storage:   jetstream.FileStorage,
		},
		Consumers: []jetstream.ConsumerConfig{
			{
				Name:        "SmsConsumer",
				Durable:     "SmsConsumer",
				Description: "consumes normal sms work queue",
			},
		},
	}
	expressSms := &nats.StreamConsumersConfig{
		Stream: jetstream.StreamConfig{
			Name:        "SmsExpress",
			Description: "work queue for handling sms with high priority",
			Subjects: []string{
				"sms.ex.send.request",
				"sms.ex.send.status",
				"sms.ex.send.error",
			},
			Retention: jetstream.WorkQueuePolicy,
			Storage:   jetstream.FileStorage,
		},
		Consumers: []jetstream.ConsumerConfig{
			{
				Name:        "SmsExpressConsumer",
				Durable:     "SmsExpressConsumer",
				Description: "consumes high priority sms work queue",
			},
		},
	}
	return s.BindConsumers(ctx, normalSms, expressSms)
}
