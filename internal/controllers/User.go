package controllers

import "github.com/gin-gonic/gin"

type User struct {
	*Base
}

func NewUser(parent *gin.RouterGroup) *User {
	base := NewBase("/user", parent)

	user := &User{
		base,
	}

	base.RegisterRoutes(func(gp *gin.RouterGroup) {
		gp.POST("/", user.CreateNewUser)
	})

	return user
}

func (u *User) CreateNewUser(ctx *gin.Context) {
	ctx.JSON(200, struct{ Msg string }{"Hello World"})
	return
}
