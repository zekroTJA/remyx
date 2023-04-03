package routers

import (
	"crypto/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/zekrotja/jwt"
	"github.com/zekrotja/remyx/internal/shared"
	"github.com/zekrotja/remyx/internal/webserver/models"
	"github.com/zekrotja/remyx/internal/webserver/tokens"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

type oauthState struct {
	jwt.PublicClaims
	Redirect string `json:"redirect,omitempty"`
}

type routerOAuth struct {
	auth  *spotifyauth.Authenticator
	cache tokens.Cache
	debug bool

	jwtHandler *jwt.Handler[oauthState]
}

func OAuth(rg *gin.RouterGroup, auth *spotifyauth.Authenticator, cache tokens.Cache, debug bool) {
	randKey := make([]byte, 512)
	_, err := rand.Read(randKey)
	if err != nil {
		panic(err)
	}

	jwtHandler := jwt.NewHandler[oauthState](jwt.NewHmacSha512(randKey))

	r := routerOAuth{
		auth:       auth,
		cache:      cache,
		debug:      debug,
		jwtHandler: &jwtHandler,
	}

	rg.GET("/login", r.login)
	rg.GET("/callback", r.callback)
}

func (t routerOAuth) login(ctx *gin.Context) {
	redirect := ctx.Query("redirect")

	var state oauthState
	state.Iss = "remyx"
	state.SetIat()
	state.SetNbfTime(time.Now())
	state.SetExpDuration(5 * time.Minute)
	state.Redirect = redirect

	stateStr, err := t.jwtHandler.EncodeAndSign(state)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Error{
			Message: "failed encoding and signing OAuth2 state", Details: err.Error(),
		})
		return
	}

	ctx.Redirect(http.StatusTemporaryRedirect, t.auth.AuthURL(stateStr))
}

func (t routerOAuth) callback(ctx *gin.Context) {
	stateStr := ctx.Query("state")
	state, err := t.jwtHandler.DecodeAndValidate(stateStr)
	if jwt.IsJWTError(err) {
		ctx.JSON(http.StatusBadRequest, models.Error{
			Message: "invalid state token", Details: err.Error(),
		})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Error{
			Message: "failed decode and validate state", Details: err.Error(),
		})
		return
	}

	// This is super hacky but required because the actual state check
	// occurs above by checking the signed JWT and will not be handled by
	// the Token method of the Spotify Authenticator.
	query := ctx.Request.URL.Query()
	query.Del("state")
	ctx.Request.URL.RawQuery = query.Encode()

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

	redirect := "/"
	if state.Redirect != "" {
		redirect += state.Redirect
	}

	ctx.Redirect(http.StatusTemporaryRedirect, redirect)
}
