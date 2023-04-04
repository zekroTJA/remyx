package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/davecgh/go-spew/spew"
	"github.com/zekrotja/remyx/internal/config"
	"github.com/zekrotja/remyx/internal/database"
	"github.com/zekrotja/remyx/internal/myxer"
	"github.com/zekrotja/remyx/internal/scheduler"
	"github.com/zekrotja/remyx/internal/webserver"
	"github.com/zekrotja/remyx/internal/webserver/tokens"
	"github.com/zekrotja/rogu/log"

	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

var scopes = []string{
	spotifyauth.ScopePlaylistReadPrivate,
	spotifyauth.ScopePlaylistReadCollaborative,
	spotifyauth.ScopePlaylistModifyPrivate,
	spotifyauth.ScopeUserLibraryRead,
}

func main() {
	cfg, err := config.Parse()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed initializing config")
	}

	log.SetLevel(cfg.LogLevel)
	log.Debug().Msgf("Parsed Config:\n%s", spew.Sdump(cfg))

	db, err := database.NewPostgresDriver(cfg.Database.URL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed setting up database connection")
	}
	defer db.Close()

	cache, err := tokens.NewCache()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed setting up token cache")
	}
	defer func() {
		err = cache.Close()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed closing token cache")
		}
	}()

	// TODO: Wrap authorizer stuff
	redirectUrl := fmt.Sprintf("%s/api/oauth/callback", cfg.Oauth.PublicAddress)
	auth := spotifyauth.New(
		spotifyauth.WithRedirectURL(redirectUrl),
		spotifyauth.WithClientID(cfg.Oauth.ClientID),
		spotifyauth.WithClientSecret(cfg.Oauth.ClientSecret),
		spotifyauth.WithScopes(scopes...),
	)

	mxr := myxer.New(db, auth, cfg.Oauth.PublicAddress)

	err = scheduler.Run(mxr, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed registering schedules")
	}

	log.Info().Field("addr", cfg.Webserver.BindAddress).Msg("Starting web server ...")
	go func() {
		err = webserver.Run(cfg, auth, db, cache, mxr)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed starting web server")
		}
	}()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	log.Info().Msg("Shutting down ...")
}
