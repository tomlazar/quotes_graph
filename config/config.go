package config

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/getsentry/raven-go"
	"github.com/spf13/viper"

	lumberjack "gopkg.in/natefinch/lumberjack.v2"
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
func NewLogger(context string) *zap.SugaredLogger {
	filename := fmt.Sprintf("/logs/%v.log", context)
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filename,
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
	})

	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel
	})

	consoleDebugging := zapcore.Lock(os.Stdout)
	consoleErrors := zapcore.Lock(os.Stderr)

	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())

	core := zapcore.NewTee(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(zap.NewDevelopmentEncoderConfig()),
			w,
			zap.DebugLevel,
		),
		zapcore.NewCore(consoleEncoder, consoleErrors, highPriority),
		zapcore.NewCore(consoleEncoder, consoleDebugging, lowPriority),
	)
	return zap.New(core).Sugar()
}
