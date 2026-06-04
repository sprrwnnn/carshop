package pgdb

import (
	dberrors "carshop/internal/database/errors"
	dbtypes "carshop/internal/database/types"
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

func (c *client) GetCars(
	ctx context.Context,
	req *dbtypes.CarsGetRequest,
) (*dbtypes.CarsGetResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: request is nil", dberrors.ErrBadRequest)
	}

	conn, err := c.pool.PeekReadConn()
	if err != nil {
		return nil, fmt.Errorf("%w: error getting read connection: %v", dberrors.ErrInternal, err)
	}

	defer conn.Release()

	res, err := c.getCars(ctx, conn, req)
	if err != nil {
		return nil, fmt.Errorf("error getting cars: %w", err)
	}

	return res, nil
}

func (c *client) getCars(
	ctx context.Context,
	driver Driver,
	req *dbtypes.CarsGetRequest,
	opts ...SelectOption,
) (*dbtypes.CarsGetResponse, error) {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	getCars := psql.Select(
		"id", "name", "colour", "price", "build_date",
	).
		From(carReadModelsTableName).
		Where(squirrel.Eq{"deleted_at": nil}).
		OrderBy("id ASC").
		Limit(req.Limit)

	if req.PriceFrom != nil {
		getCars = getCars.Where(squirrel.GtOrEq{"price": *req.PriceFrom})
	}

	if req.PriceTo != nil {
		getCars = getCars.Where(squirrel.LtOrEq{"price": *req.PriceTo})
	}

	if req.Cursor != nil {
		getCars = getCars.Where(squirrel.Gt{"id": *req.Cursor})
	}

	for _, opt := range opts {
		opt(&getCars)
	}

	query, args, err := getCars.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%w: error building query: %v", dberrors.ErrInternal, err)
	}

	rows, err := driver.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: error executing query: %v", dberrors.ErrInternal, err)
	}

	defer rows.Close()

	cars := make([]dbtypes.Car, 0)

	for rows.Next() {
		var car dbtypes.Car

		if err := rows.Scan(
			&car.ID,
			&car.Name,
			&car.Colour,
			&car.Price,
			&car.BuildDate,
		); err != nil {
			return nil, fmt.Errorf("%w: error scanning row: %v", dberrors.ErrInternal, err)
		}

		cars = append(cars, car)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: error iterating over rows: %v", dberrors.ErrInternal, err)
	}

	if len(cars) == 0 {
		return nil, dberrors.ErrNotFound
	}

	res := dbtypes.CarsGetResponse{
		Cars: cars,
	}

	if len(cars) == int(req.Limit) {
		res.NextCursor = &cars[len(cars)-1].ID
	}

	return &res, nil
}

func (c *client) GetCarByID(
	ctx context.Context,
	req *dbtypes.CarGetByIDRequest,
) (*dbtypes.CarGetByIDResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: request is nil", dberrors.ErrBadRequest)
	}

	conn, err := c.pool.PeekReadConn()
	if err != nil {
		return nil, fmt.Errorf("%w: error getting read connection: %v", dberrors.ErrInternal, err)
	}

	defer conn.Release()

	res, err := c.getCarByID(ctx, conn, req)
	if err != nil {
		return nil, fmt.Errorf("error getting car by id: %w", err)
	}

	return res, nil
}

func (c *client) getCarByID(
	ctx context.Context,
	driver Driver,
	req *dbtypes.CarGetByIDRequest,
	opts ...SelectOption,
) (*dbtypes.CarGetByIDResponse, error) {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	getCarByID := psql.Select(
		"id", "name", "colour", "price", "build_date",
	).
		From(carReadModelsTableName).
		Where(squirrel.Eq{"id": req.ID}).
		Where(squirrel.Eq{"deleted_at": nil})

	for _, opt := range opts {
		opt(&getCarByID)
	}

	query, args, err := getCarByID.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%w: error building query: %v", dberrors.ErrInternal, err)
	}

	var car dbtypes.Car

	if err := driver.QueryRow(ctx, query, args...).Scan(
		&car.ID,
		&car.Name,
		&car.Colour,
		&car.Price,
		&car.BuildDate,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, dberrors.ErrNotFound
		}

		return nil, fmt.Errorf("%w: error scanning row: %v", dberrors.ErrInternal, err)
	}

	return &dbtypes.CarGetByIDResponse{
		Car: car,
	}, nil
}
