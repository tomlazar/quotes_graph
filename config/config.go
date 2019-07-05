package config

import (
	"go.uber.org/zap"

	"github.com/getsentry/raven-go"
	"github.com/spf13/viper"
)

// Logger is the global logger to be shared by all the children
var (
	Version = "v0.2.1"
)

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

	if viper.GetString("environment") != "dev" {
		dsn := viper.Get("raven.dsn")
		if dsn != nil {
			raven.SetDSN(dsn.(string))
		}
	}
}

// NewLogger creates a new logger with the context specified
func NewLogger(contxt string) *zap.SugaredLogger {
	var logger *zap.Logger
	var err error
	if viper.GetBool("interactive") {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}

	if err != nil {
		panic(err)
	}

	return logger.Sugar().With("context", contxt)
}
