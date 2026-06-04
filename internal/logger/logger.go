package logger

import (
	"carshop/internal/config"
	"errors"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var ErrInvalidLogLevel = errors.New("invalid log level")

func New(env string) (*zap.Logger, error) {
	switch env {
	case config.EnvLocal:
		return zap.NewExample(), nil

	case config.EnvDebug:
		cfg := zap.NewDevelopmentConfig()

		cfg.OutputPaths = []string{"stdout"}
		cfg.ErrorOutputPaths = []string{"stderr"}
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		return cfg.Build()

	case config.EnvProduction:
		cfg := zap.NewProductionConfig()

		cfg.OutputPaths = []string{"stdout"}
		cfg.ErrorOutputPaths = []string{"stderr"}
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		return cfg.Build()

	default:
		return nil, ErrInvalidLogLevel
	}
}
