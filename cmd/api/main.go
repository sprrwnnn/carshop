// @title Carshop Backend
// @version 1.0
// @description Carshop Backend
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8000
// @BasePath /
// @schemes http https

package main

import (
	"carshop/internal/application"
	"carshop/internal/config"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "carshop/internal/application/http_server/swagger/docs"

	"carshop/internal/logger"

	"go.uber.org/zap"
)

func main() {
	cfg := config.MustLoad()

	log, err := logger.New(cfg.Env)
	if err != nil {
		panic(err)
	}

	log.With(
		zap.String("version", cfg.Build.Version),
		zap.String("commit", cfg.Build.Commit),
		zap.String("build_date", cfg.Build.Date),
	).Info("Starting App")

	app, err := application.New(cfg, log)
	if err != nil {
		panic(fmt.Errorf("error creating application %v", err))
	}

	app.MustStart()

	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, os.Interrupt, syscall.SIGTERM)

	sig := <-shutdownSignal
	log.With(
		zap.String("signal", sig.String()),
	).Info("received signal. stopping application")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	app.Stop(ctx)
}
