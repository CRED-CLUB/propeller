package kv

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockNatsKVClient is a mock implementation of NatsKVClient
type MockNatsKVClient struct {
	mock.Mock
}

// Put mocks the Put method
func (m *MockNatsKVClient) Put(ctx context.Context, key string, value []byte) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

// Get mocks the Get method
func (m *MockNatsKVClient) Get(ctx context.Context, key string) ([]byte, error) {
	args := m.Called(ctx, key)
	return args.Get(0).([]byte), args.Error(1)
}

// Delete mocks the Delete method
func (m *MockNatsKVClient) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func TestNewNats(t *testing.T) {
	mockClient := new(MockNatsKVClient)
	natsKV, err := NewNats(mockClient)

	assert.NoError(t, err)
	assert.NotNil(t, natsKV)

	// Type assertion to access the internal kvClient field
	natsImpl, ok := natsKV.(*Nats)
	assert.True(t, ok)
	assert.Equal(t, mockClient, natsImpl.kvClient)
}

func TestNats_Store(t *testing.T) {
	ctx := context.Background()

	t.Run("successful store", func(t *testing.T) {
		mockClient := new(MockNatsKVClient)
		natsKV, err := NewNats(mockClient)
		assert.NoError(t, err)

		// Mock Get to return empty data (no existing values)
		mockClient.On("Get", ctx, "test-key").Return([]byte{}, nil).Once()

		// Mock Put to succeed
		mockClient.On("Put", ctx, "test-key", mock.Anything).Return(nil).Once()

		err = natsKV.Store(ctx, "test-key", "field1", "value1")

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("error on put", func(t *testing.T) {
		mockClient := new(MockNatsKVClient)
		natsKV, err := NewNats(mockClient)
		assert.NoError(t, err)

		// Mock Get to return empty data
		mockClient.On("Get", ctx, "test-key").Return([]byte{}, nil).Once()

		// Mock Put to fail
		mockClient.On("Put", ctx, "test-key", mock.Anything).Return(errors.New("put error")).Once()

		err = natsKV.Store(ctx, "test-key", "field1", "value1")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "put error")
		mockClient.AssertExpectations(t)
	})
}

func TestNats_Load(t *testing.T) {
	ctx := context.Background()

	t.Run("successful load with data", func(t *testing.T) {
		mockClient := new(MockNatsKVClient)
		natsKV, err := NewNats(mockClient)
		assert.NoError(t, err)

		// Create test data - a serialized map with gob encoding
		testMap := map[string]string{"field1": "value1"}
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		encErr := enc.Encode(testMap)
		assert.NoError(t, encErr)

		// Mock Get to return the serialized map
		mockClient.On("Get", ctx, "test-key").Return(buf.Bytes(), nil).Once()

		result, err := natsKV.Load(ctx, "test-key")

		assert.NoError(t, err)
		assert.Equal(t, testMap, result)
		mockClient.AssertExpectations(t)
	})

	t.Run("load with invalid data", func(t *testing.T) {
		mockClient := new(MockNatsKVClient)
		natsKV, err := NewNats(mockClient)
		assert.NoError(t, err)

		// Mock Get to return invalid data
		mockClient.On("Get", ctx, "test-key").Return([]byte("invalid data"), nil).Once()

		result, err := natsKV.Load(ctx, "test-key")

		assert.Error(t, err)
		assert.Nil(t, result)
		mockClient.AssertExpectations(t)
	})
}

func TestNats_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("delete single field", func(t *testing.T) {
		mockClient := new(MockNatsKVClient)
		natsKV, err := NewNats(mockClient)
		assert.NoError(t, err)

		// Create test data - a serialized map with gob encoding
		testMap := map[string]string{"field1": "value1", "field2": "value2"}
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		encErr := enc.Encode(testMap)
		assert.NoError(t, encErr)

		// Mock Get to return the serialized map
		mockClient.On("Get", ctx, "test-key").Return(buf.Bytes(), nil).Once()

		// Mock Delete to succeed
		mockClient.On("Delete", ctx, "test-key").Return(nil).Once()

		// Mock Get again for the Store operation
		mockClient.On("Get", ctx, "test-key").Return([]byte{}, nil).Once()

		// Mock Put for the remaining field
		mockClient.On("Put", ctx, "test-key", mock.Anything).Return(nil).Once()

		err = natsKV.Delete(ctx, "test-key", "field1")

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("delete error", func(t *testing.T) {
		mockClient := new(MockNatsKVClient)
		natsKV, err := NewNats(mockClient)
		assert.NoError(t, err)

		// Create test data - a serialized map with gob encoding
		testMap := map[string]string{"field1": "value1"}
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		encErr := enc.Encode(testMap)
		assert.NoError(t, encErr)

		// Mock Get to return the serialized map
		mockClient.On("Get", ctx, "test-key").Return(buf.Bytes(), nil).Once()

		// Mock Delete to fail
		mockClient.On("Delete", ctx, "test-key").Return(errors.New("delete error")).Once()

		err = natsKV.Delete(ctx, "test-key", "field1")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "delete error")
		mockClient.AssertExpectations(t)
	})
}
