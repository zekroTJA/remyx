package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/zekrotja/remyx/internal/shared"
	"github.com/zekrotja/remyx/internal/webserver/models"
	"github.com/zekrotja/remyx/internal/webserver/tokens"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

type routerOAuth struct {
	auth  *spotifyauth.Authenticator
	cache tokens.Cache
	debug bool
}

func OAuth(rg *gin.RouterGroup, auth *spotifyauth.Authenticator, cache tokens.Cache, debug bool) {
	r := routerOAuth{
		auth:  auth,
		cache: cache,
		debug: debug,
	}

	rg.GET("/login", r.login)
	rg.GET("/callback", r.callback)
}

func (t routerOAuth) login(ctx *gin.Context) {
	ctx.Redirect(http.StatusTemporaryRedirect, t.auth.AuthURL(""))
}

func (t routerOAuth) callback(ctx *gin.Context) {
	token, err := t.auth.Token(ctx.Request.Context(), "", ctx.Request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.Error{
			Message: "invalid code", Details: err.Error(),
		})
		return
	}

	client := t.auth.Client(ctx.Request.Context(), token)
	_, err = spotify.New(client).CurrentUser(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.Error{
			Message: "invalid token", Details: err.Error(),
		})
		return
	}

	uid := uuid.New().String()
	t.cache.Set(uid, token, shared.SessionLifetime)

	ctx.SetCookie(
		shared.AuthTokenCookie,
		uid,
		int(shared.SessionLifetime.Seconds()),
		"/",
		"",
		!t.debug,
		true,
	)

	ctx.Redirect(http.StatusTemporaryRedirect, "/")
}
