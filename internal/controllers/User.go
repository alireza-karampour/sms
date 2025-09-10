package controllers

import (
	"errors"
	"net/http"
	"sync"

	"github.com/alireza-karampour/sms/pkg/middlewares"
	. "github.com/alireza-karampour/sms/pkg/utils"
	"github.com/alireza-karampour/sms/proto/api"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
)

type User struct {
	*Base
	db *pgx.Conn
}

func NewUser(parent *gin.RouterGroup, db *pgx.Conn) *User {
	base := NewBase("/user", parent, middlewares.WriteErrorBody)
	user := &User{
		base,
		db,
	}

	base.RegisterRoutes(func(gp *gin.RouterGroup) {
		gp.POST("", user.CreateNewUser)
		gp.PUT("/balance", user.AddBalance)
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
	rows, err := u.db.Query(ctx, "CreateUser", user.Username, 0)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		if ErrContains(err, "duplicate key value") {
			ctx.AbortWithError(http.StatusConflict, ErrUserAlreadyExists)
			return
		}
		ctx.AbortWithError(http.StatusInternalServerError, rows.Err())
		return
	}
	ctx.String(200, "OK")
	return
}

func (u *User) AddBalance(ctx *gin.Context) {
	err := sync.OnceValue(func() error {
		_, err := u.db.Prepare(ctx, "AddBalance", `UPDATE users SET balance = balance + $1 WHERE username = $2 RETURNING balance;`)
		if err != nil {
			return err
		}
		return nil
	})()
	user := &api.User{}
	err = ctx.BindJSON(user)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	rows, err := u.db.Query(ctx, "AddBalance", user.Balance, user.Username)
	defer rows.Close()
	newBalance := 0
	if rows.Next() {
		err := rows.Scan(&newBalance)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}
	ctx.JSON(200, map[string]any{
		"new_balance": newBalance,
	})
	return
}
