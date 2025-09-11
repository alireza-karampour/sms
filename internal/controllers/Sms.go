package controllers

import (
	"context"

	"github.com/alireza-karampour/sms/pkg/middlewares"
	mynats "github.com/alireza-karampour/sms/pkg/nats"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/spf13/viper"
)

type Sms struct {
	*Base
	db   *pgx.Conn
	sp   *mynats.Publisher
	cost uint64
}

func NewSms(parent *gin.RouterGroup, db *pgx.Conn, nc *nats.Conn) (*Sms, error) {
	base := NewBase("/sms", parent, middlewares.WriteErrorBody)
	sp, err := mynats.NewSimplePublisher(nc)
	if err != nil {
		return nil, err
	}

	sms := &Sms{
		Base: base,
		db:   db,
		sp:   sp,
		cost: viper.GetUint64("api.sms.cost"),
	}

	err = sp.BindStreams(context.Background(),
		jetstream.StreamConfig{
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
		jetstream.StreamConfig{
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
	)
	if err != nil {
		return nil, err
	}

	base.RegisterRoutes(func(gp *gin.RouterGroup) {
		gp.POST("", sms.SendSms)
	})

	return sms, nil
}

func (s *Sms) SendSms(ctx *gin.Context) {
}
