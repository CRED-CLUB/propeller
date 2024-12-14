package pubsub

import (
	"context"
	"sync"
	"testing"

	"github.com/CRED-CLUB/propeller/pkg/broker"
	natspkg "github.com/CRED-CLUB/propeller/pkg/broker/nats"
	"github.com/CRED-CLUB/propeller/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockNatsClient struct {
	mock.Mock
}

func (m *mockNatsClient) Publish(ctx context.Context, req natspkg.PublishRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *mockNatsClient) Subscribe(ctx context.Context, subject string) (broker.ISubscription, error) {
	args := m.Called(ctx, subject)
	return args.Get(0).(broker.ISubscription), args.Error(1)
}

func (m *mockNatsClient) UnSubscribe(ctx context.Context, subscription broker.ISubscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

type mockSubscription struct {
	mock.Mock
}

func (m *mockSubscription) GetTopicEventChan() chan broker.TopicEvent {
	args := m.Called()
	return args.Get(0).(chan broker.TopicEvent)
}

func (m *mockSubscription) GetTopics() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func TestNewNats(t *testing.T) {
	mockClient := &mockNatsClient{}
	nats := NewNats(mockClient)
	assert.NotNil(t, nats)
}

func TestNats_Publish(t *testing.T) {
	ctx := context.Background()
	mockClient := &mockNatsClient{}
	nats := NewNats(mockClient).(*Nats)

	req := PublishRequest{
		Channel: "test-channel",
		Data:    []byte("test-data"),
	}

	mockClient.On("Publish", ctx, natspkg.PublishRequest{
		Channel: req.Channel,
		Data:    req.Data,
	}).Return(nil)

	err := nats.Publish(ctx, req)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestNats_PublishBulk(t *testing.T) {
	ctx := context.Background()
	mockClient := &mockNatsClient{}
	nats := NewNats(mockClient).(*Nats)

	requests := []PublishRequest{
		{Channel: "channel1", Data: []byte("data1")},
		{Channel: "channel2", Data: []byte("data2")},
	}

	for _, req := range requests {
		mockClient.On("Publish", ctx, natspkg.PublishRequest{
			Channel: req.Channel,
			Data:    req.Data,
		}).Return(nil)
	}

	err := nats.PublishBulk(ctx, requests)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestNats_AsyncSubscribe(t *testing.T) {
	ctx := context.Background()
	mockClient := &mockNatsClient{}
	mockSub := &mockSubscription{}
	nats := NewNats(mockClient).(*Nats)

	eventChan := make(chan broker.TopicEvent)
	mockSub.On("GetTopicEventChan").Return(eventChan)
	mockClient.On("Subscribe", ctx, "test-subject").Return(mockSub, nil)

	sub, err := nats.AsyncSubscribe(ctx, "test-subject")
	assert.NoError(t, err)
	assert.NotNil(t, sub)
	mockClient.AssertExpectations(t)
	mockSub.AssertExpectations(t)
}

func TestNats_Unsubscribe(t *testing.T) {
	ctx := context.Background()
	mockClient := &mockNatsClient{}
	mockSub := &mockSubscription{}
	nats := NewNats(mockClient).(*Nats)

	nats.BasePubSub = BasePubSub{ChannelSubscriptionMap: &sync.Map{}}
	nats.BasePubSub.Store(ctx, "test-id", mockSub)

	mockClient.On("UnSubscribe", ctx, mockSub).Return(nil)

	err := nats.removeSubscription(ctx, mockSub)
	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestNats_RemoveSubscription(t *testing.T) {
	// Initialize logger for tests
	serviceKV := map[string]interface{}{
		"serviceName":   "test-service",
		"gitCommitHash": "test-hash",
	}
	_, err := logger.NewLogger("dev", serviceKV, nil)
	assert.NoError(t, err)

	ctx := context.Background()
	mockClient := &mockNatsClient{}
	mockSub := &mockSubscription{}
	nats := NewNats(mockClient).(*Nats)

	t.Run("successful removal", func(t *testing.T) {
		// Setup
		subject := "test-subject"
		nats.BasePubSub = BasePubSub{ChannelSubscriptionMap: &sync.Map{}}
		nats.BasePubSub.Store(ctx, subject, mockSub)

		// Set expectations
		mockClient.On("UnSubscribe", ctx, mockSub).Return(nil).Once()

		// Execute
		err := nats.RemoveSubscription(ctx, subject, nil)

		// Assert
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)

		// Verify subscription was removed
		_, err = nats.BasePubSub.Load(ctx, subject)
		assert.Error(t, err)
	})

	t.Run("subscription not found", func(t *testing.T) {
		// Setup
		subject := "non-existent-subject"
		nats.BasePubSub = BasePubSub{ChannelSubscriptionMap: &sync.Map{}}

		// Execute
		err := nats.RemoveSubscription(ctx, subject, nil)

		// Assert
		assert.Error(t, err)
		mockClient.AssertNotCalled(t, "UnSubscribe")
	})

	t.Run("unsubscribe error", func(t *testing.T) {
		// Setup
		subject := "test-subject"
		nats.BasePubSub = BasePubSub{ChannelSubscriptionMap: &sync.Map{}}
		nats.BasePubSub.Store(ctx, subject, mockSub)

		// Set expectations
		mockClient.On("UnSubscribe", ctx, mockSub).Return(assert.AnError).Once()

		// Execute
		err := nats.RemoveSubscription(ctx, subject, nil)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)
		mockClient.AssertExpectations(t)
	})
}
