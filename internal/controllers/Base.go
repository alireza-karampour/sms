package controllers

import "github.com/gin-gonic/gin"

type Base struct {
	Prefix string
	gp     *gin.RouterGroup
}

func NewBase(self string, parent *gin.RouterGroup, middlewares ...gin.HandlerFunc) *Base {
	gp := parent.Group(self, middlewares...)
	return &Base{
		Prefix: self,
		gp:     gp,
	}
}

func (b *Base) RegisterRoutes(fn func(gp *gin.RouterGroup)) {
	fn(b.gp)
}
