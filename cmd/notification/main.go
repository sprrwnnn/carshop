package main

import (
	"carshop/internal/events"
	"carshop/internal/events/rabbit"
	"carshop/internal/logger"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ilyakaznacheev/cleanenv"
	"go.uber.org/zap"
)

type notificationConfig struct {
	Env      string `yaml:"env" env-default:"local"`
	RabbitMQ struct {
		URL string `yaml:"-" env:"RABBITMQ_URL" env-default:"amqp://guest:guest@localhost:5672/"`
	} `yaml:"rabbitmq"`
}

type localDeploy struct {
	Deploy struct {
		Env []struct {
			Name  string `yaml:"name"`
			Value string `yaml:"value"`
		} `yaml:"env"`
	} `yaml:"deploy"`
}

func mustLoadConfig() *notificationConfig {
	var configPath string
	flag.StringVar(&configPath, "config_path", ".build/config/notification.yaml", "path to config file")
	flag.Parse()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic(fmt.Sprintf("config file not found: %s", configPath))
	}

	var deploy localDeploy
	if err := cleanenv.ReadConfig(configPath, &deploy); err == nil {
		for _, env := range deploy.Deploy.Env {
			_ = os.Setenv(env.Name, env.Value)
		}
	}

	var cfg notificationConfig
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic(err)
	}

	return &cfg
}

func main() {
	cfg := mustLoadConfig()

	log, err := logger.New(cfg.Env)
	if err != nil {
		panic(err)
	}

	log.Info("starting notification service")

	consumer, err := rabbit.NewConsumer(cfg.RabbitMQ.URL, log)
	if err != nil {
		panic(fmt.Errorf("failed to create consumer: %w", err))
	}
	defer consumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-shutdownCh
		log.With(zap.String("signal", sig.String())).Info("received signal, stopping")
		cancel()
	}()

	if err := consumer.Consume(ctx, handleCarCreated(log)); err != nil {
		log.With(zap.Error(err)).Error("consumer stopped with error")
	}

	log.Info("notification service stopped")
}

func handleCarCreated(log *zap.Logger) rabbit.EventHandler {
	return func(ctx context.Context, event events.CarCreatedEvent) error {
		log.With(
			zap.Uint64("id", event.ID),
			zap.String("name", event.Name),
			zap.String("colour", event.Colour),
			zap.Float64("price", event.Price),
			zap.String("build_date", event.BuildDate),
		).Info("notification: new car created")
		return nil
	}
}
