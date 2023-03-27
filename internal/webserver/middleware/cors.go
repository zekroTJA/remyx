package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Cors(originUrl string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("Access-Control-Allow-Origin", originUrl)
		ctx.Header("Access-Control-Allow-Methods", "GET,POST,DELETE")
		ctx.Header("Access-Control-Allow-Headers", "Content-Type,Cookie")
		ctx.Header("Access-Control-Allow-Credentials", "true")

		if ctx.Request.Method == http.MethodOptions {
			ctx.Status(http.StatusOK)
			ctx.Abort()
		}
	}
}
