package natspkg

import (
	"context"
	"time"

	"github.com/CRED-CLUB/propeller/internal/perror"
	"github.com/CRED-CLUB/propeller/pkg/logger"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

// PubSub provides nats pubsub implementation
type PubSub struct {
	c *Client
}

// wait period for the embedded server to start
const waitPeriod = 5

// NewPubSub returns nats PubSub
func NewPubSub(c *Client) *PubSub {
	return &PubSub{c}
}

// Publish data to the subject
func (s PubSub) Publish(ctx context.Context, request PublishRequest) error {
	err := s.c.conn.Publish(request.Channel, request.Data)
	if err != nil {
		pErr := perror.Newf(perror.Internal, "unable to publish to nats %w", err)
		logger.Ctx(ctx).Error(pErr.Error())
		return pErr
	}
	return nil
}

// UnSubscribe a subscriber
func (s PubSub) UnSubscribe(ctx context.Context, subscription ISubscription) error {
	switch subscription := subscription.(type) {
	case *PubSubSubscription:
		err := subscription.subs.Unsubscribe()
		if err != nil {
			pErr := perror.Newf(perror.Internal, "unable to unsubscribe %w", err)
			logger.Ctx(ctx).Error(pErr.Error())
			return pErr
		}
		return nil
	default:
		pErr := perror.Newf(perror.Internal, "invalid subscription type: %T", subscription)
		logger.Ctx(ctx).Error(pErr.Error())
		return pErr
	}
}

// Subscribe to a subject
func (s PubSub) Subscribe(ctx context.Context, subject string) (ISubscription, error) {
	ps := &PubSubSubscription{
		baseSubscription: baseSubscription{
			dataChan: make(chan []byte),
			topics:   subject,
		},
		subs: nil,
	}
	subs, err := s.c.conn.Subscribe(subject, ps.handlerFunc)
	if err != nil {
		pErr := perror.Newf(perror.Internal, "error in creating subscription %w", err)
		logger.Ctx(ctx).Error(pErr.Error())
		return nil, pErr
	}
	ps.subs = subs
	return ps, nil
}

// NewEmbeddedServer start an embedded NATS server for stage/testing
func NewEmbeddedServer(ctx context.Context) (string, error) {
	opts := &server.Options{JetStream: true}
	ns, err := server.NewServer(opts)
	if err != nil {
		pErr := perror.Newf(perror.Internal, "unable to init nats server %w", err)
		logger.Ctx(ctx).Error(pErr.Error())
		return "", pErr
	}

	go ns.Start()
	if !ns.ReadyForConnections(waitPeriod * time.Second) {
		pErr := perror.Newf(perror.Internal, "not ready for connection after %d secs", waitPeriod)
		logger.Ctx(ctx).Error(pErr.Error())
		return "", pErr
	}
	return ns.ClientURL(), nil
}

// PubSubSubscription ...
type PubSubSubscription struct {
	baseSubscription
	subs *nats.Subscription
}

func (ps PubSubSubscription) handlerFunc(msg *nats.Msg) {
	ps.dataChan <- msg.Data
}
