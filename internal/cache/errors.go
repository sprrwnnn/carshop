package cache

import "errors"

var (
	ErrInvalidConfig   = errors.New("invalid config")
	ErrConnect         = errors.New("connecting to server")
	ErrInvalidInstance = errors.New("check instance failed")
	ErrKeyNotFound     = errors.New("key not found")
	ErrKeyExists       = errors.New("key already exists")
	ErrGet             = errors.New("failed to get value")
	ErrSet             = errors.New("failed to set value")
	ErrDel             = errors.New("failed to delete value")
)
