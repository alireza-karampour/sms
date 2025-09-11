package controllers

import (
	"errors"
	"github.com/alireza-karampour/sms/pkg/middlewares"
	. "github.com/alireza-karampour/sms/pkg/utils"
	"github.com/alireza-karampour/sms/proto/api"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
)

var (
	ErrPhoneNumberAlreadyExists = errors.New("phone number already exists")
	ErrPhoneNumberNotFound      = errors.New("phone number not found")
)

type PhoneNumber struct {
	*Base
	db *pgxpool.Pool
}

func NewPhoneNumber(parent *gin.RouterGroup, db *pgxpool.Pool) *PhoneNumber {
	base := NewBase("/phone-number", parent, middlewares.WriteErrorBody)
	pn := &PhoneNumber{
		base,
		db,
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
	request := &api.AddPhoneNumberRequest{}
	err := ctx.BindJSON(request)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	rows, err := pn.db.Query(ctx, `INSERT INTO phone_numbers (user_id, phone_number) VALUES ((SELECT id FROM users WHERE username = $1), $2);`, request.Username, request.PhoneNumber)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		if ErrContains(err, "duplicate key value") {
			ctx.AbortWithError(http.StatusConflict, ErrPhoneNumberAlreadyExists)
			return
		}
		if ErrContains(err, "violates foreign key constraint") {
			ctx.AbortWithError(http.StatusNotFound, errors.New("user not found"))
			return
		}
		ctx.AbortWithError(http.StatusInternalServerError, rows.Err())
		return
	}
	ctx.JSON(200, gin.H{
		"status": 200,
		"msg":    "OK",
	})
}

func (pn *PhoneNumber) GetPhoneNumber(ctx *gin.Context) {
	id := ctx.Param("id")
	rows, err := pn.db.Query(ctx, `SELECT id, user_id, phone_number FROM phone_numbers WHERE id = $1;`, id)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var phoneNumber api.PhoneNumber
	if rows.Next() {
		err = rows.Scan(&phoneNumber.Id, &phoneNumber.UserId, &phoneNumber.Number)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	} else {
		ctx.AbortWithError(http.StatusNotFound, ErrPhoneNumberNotFound)
		return
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(200, &phoneNumber)
}

func (pn *PhoneNumber) DeletePhoneNumber(ctx *gin.Context) {
	id := ctx.Param("id")
	rows, err := pn.db.Query(ctx, `DELETE FROM phone_numbers WHERE id = $1 RETURNING id;`, id)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if !rows.Next() {
		ctx.AbortWithError(http.StatusNotFound, ErrPhoneNumberNotFound)
		return
	}
	rows.Close()
	if err = rows.Err(); err != nil {
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
	rows, err := pn.db.Query(ctx, `SELECT pn.id, pn.user_id, pn.phone_number FROM phone_numbers pn JOIN users u ON pn.user_id = u.id WHERE u.username = $1;`, username)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	var phoneNumbers []*api.PhoneNumber
	for rows.Next() {
		var phoneNumber api.PhoneNumber
		err = rows.Scan(&phoneNumber.Id, &phoneNumber.UserId, &phoneNumber.Number)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		phoneNumbers = append(phoneNumbers, &phoneNumber)
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(200, phoneNumbers)
}
