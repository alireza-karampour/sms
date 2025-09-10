package controllers

import (
	"net/http"
	"sync"

	"github.com/alireza-karampour/sms/proto/api"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type User struct {
	*Base
	db *pgx.Conn
}

func NewUser(parent *gin.RouterGroup, db *pgx.Conn) *User {
	base := NewBase("/user", parent)

	user := &User{
		base,
		db,
	}

	base.RegisterRoutes(func(gp *gin.RouterGroup) {
		gp.POST("/", user.CreateNewUser)
	})

	return user
}

func (u *User) CreateNewUser(ctx *gin.Context) {
	err := sync.OnceValue(func() error {
		_, err := u.db.Prepare(ctx, "CreateUser", `INSERT INTO users (username, balance) VALUES ($1, $2);`)
		if err != nil {
			return err
		}
		return nil
	})()

	if err != nil {
		ctx.AbortWithError(500, err)
		return
	}
	user := &api.User{}
	err = ctx.BindJSON(user)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	_, err = u.db.Query(ctx, "CreateUser", user.Username, 0)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	ctx.String(200, "OK")
	return
}
