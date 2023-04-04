package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zekroTJA/ratelimit"
	"github.com/zekroTJA/timedmap"
	"github.com/zekrotja/remyx/internal/shared"
	"github.com/zekrotja/remyx/internal/webserver/models"
)

var cleanupTicker = time.NewTicker(10 * time.Minute)

func Ratelimit(burst int, regen time.Duration) gin.HandlerFunc {
	limiters := timedmap.New(0, cleanupTicker.C)

	return func(ctx *gin.Context) {
		uid, _ := ctx.Cookie(shared.AuthTokenCookie)
		if uid == "" {
			ctx.JSON(http.StatusBadRequest,
				models.Error{Message: "request does not contain any session ID"})
			ctx.Abort()
			return
		}

		limiter, ok := limiters.GetValue(uid).(*ratelimit.Limiter)
		if !ok {
			limiter = ratelimit.NewLimiter(regen, burst)
			limiters.Set(uid, limiter, time.Duration(burst)*regen)
		}

		ok, res := limiter.Reserve()

		ctx.Header("X-Ratelimit-Limit", strconv.Itoa(burst))
		ctx.Header("X-Ratelimit-Remaining", strconv.Itoa(res.Remaining))
		ctx.Header("X-Ratelimit-Reset", strconv.Itoa(int(res.Reset.Unix())))

		if !ok {
			ctx.JSON(http.StatusTooManyRequests,
				models.Error{Message: "you have been rate limited"})
			ctx.Abort()
			return
		}
	}
}
