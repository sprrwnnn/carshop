package cache

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/go-redis/redis/v7"
)

type KeyDBCacheConfig struct {
	Addrs       []string
	Password    string
	ReadOnly    bool
	DialTimeout time.Duration
	PoolSize    int
	DefaultTTL  time.Duration
}

func (c KeyDBCacheConfig) Validate() error {
	if c.Addrs == nil {
		return fmt.Errorf("%w: addrs is required", ErrInvalidConfig)
	}

	if c.Password == "" {
		return fmt.Errorf("%w: password is required", ErrInvalidConfig)
	}

	if c.PoolSize == 0 {
		return fmt.Errorf("%w: pool size is required", ErrInvalidConfig)
	}

	return nil
}

type keyDBClient interface {
	HSet(key string, values ...interface{}) *redis.IntCmd
	HGet(key string, field string) *redis.StringCmd
	HDel(key string, fields ...string) *redis.IntCmd
	Exists(keys ...string) *redis.IntCmd
	Ping() *redis.StatusCmd
}

type keyDBCache struct {
	config  KeyDBCacheConfig
	sfGroup singleflight.Group
	client  keyDBClient
}

func NewKeyDBCache(cfg KeyDBCacheConfig) (*keyDBCache, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}

	client := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:       cfg.Addrs,
		Password:    cfg.Password,
		ReadOnly:    cfg.ReadOnly,
		DialTimeout: cfg.DialTimeout,
		PoolSize:    cfg.PoolSize,
		TLSConfig:   nil,
	})

	if err := client.Ping().Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnect, err)
	}

	return &keyDBCache{
		config: cfg,
		client: client,
	}, nil
}

func (c *keyDBCache) HGet(ctx context.Context, valueType ValueType, hashSetName, field string) (string, error) {
	if err := c.checkInstance(); err != nil {
		return "", fmt.Errorf("%w: %v", ErrGet, err)
	}

	val, err := c.client.HGet(getKey(hashSetName, valueType), field).Result()
	if err == redis.Nil {
		return "", ErrKeyNotFound
	}

	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrGet, err)
	}

	return val, nil
}

func (c *keyDBCache) HSet(ctx context.Context, valueType ValueType, hashSetName, field string, value any) error {
	if err := c.checkInstance(); err != nil {
		return fmt.Errorf("%w: %v", ErrSet, err)
	}

	var (
		key   = getKey(hashSetName, valueType)
		sfKey = fmt.Sprintf("%s:%s", key, field)
	)

	defer c.sfGroup.Forget(sfKey)

	_, err, _ := c.sfGroup.Do(sfKey, func() (any, error) {
		if err := c.client.HSet(key, field, value).Err(); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrSet, err)
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *keyDBCache) HDel(ctx context.Context, valueType ValueType, hashSetName, field string) error {
	if err := c.checkInstance(); err != nil {
		return fmt.Errorf("%w: %v", ErrDel, err.Error())
	}

	var (
		key   = getKey(hashSetName, valueType)
		sfKey = fmt.Sprintf("%s:%s", key, field)
	)

	defer c.sfGroup.Forget(sfKey)

	_, err, _ := c.sfGroup.Do(sfKey, func() (any, error) {
		if res, err := c.client.HDel(key, field).Result(); err != nil || res == 0 {
			if res == 0 {
				return nil, ErrKeyNotFound
			}

			return nil, fmt.Errorf("%w: %s", ErrDel, err.Error())
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *keyDBCache) checkInstance() error {
	if c == nil {
		return ErrInvalidInstance
	}

	if s := c.client.Ping(); s.Err() != nil {
		return fmt.Errorf("%w: %v", ErrInvalidInstance, s.Err().Error())
	}

	return nil
}
