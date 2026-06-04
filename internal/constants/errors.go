package constants

import "errors"

var (
	ErrInternalServerError = errors.New("internal server error")
	ErrBadRequest          = errors.New("bad request")
	ErrForbidden           = errors.New("forbidden")
	ErrNotFound            = errors.New("not found")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrConflict            = errors.New("conflict")
	ErrRequestTimeout      = errors.New("request timeout")
	ErrNotImplemented      = errors.New("not implemented")
	ErrBadGateway          = errors.New("bad gateway")
	ErrServiceUnavailable  = errors.New("service unavailable")
	ErrIntOutOfRange       = errors.New("value is out of range for int32")
)
