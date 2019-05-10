package config

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/getsentry/raven-go"
	"github.com/spf13/viper"
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

	fmt.Println("it works?")
	err := viper.ReadInConfig()

	if err != nil {
		logrus.Warnln("Could not load in config file")
	}

	if viper.GetBool("interactive") {
		logrus.SetFormatter(&logrus.TextFormatter{})
	}

	if viper.GetBool("debug") {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if viper.GetString("environment") != "dev" {
		dsn := viper.Get("raven.dsn")
		if dsn != nil {
			raven.SetDSN(dsn.(string))
		}
	}
}
