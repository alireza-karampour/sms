package controllers

import (
	"errors"
	"net/http"

	"github.com/alireza-karampour/sms/pkg/middlewares"
	"github.com/alireza-karampour/sms/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
)

type User struct {
	*Base
	db *sqlc.Queries
}

func NewUser(parent *gin.RouterGroup, db *pgxpool.Pool) *User {
	base := NewBase("/user", parent, middlewares.WriteErrorBody)
	user := &User{
		base,
		sqlc.New(db),
	}

	base.RegisterRoutes(func(gp *gin.RouterGroup) {
		gp.POST("", user.CreateNewUser)
		gp.PUT("/balance", user.AddBalance)
	})

	return user
}

func (u *User) CreateNewUser(ctx *gin.Context) {
	user := new(sqlc.User)
	err := ctx.BindJSON(user)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	err = u.db.AddUser(ctx, sqlc.AddUserParams{
		Username: user.Username,
		Balance:  user.Balance,
	})
	if err != nil {
		ctx.AbortWithError(500, err)
		return
	}

	ctx.String(200, "OK")
	return
}

func (u *User) AddBalance(ctx *gin.Context) {
	user := new(sqlc.User)
	err := ctx.BindJSON(user)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	newBalance, err := u.db.AddBalance(ctx, sqlc.AddBalanceParams{
		Balance:  user.Balance,
		Username: user.Username,
	})
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(200, map[string]any{
		"status":      200,
		"new_balance": newBalance,
	})
	return
}
