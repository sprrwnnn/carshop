package domain

import "errors"

var (
	ErrBadRequest = errors.New("bad request")
	ErrInternal   = errors.New("internal error")
	ErrValidation = errors.New("validation error")
)

// ErrorResponse представляет стандартный ответ об ошибке
type ErrorResponse struct {
	Error   string   `json:"error"`
	Details []string `json:"details"`
}
