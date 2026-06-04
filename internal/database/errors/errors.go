package dberrors

import (
	"errors"
)

var (
	ErrClientType    = errors.New("invalid client type")
	ErrInvalidConfig = errors.New("invalid config")

	ErrUnimplemented = errors.New("unimplemented")

	ErrBadRequest  = errors.New("bad request")
	ErrNotFound    = errors.New("not found")
	ErrInternal    = errors.New("internal")
	ErrUnavailable = errors.New("unavailable")
)
