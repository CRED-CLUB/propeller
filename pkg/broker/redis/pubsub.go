package redispkg

import (
	"context"

	"github.com/CRED-CLUB/propeller/pkg/broker"

	"github.com/CRED-CLUB/propeller/internal/perror"
	"github.com/CRED-CLUB/propeller/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// PubSub provides redis pubsub implementation
type PubSub struct {
	c *Client
}

// NewPubSub returns redis PubSub
func NewPubSub(c *Client) *PubSub {
	return &PubSub{c}
}

// Publish message to redis
func (p PubSub) Publish(ctx context.Context, request PublishRequest) error {
	_, err := p.c.client.Publish(ctx, request.Channel, request.Data).Result()
	if err != nil {
		pErr := perror.Newf(perror.Internal, "error in publishing %w", err)
		logger.Ctx(ctx).Error(pErr.Error())
		return pErr
	}
	return nil
}

// PublishBulk publishes messages in bulk
func (p PubSub) PublishBulk(ctx context.Context, publishRequest []PublishRequest) error {
	pipe := p.c.client.TxPipeline()

	for _, v := range publishRequest {
		err := pipe.Publish(ctx, v.Channel, v.Data).Err()
		if err != nil {
			pErr := perror.Newf(perror.Internal, "error in redis publishing %w for request %+v", err, v)
			logger.Ctx(ctx).Error(pErr.Error())
		}
	}

	_, err := pipe.Exec(ctx)

	if err != nil {
		pErr := perror.Newf(perror.Internal, "error in redis publishing %w", err)
		logger.Ctx(ctx).Error(pErr.Error())
		return pErr
	}
	return nil
}

// Subscribe to a redis pubsub channel
func (p PubSub) Subscribe(ctx context.Context, channel ...string) broker.ISubscription {
	s := p.c.client.Subscribe(ctx, channel...)
	pubSubSubscription := PubSubSubscription{
		BaseSubscription: broker.BaseSubscription{
			TopicEventChan: make(chan broker.TopicEvent),
			Topics:         channel,
		},
		subs: s,
	}
	go pubSubSubscription.start(ctx)
	return pubSubSubscription
}

// AddSubscription to a redis pubsub channel
func (p PubSub) AddSubscription(ctx context.Context, channel string, s broker.ISubscription) error {
	PubSubSubscription := s.(PubSubSubscription)
	err := PubSubSubscription.subs.Subscribe(ctx, channel)
	if err != nil {
		pErr := perror.Newf(perror.Internal, "unable to add subscription of channel %s with error %w", channel, err)
		logger.Ctx(ctx).Error(pErr.Error())
		return pErr
	}
	return nil
}

// RemoveSubscription removes a subscription
func (p PubSub) RemoveSubscription(ctx context.Context, channel string, s broker.ISubscription) error {
	PubSubSubscription := s.(PubSubSubscription)
	err := PubSubSubscription.subs.Unsubscribe(ctx, channel)
	if err != nil {
		pErr := perror.Newf(perror.Internal, "unable to unsubscribe subscription %s with error %w", channel, err)
		logger.Ctx(ctx).Error(pErr.Error())
		return pErr
	}
	return nil
}

// UnSubscribe from all channels
func (p PubSub) UnSubscribe(ctx context.Context, s broker.ISubscription) error {
	PubSubSubscription := s.(PubSubSubscription)
	for _, v := range PubSubSubscription.Topics {
		err := p.RemoveSubscription(ctx, v, PubSubSubscription)
		if err != nil {
			return err
		}
	}
	err := PubSubSubscription.subs.Close()
	if err != nil {
		pErr := perror.Newf(perror.Internal, "unable to close subscription %w", err)
		logger.Ctx(ctx).Error(pErr.Error())
		return pErr
	}
	return nil
}

// PubSubSubscription provides pubsub subscription
type PubSubSubscription struct {
	broker.BaseSubscription
	subs *redis.PubSub
}

func (p PubSubSubscription) start(ctx context.Context) {
	for {
		select {
		case msg := <-p.subs.Channel():
			te := broker.TopicEvent{
				Event: []byte(msg.Payload),
				Topic: msg.Channel,
			}
			p.TopicEventChan <- te
		case <-ctx.Done():
			logger.Ctx(ctx).Debug("stopping redis subscription")
			return
		}
	}
}
