package workers

import (
	"context"

	"github.com/alireza-karampour/sms/pkg/nats"
	"github.com/nats-io/nats.go/jetstream"
)

const (
	EXPRESS_SMS_CONSUMER_NAME string = "SmsExpress"
	NORMAL_SMS_CONSUMER_NAME  string = "Sms"
)

type Sms struct {
	*nats.Consumer
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
		Consumer: sc,
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
			Name:        NORMAL_SMS_CONSUMER_NAME,
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
				Name:        NORMAL_SMS_CONSUMER_NAME,
				Durable:     NORMAL_SMS_CONSUMER_NAME,
				Description: "consumes normal sms work queue",
			},
		},
	}
	expressSms := &nats.StreamConsumersConfig{
		Stream: jetstream.StreamConfig{
			Name:        EXPRESS_SMS_CONSUMER_NAME,
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
				Name:        EXPRESS_SMS_CONSUMER_NAME,
				Durable:     EXPRESS_SMS_CONSUMER_NAME,
				Description: "consumes high priority sms work queue",
			},
		},
	}
	return s.BindConsumers(ctx, normalSms, expressSms)
}

func (s *Sms) Start() error {
	// for strName, cons := range s.Consumers {
	//
	// }
	return nil
}
