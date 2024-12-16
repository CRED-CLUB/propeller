package redispkg

import (
	"context"
	"testing"

	"github.com/CRED-CLUB/propeller/pkg/broker"
	"github.com/CRED-CLUB/propeller/pkg/logger"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRedisClient is a mock implementation of redis.UniversalClient
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd {
	args := m.Called(ctx, channel, message)
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	args := m.Called(ctx, channels)
	return args.Get(0).(*redis.PubSub)
}

func (m *MockRedisClient) TxPipeline() redis.Pipeliner {
	args := m.Called()
	return args.Get(0).(redis.Pipeliner)
}

func (m *MockRedisClient) AddHook(hook redis.Hook) {
	m.Called(hook)
}

func (m *MockRedisClient) Append(ctx context.Context, key, value string) *redis.IntCmd {
	args := m.Called(ctx, key, value)
	return args.Get(0).(*redis.IntCmd)
}

// MockPipeliner is a mock implementation of redis.Pipeliner
type MockPipeliner struct {
	mock.Mock
}

func (m *MockPipeliner) Exec(ctx context.Context) ([]redis.Cmder, error) {
	args := m.Called(ctx)
	return args.Get(0).([]redis.Cmder), args.Error(1)
}

func (m *MockPipeliner) Publish(ctx context.Context, channel string, message interface{}) *redis.IntCmd {
	args := m.Called(ctx, channel, message)
	return args.Get(0).(*redis.IntCmd)
}

func TestPubSub_Publish(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer client.Close()

	ps := &PubSub{
		c: &Client{client: client},
	}

	// Test cases using real Redis commands with miniredis
	err = ps.Publish(context.Background(), PublishRequest{
		Channel: "test-channel",
		Data:    []byte("test-data"),
	})
	assert.NoError(t, err)
}

func TestPubSub_PublishBulk(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer client.Close()

	ps := &PubSub{
		c: &Client{client: client},
	}

	requests := []PublishRequest{
		{Channel: "channel1", Data: []byte("data1")},
		{Channel: "channel2", Data: []byte("data2")},
	}

	err = ps.PublishBulk(context.Background(), requests)
	assert.NoError(t, err)
}

func TestPubSub_Subscribe(t *testing.T) {
	// Initialize logger for tests
	serviceKV := map[string]interface{}{
		"serviceName":   "test-service",
		"gitCommitHash": "test-hash",
	}
	_, err := logger.NewLogger("dev", serviceKV, nil)

	assert.NoError(t, err)
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer client.Close()

	ps := &PubSub{
		c: &Client{client: client},
	}
	ctx, cancel := context.WithCancel(context.Background())
	subscription := ps.Subscribe(ctx, "test-channel")
	assert.NotNil(t, subscription)
	assert.IsType(t, PubSubSubscription{}, subscription)
	cancel()
}

func TestPubSub_RemoveSubscription(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer client.Close()

	ps := &PubSub{
		c: &Client{client: client},
	}

	subscription := PubSubSubscription{
		BaseSubscription: broker.BaseSubscription{
			TopicEventChan: make(chan broker.TopicEvent),
			Topics:         []string{"test-channel"},
		},
		subs: client.Subscribe(context.Background()),
	}

	err = ps.RemoveSubscription(context.Background(), "test-channel", subscription)
	assert.NoError(t, err)
}

func TestPubSub_UnSubscribe(t *testing.T) {
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer client.Close()

	ps := &PubSub{
		c: &Client{client: client},
	}

	subscription := PubSubSubscription{
		BaseSubscription: broker.BaseSubscription{
			TopicEventChan: make(chan broker.TopicEvent),
			Topics:         []string{"test-channel"},
		},
		subs: client.Subscribe(context.Background()),
	}

	err = ps.UnSubscribe(context.Background(), subscription)
	assert.NoError(t, err)
}
