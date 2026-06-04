package cars

import (
	dbtypes "carshop/internal/database/types"
	"errors"
	"regexp"
	"strings"
	"time"
)

var colourRegex = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

const defaultCarsLimit = 10

type CarID = uint64

type Car struct {
	ID        CarID   `json:"id"`
	Name      string  `json:"name"`
	Colour    string  `json:"colour"`
	Price     float64 `json:"price"`
	BuildDate string  `json:"build_date"`
}

func (c *Car) From(msg *dbtypes.Car) *Car {
	if msg == nil {
		return &Car{}
	}

	return &Car{
		ID:        msg.ID,
		Name:      msg.Name,
		Colour:    msg.Colour,
		Price:     msg.Price,
		BuildDate: msg.BuildDate.Format(time.RFC3339),
	}
}

type (
	CarsGetRequest struct {
		PriceFrom *float64
		PriceTo   *float64

		Cursor *uint64
	}

	CarsGetResponse struct {
		Cars       []Car   `json:"cars"`
		NextCursor *uint64 `json:"next_cursor"`
	}
)

type CarsGetQuery = CarsGetRequest

func (r *CarsGetRequest) Validate() error {
	if r == nil {
		return errors.New("request is nil")
	}

	if r.PriceFrom != nil && r.PriceTo != nil && *r.PriceFrom > *r.PriceTo {
		return errors.New("price_from must be less than price_to")
	}

	return nil
}

func (*CarsGetResponse) From(msg *dbtypes.CarsGetResponse) *CarsGetResponse {
	if msg == nil {
		return nil
	}

	cars := make([]Car, len(msg.Cars))

	for i, car := range msg.Cars {
		cars[i] = *new(Car).From(&car)
	}

	return &CarsGetResponse{
		Cars:       cars,
		NextCursor: msg.NextCursor,
	}
}

type (
	CarGetByIDRequest struct {
		ID CarID
	}

	CarGetByIDResponse struct {
		Car Car `json:"car"`
	}
)

type CarGetByIDQuery = CarGetByIDRequest

func (r *CarGetByIDRequest) Validate() error {
	if r == nil {
		return errors.New("request is nil")
	}

	return nil
}

func (*CarGetByIDResponse) From(msg *dbtypes.CarGetByIDResponse) *CarGetByIDResponse {
	if msg == nil {
		return nil
	}

	return &CarGetByIDResponse{
		Car: *new(Car).From(&msg.Car),
	}
}

type (
	CarCreateRequest struct {
		Name      string
		Colour    string
		Price     float64
		BuildDate time.Time
	}

	CarCreateResponse struct {
		ID CarID
	}
)

type CreateCarCommand = CarCreateRequest

func (r *CarCreateRequest) Validate() error {
	if r == nil {
		return errors.New("request is nil")
	}

	if strings.TrimSpace(r.Name) == "" {
		return errors.New("name is required")
	}

	if !colourRegex.MatchString(r.Colour) {
		return errors.New("colour must be in hex format")
	}

	if r.Price <= 0 {
		return errors.New("price must be greater than 0")
	}

	if r.BuildDate.IsZero() {
		return errors.New("build_date is required")
	}

	return nil
}

func (*CarCreateResponse) From(msg *dbtypes.CarCreateCommandResponse) *CarCreateResponse {
	if msg == nil {
		return nil
	}

	return &CarCreateResponse{
		ID: msg.ID,
	}
}

type (
	CarUpdateRequest struct {
		ID        CarID
		Name      *string
		Colour    *string
		Price     *float64
		BuildDate *time.Time
	}

	CarUpdateResponse struct {
	}
)

type UpdateCarCommand = CarUpdateRequest

func (r *CarUpdateRequest) Validate() error {
	if r == nil {
		return errors.New("request is nil")
	}

	if r.ID == 0 {
		return errors.New("id is required")
	}

	if r.Name != nil && strings.TrimSpace(*r.Name) == "" {
		return errors.New("name is required")
	}

	if r.Colour != nil && !colourRegex.MatchString(*r.Colour) {
		return errors.New("colour must be in hex format")
	}

	if r.Price != nil && *r.Price <= 0 {
		return errors.New("price must be greater than 0")
	}

	if r.BuildDate != nil && r.BuildDate.IsZero() {
		return errors.New("build_date is required")
	}

	return nil
}

type (
	CarDeleteRequest struct {
		ID CarID
	}

	CarDeleteResponse struct {
	}
)

type DeleteCarCommand = CarDeleteRequest

func (r *CarDeleteRequest) Validate() error {
	if r == nil {
		return errors.New("request is nil")
	}

	return nil
}
