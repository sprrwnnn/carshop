package cars

import (
	"carshop/internal/events"
	"context"
)

type CarsQueryService interface {
	GetCars(ctx context.Context, req CarsGetQuery) (*CarsGetResponse, error)
	GetCarByID(ctx context.Context, req CarGetByIDQuery) (*CarGetByIDResponse, error)
}

type CarsCommandService interface {
	CreateCar(ctx context.Context, cmd CreateCarCommand) (*CarCreateResponse, error)
	UpdateCar(ctx context.Context, cmd UpdateCarCommand) (*CarUpdateResponse, error)
	DeleteCar(ctx context.Context, cmd DeleteCarCommand) (*CarDeleteResponse, error)
}

type EventPublisher interface {
	PublishCarCreated(ctx context.Context, event events.CarCreatedEvent) error
}
