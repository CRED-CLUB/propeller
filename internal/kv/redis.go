package kv

import (
	"context"

	redispkg "github.com/CRED-CLUB/propeller/pkg/broker/redis"
)

// Redis implements IKV interface using Redis
type Redis struct {
	redisClient redispkg.IKVClient
}

// NewRedis returns redis kv client
func NewRedis(client redispkg.IKVClient) IKV {
	return &Redis{client}
}

// Store key with values
func (r *Redis) Store(ctx context.Context, key string, field string, attrs string) error {
	err := r.redisClient.HSet(ctx, key, field, attrs)
	if err != nil {
		return err
	}
	return nil
}

// Load values for a key
func (r *Redis) Load(ctx context.Context, key string) (map[string]string, error) {
	return r.redisClient.HGetAll(ctx, key)
}

// Delete values for a key
func (r *Redis) Delete(ctx context.Context, key string, fields ...string) error {
	return r.redisClient.Delete(ctx, key, fields...)
}
