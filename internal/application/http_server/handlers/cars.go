package handlers

import (
	"carshop/internal/constants"
	"carshop/internal/domain"
	"carshop/internal/services/cars"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

// =============== CARS HANDLERS ===============

// @Summary Get cars
// @Description Returns cars
// @Tags Cars
// @Produce application/json
// @Param price_from query string false "Price from to filter"
// @Param price_to query string false "Price to to filter"
// @Param cursor query string false "Cursor"
// @Success 200 {object} domain.CarsGetResponse "Cars"
// @Failure 400 {object} domain.ErrorResponse "Bad request"
// @Failure 404 {object} domain.ErrorResponse "Cars not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/cars/q [get]
func (h *Handler) GetCarsHandler(w http.ResponseWriter, r *http.Request) {
	const handlerName = "GetCarsHandler"

	req, err := new(domain.CarsGetRequest).From(r)
	if err != nil {
		writeErrorResponse(w, constants.ErrBadRequest, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.carsQueryService.GetCars(r.Context(), cars.CarsGetQuery{
		PriceFrom: req.PriceFrom,
		PriceTo:   req.PriceTo,

		Cursor: req.Cursor,
	})
	if err != nil {
		h.handleCarsError(w, err, handlerName)
		h.logger.With(
			zap.String("error", err.Error()),
			zap.String("request_id", r.RequestURI),
		).Error("failed to get cars")

		return
	}

	result := new(domain.CarsGetResponse).From(resp)

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(result); err != nil {
		h.logger.With(
			zap.String("error", err.Error()),
			zap.String("request_id", r.RequestURI),
		).Error("failed to encode response")

		writeErrorResponse(w, constants.ErrInternalServerError, http.StatusInternalServerError, "error marshalling response")
	}
}

// @Summary Get car by ID
// @Description Returns car by its ID
// @Tags Cars
// @Produce application/json
// @Param id path string true "Car ID"
// @Success 200 {object} domain.CarGetByIDResponse "Car"
// @Failure 400 {object} domain.ErrorResponse "Bad request"
// @Failure 404 {object} domain.ErrorResponse "Cars not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/cars/q/{id} [get]
func (h *Handler) GetCarByIDHandler(w http.ResponseWriter, r *http.Request) {
	const handlerName = "GetCarByIDHandler"

	req, err := new(domain.CarGetByIDRequest).From(r)
	if err != nil {
		writeErrorResponse(w, constants.ErrBadRequest, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.carsQueryService.GetCarByID(r.Context(), cars.CarGetByIDQuery{
		ID: req.ID,
	})
	if err != nil {
		h.handleCarsError(w, err, handlerName)
		h.logger.With(
			zap.String("error", err.Error()),
			zap.String("request_id", r.RequestURI),
		).Error("failed to get car by id")

		return
	}

	result := new(domain.CarGetByIDResponse).From(resp)

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(result); err != nil {
		h.logger.With(
			zap.String("error", err.Error()),
			zap.String("request_id", r.RequestURI),
		).Error("failed to encode response")

		writeErrorResponse(w, constants.ErrInternalServerError, http.StatusInternalServerError, "error marshalling response")
	}
}

// @Summary Create car
// @Description Creates new car
// @Tags Cars
// @Accept application/json
// @Produce application/json
// @Param request body domain.CarCreateRequest true "Car"
// @Success 201 {object} domain.CarCreateResponse "Car"
// @Failure 400 {object} domain.ErrorResponse "Bad request"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/cars/c [post]
func (h *Handler) CreateCarHandler(w http.ResponseWriter, r *http.Request) {
	const handlerName = "CreateCarHandler"

	req, err := new(domain.CarCreateRequest).From(r)
	if err != nil {
		writeErrorResponse(w, constants.ErrBadRequest, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.carsCommandService.CreateCar(r.Context(), cars.CreateCarCommand{
		Name:      req.Name,
		Colour:    req.Colour,
		Price:     req.Price,
		BuildDate: req.BuildDate,
	})
	if err != nil {
		h.handleCarsError(w, err, handlerName)
		h.logger.With(
			zap.String("error", err.Error()),
			zap.String("request_id", r.RequestURI),
		).Error("failed to create car")

		return
	}

	result := new(domain.CarCreateResponse).From(resp)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(result); err != nil {
		h.logger.With(
			zap.String("error", err.Error()),
			zap.String("request_id", r.RequestURI),
		).Error("failed to encode response")

		writeErrorResponse(w, constants.ErrInternalServerError, http.StatusInternalServerError, "error marshalling response")
	}
}

// @Summary Update car
// @Description Updates car by its ID
// @Tags Cars
// @Accept application/json
// @Param id path string true "Car ID"
// @Param request body domain.CarUpdateRequest true "Car"
// @Success 204 "Car"
// @Failure 400 {object} domain.ErrorResponse "Bad request"
// @Failure 404 {object} domain.ErrorResponse "Cars not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/cars/c/{id} [patch]
func (h *Handler) UpdateCarHandler(w http.ResponseWriter, r *http.Request) {
	const handlerName = "UpdateCarHandler"

	req, err := new(domain.CarUpdateRequest).From(r)
	if err != nil {
		writeErrorResponse(w, constants.ErrBadRequest, http.StatusBadRequest, err.Error())
		return
	}

	if _, err := h.carsCommandService.UpdateCar(r.Context(), cars.UpdateCarCommand{
		ID:        req.ID,
		Name:      req.Name,
		Colour:    req.Colour,
		Price:     req.Price,
		BuildDate: req.BuildDate,
	}); err != nil {
		h.handleCarsError(w, err, handlerName)
		h.logger.With(
			zap.String("error", err.Error()),
			zap.String("request_id", r.RequestURI),
		).Error("failed to update car")

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Delete car
// @Description Deletes car by its ID
// @Tags Cars
// @Param id path string true "Car ID"
// @Success 204 "Car"
// @Failure 400 {object} domain.ErrorResponse "Bad request"
// @Failure 404 {object} domain.ErrorResponse "Cars not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /api/v1/cars/c/{id} [delete]
func (h *Handler) DeleteCarHandler(w http.ResponseWriter, r *http.Request) {
	const handlerName = "DeleteCarHandler"

	req, err := new(domain.CarDeleteRequest).From(r)
	if err != nil {
		writeErrorResponse(w, constants.ErrBadRequest, http.StatusBadRequest, err.Error())
		return
	}

	if _, err := h.carsCommandService.DeleteCar(r.Context(), cars.DeleteCarCommand{
		ID: req.ID,
	}); err != nil {
		h.handleCarsError(w, err, handlerName)
		h.logger.With(
			zap.String("error", err.Error()),
			zap.String("request_id", r.RequestURI),
		).Error("failed to delete car")

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
