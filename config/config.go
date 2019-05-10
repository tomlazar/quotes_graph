package config

import (
	"go.uber.org/zap"

	"github.com/getsentry/raven-go"
	"github.com/spf13/viper"
)

// Logger is the global logger to be shared by all the children
var Logger *zap.SugaredLogger

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/quote_graph/")
	viper.AddConfigPath("$HOME/.quote_graph")
	viper.AddConfigPath(".")

	viper.SetDefault("interactive", true)
	viper.SetDefault("raven.dsn", nil)
	viper.SetDefault("debug", false)
	viper.SetDefault("environment", "dev")

	err := viper.ReadInConfig()

	if err != nil {
		panic("Could not load in config file")
	}

	var logger *zap.Logger
	if viper.GetBool("interactive") {
		logger, err = zap.NewDevelopment()

	} else {
		logger, err = zap.NewProduction()
	}

	if err != nil {
		panic(err)
	}
	Logger = logger.Sugar()

	if viper.GetString("environment") != "dev" {
		dsn := viper.Get("raven.dsn")
		if dsn != nil {
			raven.SetDSN(dsn.(string))
		}
	}
}
