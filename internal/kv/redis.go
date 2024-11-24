package kv

import (
	"context"
	"encoding/json"

	redispkg "github.com/CRED-CLUB/propeller/pkg/broker/redis"

	"github.com/CRED-CLUB/propeller/internal/perror"
	"github.com/CRED-CLUB/propeller/pkg/logger"
)

// Redis ...
type Redis struct {
	redisClient *redispkg.Client
}

// NewRedis returns redis kv client
func NewRedis(client *redispkg.Client) IKV {
	return &Redis{client}
}

// Store key with values
func (r *Redis) Store(ctx context.Context, key string, field string, attrs map[string]string) error {
	jsonData, err := json.Marshal(attrs)
	if err != nil {
		pErr := perror.Newf(perror.Internal, "error in json marshalling %v", err)
		logger.Ctx(ctx).Error(pErr.Error())
		return pErr
	}
	err = r.redisClient.HSet(ctx, key, field, jsonData)
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
