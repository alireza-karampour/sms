package workers

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	. "github.com/alireza-karampour/sms/internal/streams"
	. "github.com/alireza-karampour/sms/internal/subjects"
	"github.com/alireza-karampour/sms/pkg/nats"
	. "github.com/alireza-karampour/sms/pkg/utils"
	"github.com/alireza-karampour/sms/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	cost pgtype.Numeric
)

func init() {
	err := cost.Scan(viper.GetString("sms.cost"))
	if err != nil {
		panic(err)
	}
}

type Sms struct {
	*nats.Consumer
	*sqlc.Queries
	db *pgxpool.Pool
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
		Queries:  sqlc.New(pool),
		db:       pool,
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
	rate := sync.OnceValue(func() uint {
		return viper.GetUint("sms.normal.ratelimit")
	})()

	t := sync.OnceValue(func() *time.Timer {
		return time.NewTimer(time.Millisecond * time.Duration(rate))
	})()
	t.Reset(time.Millisecond * time.Duration(rate))

	var sub Subject = Subject(msg.Subject())
	switch {
	case sub.Filter(ANY, ANY, REQ):
		logrus.Debugf("Msg: %s\n", string(msg.Data()))
		sms := new(sqlc.Sm)
		err := json.Unmarshal(msg.Data(), sms)
		if err != nil {
			msg.TermWithReason(err.Error())
			return
		}

		tx, err := s.db.Begin(context.Background())
		if err != nil {
			logrus.Errorf("failed to begin tx: %s\n", err.Error())
			err := msg.NakWithDelay(time.Second)
			if err != nil {
				logrus.Errorf("failed to NAK: %s\n", err.Error())
			}
			return
		}
		defer tx.Rollback(context.Background())
		q := s.WithTx(tx)
		err = q.AddSms(context.Background(), sqlc.AddSmsParams{
			UserID:        sms.UserID,
			PhoneNumberID: sms.PhoneNumberID,
			ToPhoneNumber: sms.ToPhoneNumber,
			Status:        sms.Status,
			Message:       sms.Message,
		})
		if err != nil {
			logrus.Errorf("failed to add sms: %s\n", err.Error())
			err = msg.NakWithDelay(time.Second)
			if err != nil {
				logrus.Errorf("failed to NAK msg: %s\n", err.Error())
			}
			return
		}
		newBalance, err := q.SubBalance(context.Background(), sqlc.SubBalanceParams{
			Amount: cost,
			UserID: sms.UserID,
		})
		if err != nil {
			logrus.Errorf("failed to subtract balance: %s\n", err.Error())
			err = msg.NakWithDelay(time.Second)
			if err != nil {
				logrus.Errorf("failed to NAK msg: %s\n", err.Error())
			}
			return
		}
		num, err := newBalance.Float64Value()
		if err != nil {
			logrus.Error("failed to convert balance to float64")
		} else {
			logrus.Debugf("UserID: %d NewBalance: %f\n", sms.UserID, num.Float64)
		}

		err = msg.DoubleAck(context.Background())
		if err != nil {
			logrus.Errorf("failed to DoubleAck: %s", err.Error())
			return
		}
		tx.Commit(context.Background())
		<-t.C
	case sub.Filter(ANY, ANY, STAT):
		logrus.Debugf("NORMAL Subject: %s -- Msg: %s\n", msg.Subject(), string(msg.Data()))
		err := msg.DoubleAck(context.Background())
		if err != nil {
			logrus.Errorf("failed to DoubleAck: %s", err)
			return
		}
	}

}

func (s *Sms) handleExpressSms(msg jetstream.Msg) {
	rate := sync.OnceValue(func() uint {
		return viper.GetUint("sms.express.ratelimit")
	})()

	t := sync.OnceValue(func() *time.Timer {
		return time.NewTimer(time.Millisecond * time.Duration(rate))
	})()
	t.Reset(time.Millisecond * time.Duration(rate))

	var sub Subject = Subject(msg.Subject())
	switch {
	case sub.Filter(ANY, ANY, ANY, REQ):
		logrus.Debugf("EXPRESS Subject: %s -- Msg: %s\n", msg.Subject(), string(msg.Data()))
		sms := new(sqlc.Sm)
		err := json.Unmarshal(msg.Data(), sms)
		if err != nil {
			msg.TermWithReason(err.Error())
			return
		}

		tx, err := s.db.Begin(context.Background())
		if err != nil {
			logrus.Errorf("failed to begin tx: %s\n", err.Error())
			err := msg.NakWithDelay(time.Second)
			if err != nil {
				logrus.Errorf("failed to NAK: %s\n", err.Error())
			}
			return
		}
		defer tx.Rollback(context.Background())
		q := s.WithTx(tx)
		err = q.AddSms(context.Background(), sqlc.AddSmsParams{
			UserID:        sms.UserID,
			PhoneNumberID: sms.PhoneNumberID,
			ToPhoneNumber: sms.ToPhoneNumber,
			Status:        sms.Status,
			Message:       sms.Message,
		})
		if err != nil {
			logrus.Errorf("failed to add sms: %s\n", err.Error())
			err = msg.NakWithDelay(time.Second)
			if err != nil {
				logrus.Errorf("failed to NAK msg: %s\n", err.Error())
			}
			return
		}

		newBalance, err := q.SubBalance(context.Background(), sqlc.SubBalanceParams{
			Amount: cost,
			UserID: sms.UserID,
		})

		if err != nil {
			logrus.Errorf("failed to subtract balance: %s\n", err.Error())
			err = msg.NakWithDelay(time.Second)
			if err != nil {
				logrus.Errorf("failed to NAK msg: %s\n", err.Error())
			}
			return
		}
		num, err := newBalance.Float64Value()
		if err != nil {
			logrus.Error("failed to convert balance to float64")
		} else {
			logrus.Debugf("UserID: %d NewBalance: %f\n", sms.UserID, num.Float64)
		}

		err = msg.DoubleAck(context.Background())
		if err != nil {
			logrus.Errorf("failed to DoubleAck: %s", err.Error())
			return
		}
		tx.Commit(context.Background())
		<-t.C

	case sub.Filter(ANY, ANY, ANY, STAT):
		logrus.Debugf("EXPRESS Subject: %s -- Msg: %s\n", msg.Subject(), string(msg.Data()))
		err := msg.DoubleAck(context.Background())
		if err != nil {
			logrus.Errorf("failed to DoubleAck: %s", err)
			return
		}
	}
}

func (s *Sms) errHandler(ctx jetstream.ConsumeContext, err error) {
	logrus.Errorf("ConsumerError: %s\n", err)
}
