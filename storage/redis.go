package storage

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"time"
)

// IRedisStorage specifies the contract to interact with the storage provider.
type IRedisStorage interface {
	Get(ctx context.Context, key string) (i interface{}, err error)
	Save(ctx context.Context, key string, i interface{}, ttl time.Duration) (err error)
}

type redisStorage struct {
	*redis.Client
}

// NewGoRedisV8 creates a new redis client for storage.Hub.
func NewGoRedisV8(c *redis.Client) IRedisStorage {
	if c == nil {
		return nil
	}

	return &redisStorage{
		Client: c,
	}
}

// Get returns the object from the Redis.
func (r *redisStorage) Get(ctx context.Context, key string) (i interface{}, err error) {
	cmd := r.Client.Get(ctx, key)
	err = cmd.Err()
	if err != nil {
		return
	}

	byteSlice, err := cmd.Bytes()
	if err != nil {
		return
	}

	err = json.Unmarshal(byteSlice, &i)
	return
}

// Save saves the token to the Redis storage.
func (r *redisStorage) Save(ctx context.Context, key string, i interface{}, ttl time.Duration) (err error) {
	if i == nil {
		err = errors.New("object cannot be null")
		return
	}

	byteSlice, err := json.Marshal(i)
	if err != nil {
		return
	}

	cmd := r.Client.SetEX(ctx, key, string(byteSlice), ttl)
	err = cmd.Err()
	return
}
