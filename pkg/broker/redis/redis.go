package redispkg

import (
	"context"
	"crypto/tls"

	"github.com/CRED-CLUB/propeller/pkg/broker"

	"github.com/CRED-CLUB/propeller/internal/perror"
	"github.com/CRED-CLUB/propeller/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// IRedis is an interface over redis pubusb and streams
type IRedis interface {
	Publish(ctx context.Context, publishRequest PublishRequest) error
	PublishBulk(ctx context.Context, publishRequest []PublishRequest) error
	Subscribe(ctx context.Context, channel ...string) broker.ISubscription
	UnSubscribe(ctx context.Context, s broker.ISubscription) error
	AddSubscription(ctx context.Context, channel string, s broker.ISubscription) error
	RemoveSubscription(ctx context.Context, channel string, s broker.ISubscription) error
}

// Client holds redis client
type Client struct {
	client redis.UniversalClient
}

// NewClient returns a new redis client
func NewClient(config Config) *Client {
	var tlsConfig *tls.Config = nil
	if config.TLSEnabled {
		tlsConfig = &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: true,
		}
	}
	var client redis.UniversalClient
	if config.ClusterModeEnabled == true {
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:                 []string{config.Address},
			ClientName:            "",
			NewClient:             nil,
			MaxRedirects:          0,
			ReadOnly:              false,
			RouteByLatency:        false,
			RouteRandomly:         false,
			ClusterSlots:          nil,
			Dialer:                nil,
			OnConnect:             nil,
			Username:              "",
			Password:              config.Password,
			MaxRetries:            0,
			MinRetryBackoff:       0,
			MaxRetryBackoff:       0,
			DialTimeout:           0,
			ReadTimeout:           0,
			WriteTimeout:          0,
			ContextTimeoutEnabled: false,
			PoolFIFO:              false,
			PoolSize:              0,
			PoolTimeout:           0,
			MinIdleConns:          0,
			MaxIdleConns:          0,
			ConnMaxIdleTime:       0,
			ConnMaxLifetime:       0,
			TLSConfig:             tlsConfig,
		})
	} else {
		client = redis.NewClient(&redis.Options{
			Network:               "",
			Addr:                  config.Address,
			ClientName:            "",
			Dialer:                nil,
			OnConnect:             nil,
			Username:              "",
			Password:              "",
			CredentialsProvider:   nil,
			DB:                    0,
			MaxRetries:            0,
			MinRetryBackoff:       0,
			MaxRetryBackoff:       0,
			DialTimeout:           0,
			ReadTimeout:           0,
			WriteTimeout:          0,
			ContextTimeoutEnabled: false,
			PoolFIFO:              false,
			PoolSize:              0,
			PoolTimeout:           0,
			MinIdleConns:          0,
			MaxIdleConns:          0,
			ConnMaxIdleTime:       0,
			ConnMaxLifetime:       0,
			TLSConfig:             nil,
			Limiter:               nil,
		})
	}

	return &Client{client: client}
}

// HSet set key with values
func (c *Client) HSet(ctx context.Context, key string, values ...interface{}) error {
	err := c.client.HSet(ctx, key, values...).Err()
	if err != nil {
		pErr := perror.Newf(perror.Internal, "error in redis hash set key:%v value:%v %+v", key, values, err)
		logger.Ctx(ctx).Error(pErr.Error())
		return err
	}
	return nil
}

// HGetAll returns all values map for a key
func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	v, err := c.client.HGetAll(ctx, key).Result()
	if err != nil {
		pErr := perror.Newf(perror.Internal, "error in redis hash get all key:%v %+v", key, err)
		logger.Ctx(ctx).Error(pErr.Error())
		return nil, err
	}
	return v, nil
}

// Delete a field for a key
func (c *Client) Delete(ctx context.Context, key string, fields ...string) error {
	err := c.client.HDel(ctx, key, fields...).Err()
	if err != nil {
		pErr := perror.Newf(perror.Internal, "error in redis delete for key:%v %+v", key, err)
		logger.Ctx(ctx).Error(pErr.Error())
		return err
	}
	return nil
}
