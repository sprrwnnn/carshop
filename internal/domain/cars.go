package domain

import (
	"carshop/internal/services/cars"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
)

type Car struct {
	ID        uint64  `json:"id"`
	Name      string  `json:"name"`
	Colour    string  `json:"colour"`
	Price     float64 `json:"price"`
	BuildDate string  `json:"build_date"`
}

func (*Car) From(msg *cars.Car) *Car {
	if msg == nil {
		return &Car{}
	}

	return &Car{
		ID:        msg.ID,
		Name:      msg.Name,
		Colour:    msg.Colour,
		Price:     msg.Price,
		BuildDate: msg.BuildDate,
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

func (*CarsGetRequest) From(r *http.Request) (*CarsGetRequest, error) {
	if r == nil {
		return nil, fmt.Errorf("%w: request is nil", ErrBadRequest)
	}

	var res CarsGetRequest

	query := r.URL.Query()

	if len(query) == 0 {
		return &res, nil
	}

	if priceFrom := query.Get("price_from"); priceFrom != "" {
		price, err := strconv.ParseFloat(priceFrom, 64)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid price_from: %v", ErrBadRequest, err)
		}

		res.PriceFrom = &price
	}

	if priceTo := query.Get("price_to"); priceTo != "" {
		price, err := strconv.ParseFloat(priceTo, 64)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid price_to: %v", ErrBadRequest, err)
		}

		res.PriceTo = &price
	}

	if cursor := query.Get("cursor"); cursor != "" {
		cursor, err := strconv.ParseUint(cursor, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid cursor: %v", ErrBadRequest, err)
		}

		res.Cursor = &cursor
	}

	return &res, nil
}

func (r *CarsGetResponse) From(msg *cars.CarsGetResponse) *CarsGetResponse {
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
		ID uint64
	}

	CarGetByIDResponse struct {
		Car Car `json:"car"`
	}
)

func (*CarGetByIDRequest) From(r *http.Request) (*CarGetByIDRequest, error) {
	if r == nil {
		return nil, fmt.Errorf("%w: request is nil", ErrBadRequest)
	}

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid id: %v", ErrBadRequest, err)
	}

	res := CarGetByIDRequest{
		ID: id,
	}

	return &res, nil
}

func (r *CarGetByIDResponse) From(msg *cars.CarGetByIDResponse) *CarGetByIDResponse {
	if msg == nil {
		return nil
	}

	return &CarGetByIDResponse{
		Car: *new(Car).From(&msg.Car),
	}
}

type (
	CarCreateRequest struct {
		Name      string    `json:"name"`
		Colour    string    `json:"colour"`
		Price     float64   `json:"price"`
		BuildDate time.Time `json:"build_date"`
	}

	CarCreateResponse struct {
		ID uint64 `json:"id"`
	}
)

func (*CarCreateRequest) From(r *http.Request) (*CarCreateRequest, error) {
	if r == nil {
		return nil, fmt.Errorf("%w: request is nil", ErrBadRequest)
	}

	var res CarCreateRequest

	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("%w: error decoding request: %v", ErrBadRequest, err)
	}

	defer r.Body.Close()

	return &res, nil
}

func (*CarCreateResponse) From(msg *cars.CarCreateResponse) *CarCreateResponse {
	if msg == nil {
		return nil
	}

	return &CarCreateResponse{
		ID: msg.ID,
	}
}

type (
	CarUpdateRequest struct {
		ID        uint64     `json:"-"`
		Name      *string    `json:"name,omitempty"`
		Colour    *string    `json:"colour,omitempty"`
		Price     *float64   `json:"price,omitempty"`
		BuildDate *time.Time `json:"build_date,omitempty"`
	}
)

func (*CarUpdateRequest) From(r *http.Request) (*CarUpdateRequest, error) {
	if r == nil {
		return nil, fmt.Errorf("%w: request is nil", ErrBadRequest)
	}

	var res CarUpdateRequest

	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("%w: error decoding request: %v", ErrBadRequest, err)
	}

	defer r.Body.Close()

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid id: %v", ErrBadRequest, err)
	}

	res.ID = id

	return &res, nil
}

type (
	CarDeleteRequest struct {
		ID uint64
	}
)

func (*CarDeleteRequest) From(r *http.Request) (*CarDeleteRequest, error) {
	if r == nil {
		return nil, fmt.Errorf("%w: request is nil", ErrBadRequest)
	}

	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid id: %v", ErrBadRequest, err)
	}

	res := CarDeleteRequest{
		ID: id,
	}

	return &res, nil
}
