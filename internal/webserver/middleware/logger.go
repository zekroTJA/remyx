package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zekrotja/rogu/level"
	"github.com/zekrotja/rogu/log"
)

func Logger(lvl level.Level, tag string) gin.HandlerFunc {
	l := log.Tagged(tag)

	return func(ctx *gin.Context) {
		method := ctx.Request.Method
		path := ctx.Request.URL.Path
		now := time.Now()

		ctx.Next()

		took := time.Since(now)
		l.WithLevel(lvl).Fields(
			"status", ctx.Writer.Status(),
			"took", took,
		).Msgf("%s %s", method, path)
	}
}
