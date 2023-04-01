package webserver

import (
	"github.com/gin-gonic/gin"
	"github.com/mandrigin/gin-spa/spa"
	"github.com/zekrotja/remyx/internal/config"
	"github.com/zekrotja/remyx/internal/database"
	"github.com/zekrotja/remyx/internal/myxer"
	"github.com/zekrotja/remyx/internal/webserver/middleware"
	"github.com/zekrotja/remyx/internal/webserver/routers"
	"github.com/zekrotja/remyx/internal/webserver/tokens"
	"github.com/zekrotja/rogu/level"

	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

func Run(
	cfg config.Config,
	auth *spotifyauth.Authenticator,
	db database.Database,
	cache tokens.Cache,
	mxr *myxer.Myxer,
) (err error) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	if cfg.Debug {
		router.Use(middleware.Cors("http://localhost:3000"))
	}

	api := router.Group("/api")
	api.Use(middleware.Logger(level.Info, "WebServer"))

	routers.OAuth(api.Group("/oauth"), auth, cache, cfg.Debug)

	// --- autenticated routes ---
	authApi := api.Group("/")
	authApi.Use(middleware.Auth(auth, cache))
	routers.Playlists(authApi.Group("/playlists"))
	routers.Remyxes(authApi.Group("/remyxes"), db, mxr)

	router.Use(spa.Middleware("/", "./web/dist"))

	return router.Run(cfg.Webserver.BindAddress)
}
