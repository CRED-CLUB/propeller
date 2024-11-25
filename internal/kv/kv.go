package kv

import (
	"context"

	"github.com/CRED-CLUB/propeller/internal/broker"
	"github.com/CRED-CLUB/propeller/internal/perror"
	"github.com/CRED-CLUB/propeller/pkg/logger"
)

// IKV interface
type IKV interface {
	Store(ctx context.Context, key string, field string, attrs string) error
	Load(ctx context.Context, key string) (map[string]string, error)
	Delete(ctx context.Context, key string, fields ...string) error
}

// New KV
func New(ctx context.Context, config broker.Config) (IKV, error) {
	switch config.Broker {
	case "nats":
		natsClient, err := broker.NewNATSClient(ctx, config)
		if err != nil {
			return nil, err
		}
		return NewNats(ctx, natsClient)
	case "redis":
		redisClient, err := broker.NewRedisClient(ctx, config)
		if err != nil {
			return nil, err
		}
		return NewRedis(redisClient), nil
	}
	pErr := perror.Newf(perror.Internal, "unknown kv type")
	logger.Ctx(ctx).Error(pErr.Error())
	return nil, pErr
}
