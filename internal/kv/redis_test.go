package kv

import (
	"context"
	"errors"
	"testing"

	redispkg "github.com/CRED-CLUB/propeller/pkg/broker/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Define a common error for testing
var errTestFailed = errors.New("test error")

// MockRedisClient is a mock implementation of the redispkg.IKVClient interface
type MockRedisClient struct {
	mock.Mock
}

// Ensure MockRedisClient implements redispkg.IKVClient
var _ redispkg.IKVClient = (*MockRedisClient)(nil)

func (m *MockRedisClient) HSet(ctx context.Context, key string, values ...interface{}) error {
	args := m.Called(ctx, key, values)
	return args.Error(0)
}

func (m *MockRedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(map[string]string), args.Error(1)
}

func (m *MockRedisClient) Delete(ctx context.Context, key string, fields ...string) error {
	args := m.Called(ctx, key, fields)
	return args.Error(0)
}

func TestNewRedis(t *testing.T) {
	mockClient := new(MockRedisClient)
	redis := NewRedis(mockClient)

	assert.NotNil(t, redis)
	assert.Implements(t, (*IKV)(nil), redis)

	// Type assertion to verify the internal client is set correctly
	redisImpl, ok := redis.(*Redis)
	assert.True(t, ok)
	assert.Equal(t, mockClient, redisImpl.redisClient)
}

func TestRedis_Store(t *testing.T) {
	ctx := context.Background()

	t.Run("successful store", func(t *testing.T) {
		mockClient := new(MockRedisClient)
		redis := NewRedis(mockClient)

		// The Redis implementation will call HSet with field and value as separate arguments
		mockClient.On("HSet", ctx, "test-key", []interface{}{"field1", "value1"}).Return(nil).Once()

		err := redis.Store(ctx, "test-key", "field1", "value1")

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("error on store", func(t *testing.T) {
		mockClient := new(MockRedisClient)
		redis := NewRedis(mockClient)

		mockClient.On("HSet", ctx, "test-key", []interface{}{"field1", "value1"}).Return(errTestFailed).Once()

		err := redis.Store(ctx, "test-key", "field1", "value1")

		assert.Error(t, err)
		assert.Equal(t, errTestFailed, err)
		mockClient.AssertExpectations(t)
	})
}

func TestRedis_Load(t *testing.T) {
	ctx := context.Background()

	t.Run("successful load with data", func(t *testing.T) {
		mockClient := new(MockRedisClient)
		redis := NewRedis(mockClient)

		expectedData := map[string]string{
			"field1": "value1",
			"field2": "value2",
		}

		mockClient.On("HGetAll", ctx, "test-key").Return(expectedData, nil).Once()

		result, err := redis.Load(ctx, "test-key")

		assert.NoError(t, err)
		assert.Equal(t, expectedData, result)
		mockClient.AssertExpectations(t)
	})

	t.Run("successful load with empty data", func(t *testing.T) {
		mockClient := new(MockRedisClient)
		redis := NewRedis(mockClient)

		expectedData := map[string]string{}

		mockClient.On("HGetAll", ctx, "test-key").Return(expectedData, nil).Once()

		result, err := redis.Load(ctx, "test-key")

		assert.NoError(t, err)
		assert.Empty(t, result)
		mockClient.AssertExpectations(t)
	})

	t.Run("error on load", func(t *testing.T) {
		mockClient := new(MockRedisClient)
		redis := NewRedis(mockClient)

		mockClient.On("HGetAll", ctx, "test-key").Return(map[string]string(nil), errTestFailed).Once()

		result, err := redis.Load(ctx, "test-key")

		assert.Error(t, err)
		assert.Equal(t, errTestFailed, err)
		assert.Nil(t, result)
		mockClient.AssertExpectations(t)
	})
}

func TestRedis_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("successful delete single field", func(t *testing.T) {
		mockClient := new(MockRedisClient)
		redis := NewRedis(mockClient)

		mockClient.On("Delete", ctx, "test-key", []string{"field1"}).Return(nil).Once()

		err := redis.Delete(ctx, "test-key", "field1")

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("successful delete multiple fields", func(t *testing.T) {
		mockClient := new(MockRedisClient)
		redis := NewRedis(mockClient)

		fields := []string{"field1", "field2"}
		mockClient.On("Delete", ctx, "test-key", fields).Return(nil).Once()

		err := redis.Delete(ctx, "test-key", fields...)

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("error on delete", func(t *testing.T) {
		mockClient := new(MockRedisClient)
		redis := NewRedis(mockClient)

		mockClient.On("Delete", ctx, "test-key", []string{"field1"}).Return(errTestFailed).Once()

		err := redis.Delete(ctx, "test-key", "field1")

		assert.Error(t, err)
		assert.Equal(t, errTestFailed, err)
		mockClient.AssertExpectations(t)
	})
}
