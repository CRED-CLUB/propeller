package redispkg

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/CRED-CLUB/propeller/pkg/broker"

	"github.com/CRED-CLUB/propeller/internal/perror"
	"github.com/CRED-CLUB/propeller/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// Streams provide redis streams implementation
type Streams struct {
	c             *Client
	cancelFuncMap map[string]context.CancelFunc
}

// NewStreams returns redis Streams
func NewStreams(c *Client) *Streams {
	return &Streams{c, make(map[string]context.CancelFunc)}
}

// Publish message to redis
func (ss Streams) Publish(ctx context.Context, request PublishRequest) error {
	v := map[string]interface{}{"data": request.Data}

	err := ss.c.client.XAdd(ctx, &redis.XAddArgs{
		Stream: fmt.Sprintf("%s-stream", request.Channel),
		MaxLen: 0,
		Values: v,
	}).Err()
	if err != nil {
		pErr := perror.Newf(perror.Internal, "error in redis stream publish %w", err)
		logger.Ctx(ctx).Error(pErr.Error())
		return err
	}

	return nil
}

// PublishBulk publishes messages in bulk
func (ss Streams) PublishBulk(ctx context.Context, publishRequests []PublishRequest) error {
	pipeline := ss.c.client.Pipeline()

	for _, request := range publishRequests {
		v := map[string]interface{}{"data": request.Data}
		err := pipeline.XAdd(ctx, &redis.XAddArgs{
			Stream: fmt.Sprintf("%s-stream", request.Channel),
			MaxLen: 0,
			Values: v,
		}).Err()

		if err != nil {
			pErr := perror.Newf(perror.Internal, "error in redis stream publish %w for request %+v", err, request)
			logger.Ctx(ctx).Error(pErr.Error())
		}

	}

	_, err := pipeline.Exec(ctx)
	if err != nil {
		pErr := perror.Newf(perror.Internal, "error in redis stream publish %w", err)
		logger.Ctx(ctx).Error(pErr.Error())
		return err
	}

	return nil
}

// Subscribe to a redis stream
func (ss Streams) Subscribe(ctx context.Context, channel ...string) broker.ISubscription {
	channels := make([]string, len(channel))
	for i, c := range channel {
		channels[i] = fmt.Sprintf("%s-stream", c)
	}
	StreamSubscription := StreamSubscription{
		Streams: ss,
		BaseSubscription: broker.BaseSubscription{
			TopicEventChan: make(chan broker.TopicEvent),
			Topics:         channels,
		},
	}
	go StreamSubscription.start(ctx, channels...)
	return StreamSubscription
}

// RemoveSubscription ...
func (ss Streams) RemoveSubscription(ctx context.Context, channel string, s broker.ISubscription) error {
	ss.cancelFuncMap[channel]()
	return nil
}

// AddSubscription ...
func (ss Streams) AddSubscription(ctx context.Context, channel string, s broker.ISubscription) error {
	// TODO: implement
	channels := []string{fmt.Sprintf("%s-stream", channel)}
	newCtx, cancelFunc := context.WithCancel(ctx)
	ss.cancelFuncMap[channel] = cancelFunc
	go s.(StreamSubscription).start(newCtx, channels...)
	return nil
}

// UnSubscribe ...
func (ss Streams) UnSubscribe(ctx context.Context, s broker.ISubscription) error {
	return nil
}

// StreamSubscription provides stream subscription
type StreamSubscription struct {
	Streams
	broker.BaseSubscription
}

func (st StreamSubscription) start(ctx context.Context, channels ...string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			resultStreams, err := st.Streams.c.client.XRead(ctx, &redis.XReadArgs{
				Streams: append(channels, "0"),
				Count:   1,
				Block:   60 * time.Second,
			}).Result()
			if err != nil {
				if strings.Contains(fmt.Sprint(err), "redis: nil") {
					continue
				}
				logger.Ctx(ctx).Errorw("error is", "err", err)
				return
			}
			for _, stream := range resultStreams[0].Messages {
				var msg []byte
				msg = []byte(stream.Values["data"].(string))
				te := broker.TopicEvent{
					Event: msg,
					Topic: resultStreams[0].Stream,
				}
				st.TopicEventChan <- te
				for _, chh := range channels {
					_, err := st.Streams.c.client.XDel(ctx, chh, stream.ID).Result()
					if err != nil {
						logger.Ctx(ctx).Infow("error deleting", "err", err.Error())
					}
				}
			}
		}
	}
}
