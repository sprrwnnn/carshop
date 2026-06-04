package cache

import (
	"context"
)

type Client interface {
	// HSet adds the value to a hash set defined by key and valueType
	//
	// If there is no such hash set defined, it is created
	HSet(ctx context.Context, valueType ValueType, hashSetName, field string, value any) error

	// HGet gets the value from a hash set defined by key and valueType
	//
	// If there is no such hash set defined, it returns ErrKeyNotFound
	HGet(ctx context.Context, valueType ValueType, hashSetName, field string) (string, error)

	// HDel deletes the value from a hash set defined by key and valueType
	//
	// If there is no such hash set defined, it returns ErrKeyNotFound
	HDel(ctx context.Context, valueType ValueType, hashSetName, field string) error
}
