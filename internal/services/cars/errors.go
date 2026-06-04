package cars

import "errors"

var (
	ErrBadRequest = errors.New("bad request")
	ErrNotFound   = errors.New("not found")
	ErrInternal   = errors.New("internal error")
)
