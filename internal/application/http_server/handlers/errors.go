package handlers

import (
	"carshop/internal/constants"
	"carshop/internal/domain"
	"carshop/internal/services/cars"
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/zap"
)

// =============== ERROR HANDLERS ===============

// writeErrorResponse записывает унифицированный JSON ответ об ошибке
func writeErrorResponse(w http.ResponseWriter, err error, code int, details ...string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := domain.ErrorResponse{
		Error:   err.Error(),
		Details: details,
	}

	_ = json.NewEncoder(w).Encode(response)
}

func (h *Handler) handleCarsError(w http.ResponseWriter, err error, handlerName string) {
	switch {
	case errors.Is(err, cars.ErrBadRequest):
		writeErrorResponse(w, constants.ErrBadRequest, http.StatusBadRequest, err.Error())
	case errors.Is(err, cars.ErrNotFound):
		writeErrorResponse(w, constants.ErrNotFound, http.StatusNotFound, err.Error())
	case errors.Is(err, cars.ErrInternal):
		writeErrorResponse(w, constants.ErrInternalServerError, http.StatusInternalServerError, err.Error())
	default:
		h.logger.With(
			zap.String("error", err.Error()),
		).Error("Unhandled cars error")

		writeErrorResponse(w, constants.ErrInternalServerError, http.StatusInternalServerError, err.Error())
	}
}
