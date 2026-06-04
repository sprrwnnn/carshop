package cars

import (
	"carshop/internal/cache"
	"carshop/internal/database"
	dberrors "carshop/internal/database/errors"
	dbtypes "carshop/internal/database/types"
	"carshop/internal/events"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"go.uber.org/zap"
)

type CarsServiceConfig struct {
	CarsReadRepo       database.CarsReadRepo
	CarsWriteRepo      database.CarsWriteRepo
	CarsProjectionRepo database.CarsProjectionRepo
	Cache              cache.Client
	Publisher          EventPublisher

	Logger *zap.Logger
}

func (c CarsServiceConfig) Validate() error {
	if c.CarsReadRepo == nil {
		return errors.New("cars read repo is required")
	}

	if c.CarsWriteRepo == nil {
		return errors.New("cars write repo is required")
	}

	if c.CarsProjectionRepo == nil {
		return errors.New("cars projection repo is required")
	}

	if c.Cache == nil {
		return errors.New("cache is required")
	}

	if c.Publisher == nil {
		return errors.New("publisher is required")
	}

	if c.Logger == nil {
		return errors.New("logger is required")
	}

	return nil
}

type DBCacheCarsService struct {
	carsReadRepo       database.CarsReadRepo
	carsWriteRepo      database.CarsWriteRepo
	carsProjectionRepo database.CarsProjectionRepo
	cache              cache.Client
	publisher          EventPublisher

	logger *zap.Logger
}

func NewDBCacheCarsService(cfg CarsServiceConfig) (*DBCacheCarsService, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &DBCacheCarsService{
		carsReadRepo:       cfg.CarsReadRepo,
		carsWriteRepo:      cfg.CarsWriteRepo,
		carsProjectionRepo: cfg.CarsProjectionRepo,
		cache:              cfg.Cache,
		publisher:          cfg.Publisher,

		logger: cfg.Logger,
	}, nil
}

const (
	carsHashSetName    = "cars"
	carByIDHashSetName = "car_by_id"
)

func (s *DBCacheCarsService) GetCars(ctx context.Context, req CarsGetRequest) (*CarsGetResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrBadRequest, err)
	}

	useCached := req.PriceFrom == nil && req.PriceTo == nil

	var field string

	if req.Cursor == nil {
		field = "0"
	} else {
		field = fmt.Sprintf("%d", *req.Cursor)
	}

	if useCached {
		if cachedResp, err := s.cache.HGet(ctx, cache.CarsValueType, carsHashSetName, field); err == nil {
			var resp CarsGetResponse

			if err := json.Unmarshal([]byte(cachedResp), &resp); err == nil {
				s.logger.Debug("cache hit; using cached response")

				return &resp, nil
			}

			_ = s.cache.HDel(ctx, cache.CarsValueType, carsHashSetName, field)
		}

		s.logger.Debug("cache miss; getting data from database")
	}

	dbResp, err := s.carsReadRepo.GetCars(ctx, &dbtypes.CarsGetRequest{
		PriceFrom: req.PriceFrom,
		PriceTo:   req.PriceTo,

		Limit:  defaultCarsLimit,
		Cursor: req.Cursor,
	})
	if err != nil && !errors.Is(err, dberrors.ErrNotFound) {
		s.logger.With(
			zap.String("error", err.Error()),
		).Error("failed to get cars from database")

		return nil, s.handleDBError(err)
	}

	if errors.Is(err, dberrors.ErrNotFound) {
		return nil, ErrNotFound
	}

	s.logger.Debug("got data from database")

	result := new(CarsGetResponse).From(dbResp)

	if useCached {
		cacheBody, err := json.Marshal(result)
		if err != nil {
			return result, nil
		}

		s.logger.Debug("caching response")

		err = s.cache.HSet(ctx, cache.CarsValueType, carsHashSetName, field, string(cacheBody))

		if err != nil {
			s.logger.With(
				zap.String("error", err.Error()),
			).Error("failed to cache response")
		}
	}

	return result, nil
}

func (s *DBCacheCarsService) GetCarByID(ctx context.Context, req CarGetByIDRequest) (*CarGetByIDResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrBadRequest, err)
	}

	field := fmt.Sprintf("%d", req.ID)

	if cachedResp, err := s.cache.HGet(ctx, cache.CarsValueType, carByIDHashSetName, field); err == nil {
		var resp CarGetByIDResponse

		if err := json.Unmarshal([]byte(cachedResp), &resp.Car); err == nil {
			s.logger.Debug("cache hit; using cached response")

			return &resp, nil
		}

		_ = s.cache.HDel(ctx, cache.CarsValueType, carByIDHashSetName, field)
	}

	s.logger.Debug("cache miss; getting data from database")

	dbResp, err := s.carsReadRepo.GetCarByID(ctx, &dbtypes.CarGetByIDRequest{
		ID: req.ID,
	})
	if err != nil && !errors.Is(err, dberrors.ErrNotFound) {
		s.logger.With(
			zap.String("error", err.Error()),
		).Error("failed to get car by id from database")

		return nil, s.handleDBError(err)
	}

	if errors.Is(err, dberrors.ErrNotFound) {
		return nil, ErrNotFound
	}

	s.logger.Debug("got data from database")

	result := new(CarGetByIDResponse).From(dbResp)

	cacheBody, err := json.Marshal(result.Car)
	if err != nil {
		return result, nil
	}

	s.logger.Debug("caching response")

	_ = s.cache.HSet(ctx, cache.CarsValueType, carByIDHashSetName, field, string(cacheBody))

	return result, nil
}

func (s *DBCacheCarsService) CreateCar(ctx context.Context, req CarCreateRequest) (*CarCreateResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrBadRequest, err)
	}

	dbResp, err := s.carsWriteRepo.ExecuteCreateCarCommand(ctx, &dbtypes.CarCreateCommand{
		Name:      req.Name,
		Colour:    req.Colour,
		Price:     req.Price,
		BuildDate: req.BuildDate,
	})
	if err != nil {
		s.logger.With(
			zap.String("error", err.Error()),
		).Error("failed to create car in database")

		return nil, s.handleDBError(err)
	}

	result := new(CarCreateResponse).From(dbResp)

	if err := s.rebuildReadModels(ctx); err != nil {
		return nil, err
	}

	if err := s.cache.HDel(ctx, cache.CarsValueType, carsHashSetName, fmt.Sprintf("%d", (result.ID/defaultCarsLimit)*defaultCarsLimit)); err != nil {
		s.logger.With(
			zap.String("error", err.Error()),
		).Error("failed to delete last batch of cars in cache")
	}

	if err := s.publisher.PublishCarCreated(ctx, events.CarCreatedEvent{
		ID:        result.ID,
		Name:      req.Name,
		Colour:    req.Colour,
		Price:     req.Price,
		BuildDate: req.BuildDate.Format("2006-01-02"),
	}); err != nil {
		s.logger.With(zap.Error(err)).Error("failed to publish car created event")
	}

	return result, nil
}

func (s *DBCacheCarsService) UpdateCar(ctx context.Context, req CarUpdateRequest) (*CarUpdateResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrBadRequest, err)
	}

	if _, err := s.carsWriteRepo.ExecuteUpdateCarCommand(ctx, &dbtypes.CarUpdateCommand{
		ID:        req.ID,
		Name:      req.Name,
		Colour:    req.Colour,
		Price:     req.Price,
		BuildDate: req.BuildDate,
	}); err != nil && !errors.Is(err, dberrors.ErrNotFound) {
		s.logger.With(
			zap.String("error", err.Error()),
		).Error("failed to update car in database")

		return nil, s.handleDBError(err)
	} else if errors.Is(err, dberrors.ErrNotFound) {
		return nil, ErrNotFound
	}

	if err := s.rebuildReadModels(ctx); err != nil {
		return nil, err
	}

	s.logger.Debug("updating car in cache")

	if err := s.cache.HDel(ctx, cache.CarsValueType, carByIDHashSetName, fmt.Sprintf("%d", req.ID)); err != nil {
		s.logger.With(
			zap.String("error", err.Error()),
		).Error("failed to delete car in cache with cars by ID")
	}

	if err := s.cache.HDel(ctx, cache.CarsValueType, carsHashSetName, fmt.Sprintf("%d", (req.ID/defaultCarsLimit)*defaultCarsLimit)); err != nil {
		s.logger.With(
			zap.String("error", err.Error()),
		).Error("failed to delete car from cache with car batch")
	}

	return &CarUpdateResponse{}, nil
}

func (s *DBCacheCarsService) DeleteCar(ctx context.Context, req CarDeleteRequest) (*CarDeleteResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrBadRequest, err)
	}

	if _, err := s.carsWriteRepo.ExecuteDeleteCarCommand(ctx, &dbtypes.CarDeleteCommand{
		ID: req.ID,
	}); err != nil && !errors.Is(err, dberrors.ErrNotFound) {
		s.logger.With(
			zap.String("error", err.Error()),
		).Error("failed to delete car in database")

		return nil, s.handleDBError(err)
	} else if errors.Is(err, dberrors.ErrNotFound) {
		return nil, ErrNotFound
	}

	if err := s.rebuildReadModels(ctx); err != nil {
		return nil, err
	}

	if err := s.cache.HDel(ctx, cache.CarsValueType, carByIDHashSetName, fmt.Sprintf("%d", req.ID)); err != nil {
		s.logger.With(
			zap.String("error", err.Error()),
		).Error("failed to delete car from cache")
	}

	if err := s.cache.HDel(ctx, cache.CarsValueType, carsHashSetName, fmt.Sprintf("%d", (req.ID/defaultCarsLimit)*defaultCarsLimit)); err != nil {
		s.logger.With(
			zap.String("error", err.Error()),
		).Error("failed to delete car from cache")
	}

	return &CarDeleteResponse{}, nil
}

func (s *DBCacheCarsService) rebuildReadModels(ctx context.Context) error {
	if err := s.carsProjectionRepo.RebuildCarReadModels(ctx); err != nil {
		s.logger.With(zap.Error(err)).Error("failed to rebuild cars read models from event store")

		return s.handleDBError(err)
	}

	return nil
}

func (s *DBCacheCarsService) handleDBError(err error) error {
	switch {
	case errors.Is(err, dberrors.ErrBadRequest):
		return fmt.Errorf("%w: %v", ErrBadRequest, err)
	case errors.Is(err, dberrors.ErrNotFound):
		return fmt.Errorf("%w: %v", ErrNotFound, err)
	case errors.Is(err, dberrors.ErrInternal):
		return fmt.Errorf("%w: %v", ErrInternal, err)
	default:
		return fmt.Errorf("%w: %v", ErrInternal, err)
	}
}
