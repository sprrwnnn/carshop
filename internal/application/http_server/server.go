package httpserver

import (
	"carshop/internal/application/http_server/handlers"
	"carshop/internal/application/http_server/middleware"
	"carshop/internal/cache"
	"carshop/internal/config"
	pgdb "carshop/internal/database/postgres"
	"carshop/internal/events/rabbit"
	"carshop/internal/services/cars"
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Server struct {
	server    *http.Server
	publisher *rabbit.Publisher
	logger    *zap.Logger
}

type ServerConfig struct {
	Env          config.Env
	Build        config.BuildInfo
	ServerConfig config.Server

	DBConfig    config.Database
	CacheConfig config.Cache
	RabbitMQ    config.RabbitMQ

	Logger *zap.Logger
}

//nolint:funlen // server wiring keeps dependencies in one place
func New(cfg *ServerConfig) (*Server, error) {
	postgresDB, err := pgdb.NewClient(pgdb.Config{
		Hosts:    cfg.DBConfig.Hosts,
		Port:     cfg.DBConfig.Port,
		Username: cfg.DBConfig.Username,
		Password: cfg.DBConfig.Password,
		Database: cfg.DBConfig.DBName,
	}, cfg.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create database client: %w", err)
	}

	cfg.Logger.Debug("successfully created database client")

	cache, err := cache.NewKeyDBCache(cache.KeyDBCacheConfig{
		Addrs:       []string{cfg.CacheConfig.Addrs},
		Password:    cfg.CacheConfig.Password,
		ReadOnly:    cfg.CacheConfig.ReadOnly,
		DialTimeout: cfg.CacheConfig.DialTimeout,
		PoolSize:    cfg.CacheConfig.PoolSize,
		DefaultTTL:  cfg.CacheConfig.DefaultTTL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create cache client: %w", err)
	}

	cfg.Logger.Debug("successfully created cache client")

	publisher, err := rabbit.NewPublisher(cfg.RabbitMQ.URL, cfg.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create rabbit publisher: %w", err)
	}

	cfg.Logger.Debug("successfully created rabbit publisher")

	carsService, err := cars.NewDBCacheCarsService(cars.CarsServiceConfig{
		CarsReadRepo:       postgresDB,
		CarsWriteRepo:      postgresDB,
		CarsProjectionRepo: postgresDB,
		Cache:              cache,
		Publisher:          publisher,

		Logger: cfg.Logger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create cars service: %w", err)
	}

	cfg.Logger.Debug("successfully created cars service")

	handler, err := handlers.NewHandler(&handlers.Config{
		Env:    cfg.Env,
		Logger: cfg.Logger,

		CarsQueryService:   carsService,
		CarsCommandService: carsService,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create handler: %w", err)
	}

	cfg.Logger.Debug("successfully created handler")

	loggingMiddleware := middleware.NewLoggingMiddleware(cfg.Logger)
	metricsMiddleware := middleware.NewMetricsMiddleware(cfg.Env, cfg.Build)

	router, err := NewRouter(
		cfg.Env,
		handler,

		loggingMiddleware,
		metricsMiddleware,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create router: %w", err)
	}

	cfg.Logger.Debug("successfully created router")

	router.SetupV1()

	cfg.Logger.Debug("successfully setup router")

	server := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", cfg.ServerConfig.Host, cfg.ServerConfig.Port),
		Handler:           router.Router(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	cfg.Logger.Info("successfully configured HTTP server")

	return &Server{
		server:    server,
		publisher: publisher,
		logger:    cfg.Logger,
	}, nil
}

func (s *Server) MustStart() {
	if err := s.Start(); err != nil {
		panic(err)
	}
}

func (s *Server) Start() error {
	s.logger.With(
		zap.String("addr", s.server.Addr),
	).Info("starting HTTP server")

	if err := s.server.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down HTTP server")

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	s.publisher.Close()

	return nil
}
