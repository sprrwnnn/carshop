package dbtypes

import "time"

type CarID = uint64

type Car struct {
	ID        CarID
	Name      string
	Colour    string
	Price     float64
	BuildDate time.Time
}

type (
	CarsGetRequest struct {
		PriceFrom *float64
		PriceTo   *float64

		Limit  uint64
		Cursor *uint64
	}

	CarsGetResponse struct {
		Cars       []Car
		NextCursor *uint64
	}
)

type (
	CarGetByIDRequest struct {
		ID CarID
	}

	CarGetByIDResponse struct {
		Car Car
	}
)

type (
	CarCreateCommand struct {
		Name      string
		Colour    string
		Price     float64
		BuildDate time.Time
	}

	CarCreateCommandResponse struct {
		ID CarID
	}
)

type (
	CarUpdateCommand struct {
		ID        CarID
		Name      *string
		Colour    *string
		Price     *float64
		BuildDate *time.Time
	}

	CarUpdateCommandResponse struct {
	}
)

type (
	CarDeleteCommand struct {
		ID CarID
	}

	CarDeleteCommandResponse struct {
	}
)
