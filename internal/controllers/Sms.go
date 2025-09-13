package controllers

import (
	"context"
	"encoding/json"
	"errors"

	. "github.com/alireza-karampour/sms/internal/streams"
	. "github.com/alireza-karampour/sms/internal/subjects"
	"github.com/alireza-karampour/sms/pkg/middlewares"
	mynats "github.com/alireza-karampour/sms/pkg/nats"
	. "github.com/alireza-karampour/sms/pkg/utils"
	"github.com/alireza-karampour/sms/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
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
	*Base
	db *pgxpool.Pool
	sp *mynats.Publisher
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
	var subject string
	if query.Express {
		subject = MakeSubject(SMS, EX, SEND, REQ)
	} else {
		subject = MakeSubject(SMS, SEND, REQ)
	}

	ctx.BindQuery(query)
	sms := new(sqlc.Sm)
	err := ctx.BindJSON(sms)
	if err != nil {
		ctx.AbortWithError(400, err)
		return
	}
	q := sqlc.New(s.db)
	balance, err := q.GetBalance(ctx, sms.UserID)
	if err != nil {
		ctx.AbortWithError(500, err)
		return
	}
	if balance.Int.Cmp(cost.Int) < 0 {
		ctx.AbortWithError(403, errors.New("not enough balance"))
		return
	}
	smsJson, err := json.Marshal(sms)
	if err != nil {
		ctx.AbortWithError(500, err)
		return
	}

	_, err = s.sp.JetStream.Publish(ctx, subject, smsJson)
	if err != nil {
		ctx.AbortWithError(500, err)
		return
	}
	ctx.JSON(200, gin.H{
		"msg": "OK",
	})
}
