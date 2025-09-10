package middlewares

import "github.com/gin-gonic/gin"

func WriteErrorBody(ctx *gin.Context) {
	ctx.Next()
	if len(ctx.Errors) > 0 {
		res := gin.H{
			"status": ctx.Writer.Status(),
			"errors": make([]string, 0, len(ctx.Errors)),
		}
		for _, v := range ctx.Errors {
			res["errors"] = append(res["errors"].([]string), v.Error())
		}
		ctx.JSON(ctx.Writer.Status(), res)
	}
}
