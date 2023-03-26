package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/zekrotja/rogu/level"
)

type Config struct {
	LogLevel level.Level `default:"5"`
	Debug    bool        `default:"false"`

	Database  DatabaseConfig
	Webserver WebserverConfig
	Oauth     OauthConfig
}

type DatabaseConfig struct {
	URL string `required:"true"`
}

type WebserverConfig struct {
	BindAddress string `default:"0.0.0.0:80"`
}

type OauthConfig struct {
	ClientID      string `required:"true"`
	ClientSecret  string `required:"true"`
	PublicAddress string `required:"true"`
}

func Parse() (Config, error) {
	godotenv.Load()

	var cfg Config
	err := envconfig.Process("remyx", &cfg)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}
