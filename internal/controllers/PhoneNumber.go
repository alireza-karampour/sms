package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/alireza-karampour/sms/pkg/middlewares"
	. "github.com/alireza-karampour/sms/pkg/utils"
	"github.com/alireza-karampour/sms/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrPhoneNumberAlreadyExists = errors.New("phone number already exists")
	ErrPhoneNumberNotFound      = errors.New("phone number not found")
)

type PhoneNumber struct {
	*Base
	db *sqlc.Queries
}

func NewPhoneNumber(parent *gin.RouterGroup, db *pgxpool.Pool) *PhoneNumber {
	base := NewBase("/phone-number", parent, middlewares.WriteErrorBody)
	pn := &PhoneNumber{
		base,
		sqlc.New(db),
	}

	base.RegisterRoutes(func(gp *gin.RouterGroup) {
		gp.POST("", pn.CreatePhoneNumber)
		gp.GET("/:id", pn.GetPhoneNumber)
		gp.DELETE("/:id", pn.DeletePhoneNumber)
		gp.GET("/user/:username", pn.GetPhoneNumbersByUser)
	})

	return pn
}

func (pn *PhoneNumber) CreatePhoneNumber(ctx *gin.Context) {
	request := new(sqlc.PhoneNumber)
	err := ctx.BindJSON(request)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	err = pn.db.AddPhoneNumber(ctx, sqlc.AddPhoneNumberParams{
		UserID:      request.UserID,
		PhoneNumber: request.PhoneNumber,
	})
	if err != nil {
		if ErrContains(err, "duplicate key value") {
			ctx.AbortWithError(http.StatusConflict, ErrPhoneNumberAlreadyExists)
			return
		}
		if ErrContains(err, "violates foreign key constraint") {
			ctx.AbortWithError(http.StatusNotFound, errors.New("user not found"))
			return
		}
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(200, gin.H{
		"status": 200,
		"msg":    "OK",
	})
}

func (pn *PhoneNumber) GetPhoneNumber(ctx *gin.Context) {
	id := ctx.Param("id")
	idInt, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	phoneNumber, err := pn.db.GetPhoneNumber(ctx, int32(idInt))
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(200, phoneNumber)
}

func (pn *PhoneNumber) DeletePhoneNumber(ctx *gin.Context) {
	id := ctx.Param("id")
	idInt, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("invalid id"))
		return
	}

	_, err = pn.db.DeletePhoneNumber(ctx, int32(idInt))
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(200, gin.H{
		"status": 200,
		"msg":    "OK",
	})
}

func (pn *PhoneNumber) GetPhoneNumbersByUser(ctx *gin.Context) {
	username := ctx.Param("username")
	phoneNumbers, err := pn.db.GetPhoneNumbersByUsername(ctx, username)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(200, phoneNumbers)
}
