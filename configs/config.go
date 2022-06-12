package configs

import (
	"context"

	"github.com/spf13/viper"
)

type HTTPServer interface {
	ListenAndServe() error
	Shutdown(ctx context.Context) error
}

type Config struct {
	Proxy

	LogLevel string
}

func SetDefaults() {
	viper.SetDefault("LogLevel", "info")
	if err := viper.BindEnv("LogLevel"); err != nil {
		panic(err)
	}

	viper.SetDefault("Proxy.TargetURL", "http://localhost:8001")
	if err := viper.BindEnv("Proxy.TargetURL"); err != nil {
		panic(err)
	}

	viper.SetDefault("Proxy.ListenAddr", ":8003")
	if err := viper.BindEnv("Proxy.ListenAddr"); err != nil {
		panic(err)
	}
}
