package pubsub

import (
	"context"
	"sync"

	"github.com/CRED-CLUB/propeller/internal/broker"
	"github.com/CRED-CLUB/propeller/internal/pubsub/subscription"

	"github.com/CRED-CLUB/propeller/internal/perror"
	"github.com/CRED-CLUB/propeller/pkg/logger"
	natspkg "github.com/CRED-CLUB/propeller/pkg/nats"
	redispkg "github.com/CRED-CLUB/propeller/pkg/redis"
)

// IPubSub is pubsub interface
type IPubSub interface {
	Publish(ctx context.Context, publishRequest PublishRequest) error
	PublishBulk(ctx context.Context, publishRequest []PublishRequest) error
	AsyncSubscribe(ctx context.Context, subject ...string) (*subscription.Subscription, error)
	Unsubscribe(ctx context.Context, subs *subscription.Subscription) error
	AddSubscription(ctx context.Context, subject string, subs *subscription.Subscription) error
	RemoveSubscription(ctx context.Context, subject string, subs *subscription.Subscription) error
}

// New returns a new pubsub type
func New(ctx context.Context, config broker.Config) (IPubSub, error) {
	switch config.Broker {
	case "nats":
		var psType natspkg.INats
		natsClient, err := broker.NewNATSClient(ctx, config)
		if err != nil {
			return nil, err
		}
		switch config.Persistence {
		case true:
			psType, err = natspkg.NewJetStream(ctx, natsClient)
			if err != nil {
				return nil, err
			}
			logger.Ctx(ctx).Info("initialising NATS jetstream")
		case false:
			logger.Ctx(ctx).Info("initialising NATS pubsub")
			psType = natspkg.NewPubSub(natsClient)
		}
		return NewNats(psType), nil
	case "redis":
		var psType redispkg.IRedis
		redisClient, err := broker.NewRedisClient(ctx, config)
		if err != nil {
			return nil, err
		}
		switch config.Persistence {
		case true:
			logger.Ctx(ctx).Info("initialising redis streams")
			psType = redispkg.NewStreams(redisClient)
		case false:
			logger.Ctx(ctx).Info("initialising redis pubsub")
			psType = redispkg.NewPubSub(redisClient)
		}
		return NewRedis(psType), nil
	}
	pErr := perror.Newf(perror.Internal, "unknown pubsub type")
	logger.Ctx(ctx).Error(pErr.Error())
	return nil, pErr
}

// BasePubSub ...
type BasePubSub struct {
	ChannelSubscriptionMap *sync.Map
}

// Store the subscription
func (b BasePubSub) Store(ctx context.Context, id string, val interface{}) {
	b.ChannelSubscriptionMap.Store(id, val)
}

// LoadAndDelete the subscription
func (b BasePubSub) LoadAndDelete(ctx context.Context, id string) (interface{}, error) {
	v, ok := b.ChannelSubscriptionMap.LoadAndDelete(id)
	if !ok {
		pErr := perror.Newf(perror.NotFound, "unable to get the subscription")
		logger.Ctx(ctx).Error(pErr.Error())
		return nil, pErr
	}
	return v, nil
}

// Load the subscription
func (b BasePubSub) Load(ctx context.Context, id string) (interface{}, error) {
	v, ok := b.ChannelSubscriptionMap.Load(id)
	if !ok {
		pErr := perror.Newf(perror.NotFound, "unable to get the subscription")
		logger.Ctx(ctx).Error(pErr.Error())
		return nil, pErr
	}
	return v, nil
}
