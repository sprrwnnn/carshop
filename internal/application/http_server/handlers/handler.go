package handlers

import (
	"carshop/internal/config"
	"carshop/internal/constants"
	"carshop/internal/services/cars"
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

type Handler struct {
	env    config.Env
	logger *zap.Logger

	carsQueryService   cars.CarsQueryService
	carsCommandService cars.CarsCommandService
}

type Config struct {
	Env    config.Env
	Logger *zap.Logger

	CarsQueryService   cars.CarsQueryService
	CarsCommandService cars.CarsCommandService
}

func (c Config) Validate() error {
	if c.Logger == nil {
		return errors.New("logger is required")
	}

	if c.CarsQueryService == nil {
		return errors.New("cars query service is required")
	}

	if c.CarsCommandService == nil {
		return errors.New("cars command service is required")
	}

	return nil
}

func NewHandler(c *Config) (*Handler, error) {
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %v", err)
	}

	return &Handler{
		env:    c.Env,
		logger: c.Logger,

		carsQueryService:   c.CarsQueryService,
		carsCommandService: c.CarsCommandService,
	}, nil
}

func (h *Handler) GetBuildInfoHandler(w http.ResponseWriter, r *http.Request) {
	writeErrorResponse(w, constants.ErrInternalServerError, http.StatusInternalServerError)
}
