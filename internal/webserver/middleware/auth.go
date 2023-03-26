package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zekrotja/remyx/internal/shared"
	"github.com/zekrotja/remyx/internal/webserver/tokens"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

func Auth(auth *spotifyauth.Authenticator, cache tokens.Cache) gin.HandlerFunc {
	var getToken = func(ctx *gin.Context) *oauth2.Token {
		uid, _ := ctx.Cookie(shared.AuthTokenCookie)
		if uid == "" {
			return nil
		}

		token, ok := cache.Get(uid)
		if !ok {
			return nil
		}

		return token
	}

	return func(ctx *gin.Context) {
		token := getToken(ctx)
		if token == nil {
			ctx.Status(http.StatusUnauthorized)
			ctx.Abort()
			return
		}

		ctx.Set("token", token)

		client := auth.Client(ctx.Request.Context(), token)
		ctx.Set("client", client)
	}
}
