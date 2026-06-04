package application

import (
	httpserver "carshop/internal/application/http_server"
	"carshop/internal/config"
	"context"
	"fmt"

	"go.uber.org/zap"
)

type Application struct {
	httpServer *httpserver.Server
}

func New(cfg *config.Config, logger *zap.Logger) (*Application, error) {
	httpServer, err := httpserver.New(&httpserver.ServerConfig{
		Env:          cfg.Env,
		Build:        cfg.Build,
		ServerConfig: cfg.Server,

		DBConfig:    cfg.Database,
		CacheConfig: cfg.ValkeyCache,
		RabbitMQ:    cfg.RabbitMQ,

		Logger: logger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP server: %w", err)
	}

	return &Application{
		httpServer: httpServer,
	}, nil
}

func (a *Application) MustStart() {
	go a.httpServer.MustStart()
}

func (a *Application) Stop(ctx context.Context) {
	_ = a.httpServer.Shutdown(ctx)
}
