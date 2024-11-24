package pubsub

import (
	"context"
	"sync"

	"github.com/CRED-CLUB/propeller/internal/perror"
	"github.com/CRED-CLUB/propeller/internal/pubsub/subscription"
	"github.com/CRED-CLUB/propeller/pkg/broker"
	redispkg "github.com/CRED-CLUB/propeller/pkg/broker/redis"
	"github.com/CRED-CLUB/propeller/pkg/logger"
	"github.com/google/uuid"
)

// Redis is wrapper over redis pubSub
type Redis struct {
	BasePubSub
	redisClient redispkg.IRedis
}

// NewRedis returns redis
func NewRedis(client redispkg.IRedis) IPubSub {
	return &Redis{BasePubSub{&sync.Map{}}, client}
}

// Publish a event to a subject
func (r *Redis) Publish(ctx context.Context, request PublishRequest) error {
	publishReq := redispkg.PublishRequest{Channel: request.Channel, Data: request.Data}
	return r.redisClient.Publish(ctx, publishReq)
}

// PublishBulk publishes messages in bulk
func (r *Redis) PublishBulk(ctx context.Context, request []PublishRequest) error {
	var publishReqList []redispkg.PublishRequest

	logger.Ctx(ctx).Infow("publishing to topics")

	for _, v := range request {
		publishReq := redispkg.PublishRequest{Channel: v.Channel, Data: v.Data}
		publishReqList = append(publishReqList, publishReq)
	}
	return r.redisClient.PublishBulk(ctx, publishReqList)
}

// AsyncSubscribe to a subject
func (r *Redis) AsyncSubscribe(ctx context.Context, subject ...string) (*subscription.Subscription, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		pErr := perror.Newf(perror.Internal, "error in generating uuid %v", err)
		logger.Ctx(ctx).Error(pErr.Error())
		return nil, pErr
	}
	subs := &subscription.Subscription{
		TopicEventChan: make(chan broker.TopicEvent),
		ErrChan:        make(chan error),
		ID:             id,
	}
	pubs := r.redisClient.Subscribe(ctx, subject...)
	ch := pubs.GetTopicEventChan()
	go func() {
		for {
			select {
			case p := <-ch:
				subs.TopicEventChan <- p
			case <-ctx.Done():
				logger.Ctx(ctx).Debug("stopping subscriber loop")
				return
			}
		}
	}()
	r.BasePubSub.Store(ctx, id.String(), pubs)
	return subs, nil
}

// AddSubscription ...
func (r *Redis) AddSubscription(ctx context.Context, subject string, subs *subscription.Subscription) error {
	v, err := r.BasePubSub.Load(ctx, subs.ID.String())
	if err != nil {
		return err
	}
	var rs broker.ISubscription
	switch v.(type) {
	case broker.ISubscription:
		rs = v.(broker.ISubscription)
	default:
		pErr := perror.Newf(perror.Internal, "type not defined")
		logger.Ctx(ctx).Error(pErr)
		return pErr
	}
	err = r.redisClient.AddSubscription(ctx, subject, rs)
	if err != nil {
		return err
	}
	return nil
}

// RemoveSubscription ...
func (r *Redis) RemoveSubscription(ctx context.Context, subject string, subs *subscription.Subscription) error {
	v, err := r.BasePubSub.Load(ctx, subs.ID.String())
	if err != nil {
		return err
	}
	var rs broker.ISubscription
	switch v.(type) {
	case broker.ISubscription:
		rs = v.(broker.ISubscription)
	default:
		pErr := perror.Newf(perror.Internal, "type not defined")
		logger.Ctx(ctx).Error(pErr)
		return pErr
	}
	err = r.redisClient.RemoveSubscription(ctx, subject, rs)
	if err != nil {
		return err
	}
	return nil
}

// Unsubscribe a subject
func (r *Redis) Unsubscribe(ctx context.Context, subs *subscription.Subscription) error {
	v, err := r.BasePubSub.LoadAndDelete(ctx, subs.ID.String())
	if err != nil {
		return err
	}
	var rs broker.ISubscription
	switch v.(type) {
	case broker.ISubscription:
		rs = v.(broker.ISubscription)
	default:
		pErr := perror.Newf(perror.Internal, "type not defined")
		logger.Ctx(ctx).Error(pErr)
		return pErr
	}
	err = r.redisClient.UnSubscribe(ctx, rs)
	if err != nil {
		pErr := perror.Newf(perror.Internal, "unable to unsubscribe %v", err)
		logger.Ctx(ctx).Errorw(pErr.Error(), "subscription", subs.ID)
		return pErr
	}
	return nil
}
