package database

import (
	dbtypes "carshop/internal/database/types"
	"context"
)

type CarsReadRepo interface {
	GetCars(ctx context.Context, req *dbtypes.CarsGetRequest) (*dbtypes.CarsGetResponse, error)
	GetCarByID(ctx context.Context, req *dbtypes.CarGetByIDRequest) (*dbtypes.CarGetByIDResponse, error)
}

type CarsWriteRepo interface {
	ExecuteCreateCarCommand(ctx context.Context, req *dbtypes.CarCreateCommand) (*dbtypes.CarCreateCommandResponse, error)
	ExecuteUpdateCarCommand(ctx context.Context, req *dbtypes.CarUpdateCommand) (*dbtypes.CarUpdateCommandResponse, error)
	ExecuteDeleteCarCommand(ctx context.Context, req *dbtypes.CarDeleteCommand) (*dbtypes.CarDeleteCommandResponse, error)
}

type CarsProjectionRepo interface {
	RebuildCarReadModels(ctx context.Context) error
}
