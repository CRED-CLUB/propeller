package natspkg

import (
	"context"
	"sync"

	"github.com/CRED-CLUB/propeller/pkg/broker"

	"github.com/CRED-CLUB/propeller/internal/perror"
	"github.com/CRED-CLUB/propeller/pkg/logger"
	"github.com/nats-io/nats.go/jetstream"
)

// JetStream ...
type JetStream struct {
	c         *Client
	js        jetstream.JetStream
	streamMap sync.Map
	ctx       context.Context
}

// NewJetStream returns nats NewJetStream
func NewJetStream(ctx context.Context, c *Client) (*JetStream, error) {
	js, err := jetstream.New(c.conn)
	if err != nil {
		pErr := perror.Newf(perror.Internal, "unable to create JetStream %s", err)
		logger.Ctx(ctx).Error(pErr.Error())
		return nil, pErr
	}
	return &JetStream{c, js, sync.Map{}, ctx}, nil
}

// Publish data to the subject
func (j *JetStream) Publish(ctx context.Context, request PublishRequest) error {
	streamConfig := jetstream.StreamConfig{
		Name:     request.Channel,
		Subjects: []string{request.Channel},
	}
	_, ok := j.streamMap.Load(request.Channel)
	if !ok {
		stream, err := j.js.CreateStream(ctx, streamConfig)
		if err != nil {
			pErr := perror.Newf(perror.Internal, "unable to create JetStream stream: %s", err)
			logger.Ctx(ctx).Error(pErr.Error())
			return pErr
		}
		j.streamMap.Store(request.Channel, stream)
	}

	_, err := j.js.Publish(ctx, request.Channel, request.Data)
	if err != nil {
		pErr := perror.Newf(perror.Internal, "unable to publish JetStream stream: %s", err)
		logger.Ctx(ctx).Error(pErr.Error())
		return pErr
	}
	return nil
}

// Subscribe to jetstream topic
func (j *JetStream) Subscribe(ctx context.Context, channel string) (broker.ISubscription, error) {
	js := &JetStreamSubscription{
		BaseSubscription: broker.BaseSubscription{
			make(chan broker.TopicEvent),
			[]string{channel},
		},
		JetStream: *j,
	}
	_, ok := j.streamMap.Load(channel)
	if !ok {
		streamConfig := jetstream.StreamConfig{
			Name:     channel,
			Subjects: []string{channel},
		}
		stream, err := j.js.CreateStream(ctx, streamConfig)
		if err != nil {
			pErr := perror.Newf(perror.Internal, "unable to create JetStream stream: %s", err)
			logger.Ctx(ctx).Error(pErr.Error())
			return nil, pErr
		}
		j.streamMap.Store(channel, stream)
	}
	consumerConfig := jetstream.ConsumerConfig{
		Durable:   channel,
		AckPolicy: jetstream.AckExplicitPolicy,
	}
	consumer, err := j.js.CreateConsumer(ctx, channel, consumerConfig)
	if err != nil {
		pErr := perror.Newf(perror.Internal, "unable to create JetStream consumer: %s", err)
		logger.Ctx(ctx).Error(pErr.Error())
		return nil, pErr
	}
	c, err := consumer.Consume(js.jetStreamMessageHandler)
	if err != nil {
		pErr := perror.Newf(perror.Internal, "unable to consume from JetStream: %s", err)
		logger.Ctx(ctx).Error(pErr.Error())
		return nil, pErr
	}
	js.consumeContext = c
	return js, nil
}

// UnSubscribe from jetstream topic
func (j *JetStream) UnSubscribe(ctx context.Context, s broker.ISubscription) error {
	switch t := s.(type) {
	case *JetStreamSubscription:
		t.consumeContext.Stop()
	default:
		pErr := perror.Newf(perror.Internal, "invalid subscription type: %T", t)
		logger.Ctx(ctx).Error(pErr.Error())
		return pErr
	}
	return nil
}

// JetStreamSubscription ...
type JetStreamSubscription struct {
	broker.BaseSubscription
	JetStream
	consumeContext jetstream.ConsumeContext
}

func (j *JetStreamSubscription) jetStreamMessageHandler(msg jetstream.Msg) {
	te := broker.TopicEvent{
		Event: msg.Data(),
		Topic: msg.Subject(),
	}
	j.TopicEventChan <- te
	err := msg.Ack()
	if err != nil {
		pErr := perror.Newf(perror.Internal, "unable to ack JetStream msg: %s", err)
		logger.Ctx(j.ctx).Error(pErr.Error())
	}
}
