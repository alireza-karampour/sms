package workers

import (
	"context"

	. "github.com/alireza-karampour/sms/internal/streams"
	. "github.com/alireza-karampour/sms/internal/subjects"
	"github.com/alireza-karampour/sms/pkg/nats"
	. "github.com/alireza-karampour/sms/pkg/utils"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/sirupsen/logrus"
)

type Sms struct {
	*nats.Consumer
	*pgxpool.Pool
}

func NewSms(ctx context.Context, natsAddress string, pool *pgxpool.Pool) (*Sms, error) {
	nc, err := nats.Connect(natsAddress)
	if err != nil {
		return nil, err
	}

	sc, err := nats.NewConsumer(nc)
	if err != nil {
		return nil, err
	}

	worker := &Sms{
		Consumer: sc,
		Pool:     pool,
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
				MakeSubject(SMS, SEND, REQ),
				MakeSubject(SMS, SEND, STAT),
				MakeSubject(SMS, SEND, ERR),
			},
			Retention:   jetstream.WorkQueuePolicy,
			Storage:     jetstream.FileStorage,
			AllowDirect: true,
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
				MakeSubject(SMS, EX, SEND, REQ),
				MakeSubject(SMS, EX, SEND, STAT),
				MakeSubject(SMS, EX, SEND, ERR),
			},
			Retention:   jetstream.WorkQueuePolicy,
			Storage:     jetstream.FileStorage,
			AllowDirect: true,
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

func (s *Sms) Start(ctx context.Context) error {
	var errHandlerOpt jetstream.ConsumeErrHandler = s.errHandler
	opts := []jetstream.PullConsumeOpt{
		errHandlerOpt,
	}
	err := s.StartConsumers(ctx, s.handler, opts...)
	if err != nil {
		return err
	}
	return nil
}

func (s *Sms) handler(msg jetstream.Msg) {
	sub := Subject(msg.Subject())
	switch {
	case sub.Filter(SMS, SEND, ANY):
		s.handleNormalSms(msg)
	case sub.Filter(SMS, EX, ANY, ANY):
		s.handleExpressSms(msg)
	}
}

func (s *Sms) handleNormalSms(msg jetstream.Msg) {
	var sub Subject = Subject(msg.Subject())
	switch {
	case sub.Filter(ANY, ANY, REQ):
		logrus.Debugf("Msg: %s\n", string(msg.Data()))
	case sub.Filter(ANY, ANY, STAT):
		logrus.Debugf("NORMAL Subject: %s -- Msg: %s\n", msg.Subject(), string(msg.Data()))
	}

	err := msg.DoubleAck(context.Background())
	if err != nil {
		logrus.Errorf("failed to DoubleAck: %s", err)
		return
	}

}

func (s *Sms) handleExpressSms(msg jetstream.Msg) {
	defer msg.DoubleAck(context.Background())

	var sub Subject = Subject(msg.Subject())
	switch {
	case sub.Filter(ANY, ANY, ANY, REQ):
		logrus.Debugf("EXPRESS Subject: %s -- Msg: %s\n", msg.Subject(), string(msg.Data()))
	case sub.Filter(ANY, ANY, ANY, STAT):
		logrus.Debugf("EXPRESS Subject: %s -- Msg: %s\n", msg.Subject(), string(msg.Data()))
	}
}

func (s *Sms) errHandler(ctx jetstream.ConsumeContext, err error) {
	logrus.Errorf("ConsumerError: %s\n", err)
}
