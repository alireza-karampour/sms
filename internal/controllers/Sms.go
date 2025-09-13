package controllers

import (
	"context"

	. "github.com/alireza-karampour/sms/internal/streams"
	. "github.com/alireza-karampour/sms/internal/subjects"
	"github.com/alireza-karampour/sms/pkg/middlewares"
	mynats "github.com/alireza-karampour/sms/pkg/nats"
	. "github.com/alireza-karampour/sms/pkg/utils"
	"github.com/alireza-karampour/sms/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/spf13/viper"
)

type Sms struct {
	*Base
	db   *pgxpool.Pool
	sp   *mynats.Publisher
	cost uint64
}

func NewSms(parent *gin.RouterGroup, db *pgxpool.Pool, nc *nats.Conn) (*Sms, error) {
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
			Name:        NORMAL_SMS_CONSUMER_NAME,
			Description: "work queue for handling sms with normal priority",
			Subjects: []string{
				MakeSubject(SMS, SEND, REQ),
				MakeSubject(SMS, SEND, STAT),
				MakeSubject(SMS, SEND, ERR),
			},
			Retention: jetstream.WorkQueuePolicy,
			Storage:   jetstream.FileStorage,
		},
		jetstream.StreamConfig{
			Name:        EXPRESS_SMS_CONSUMER_NAME,
			Description: "work queue for handling sms with high priority",
			Subjects: []string{
				MakeSubject(SMS, EX, SEND, REQ),
				MakeSubject(SMS, EX, SEND, STAT),
				MakeSubject(SMS, EX, SEND, ERR),
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
	query := new(struct {
		Express bool `json:"express"`
	})

	ctx.BindQuery(query)
	sms := new(sqlc.Sm)
	err := ctx.BindJSON(sms)
	if err != nil {
		ctx.AbortWithError(400, err)
		return
	}

}
