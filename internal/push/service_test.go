package push

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/CRED-CLUB/propeller/internal/config"
	kv "github.com/CRED-CLUB/propeller/internal/kv" // Required for interface
	"github.com/CRED-CLUB/propeller/internal/perror"
	"github.com/CRED-CLUB/propeller/internal/pubsub"
	"github.com/CRED-CLUB/propeller/internal/pubsub/subscription"
	"github.com/CRED-CLUB/propeller/pkg/broker"
	"github.com/CRED-CLUB/propeller/pkg/logger"
	pushv1 "github.com/CRED-CLUB/propeller/rpc/push/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// MockPubSub is a mock implementation of pubsub.IPubSub
type MockPubSub struct {
	mock.Mock
}

func (m *MockPubSub) Publish(ctx context.Context, req pubsub.PublishRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockPubSub) PublishBulk(ctx context.Context, reqs []pubsub.PublishRequest) error {
	args := m.Called(ctx, reqs)
	return args.Error(0)
}

func (m *MockPubSub) AsyncSubscribe(ctx context.Context, channels ...string) (*subscription.Subscription, error) {
	args := m.Called(ctx, channels)
	return args.Get(0).(*subscription.Subscription), args.Error(1)
}

func (m *MockPubSub) AddSubscription(ctx context.Context, channel string, subscription *subscription.Subscription) error {
	args := m.Called(ctx, channel, subscription)
	return args.Error(0)
}

func (m *MockPubSub) RemoveSubscription(ctx context.Context, channel string, subscription *subscription.Subscription) error {
	args := m.Called(ctx, channel, subscription)
	return args.Error(0)
}

func (m *MockPubSub) Unsubscribe(ctx context.Context, subscription *subscription.Subscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

// MockKV is a mock implementation of kv.IKV interface
type MockKV struct {
	mock.Mock
}

func (m *MockKV) Store(ctx context.Context, key, field, value string) error {
	args := m.Called(ctx, key, field, value)
	return args.Error(0)
}

func (m *MockKV) Load(ctx context.Context, key string) (map[string]string, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(map[string]string), args.Error(1)
}

func (m *MockKV) Delete(ctx context.Context, key string, fields ...string) error {
	args := m.Called(ctx, key, fields)
	return args.Error(0)
}

// setupTest creates a new Service instance with mocked dependencies for testing
func setupTest(t *testing.T) (*Service, *MockPubSub, *MockKV, context.Context) {
	mockPS := new(MockPubSub)
	mockKV := new(MockKV)
	var _ kv.IKV = mockKV // Verify MockKV implements kv.IKV interface
	cfg := config.Config{
		EnableDeviceSupport: true, // Enable device support for tests
	}
	svc := NewService(mockPS, mockKV, cfg)
	ctx := context.Background()
	logger.NewLogger("dev", nil, nil) // Initialize logger for tests with development config
	return svc, mockPS, mockKV, ctx
}

func TestPublishToTopic(t *testing.T) {
	svc, mockPS, _, ctx := setupTest(t)

	tests := []struct {
		name        string
		req         SendEventToTopicRequest
		setupMock   func()
		expectedErr error
	}{
		{
			name: "successful publish",
			req: SendEventToTopicRequest{
				Topic:     "test-topic",
				EventName: "test-event",
				Event:     []byte("test-data"),
			},
			setupMock: func() {
				mockPS.On("Publish", ctx, mock.MatchedBy(func(req pubsub.PublishRequest) bool {
					return req.Channel == "test-topic" && string(req.Data) == "test-data"
				})).Return(nil).Once()
			},
			expectedErr: nil,
		},
		{
			name: "empty topic",
			req: SendEventToTopicRequest{
				Topic:     "",
				EventName: "test-event",
				Event:     []byte("test-data"),
			},
			setupMock: func() {
				// No mock setup needed as validation should fail
			},
			expectedErr: perror.New(perror.InvalidArgument, "Topic is empty"),
		},
		{
			name: "nil event",
			req: SendEventToTopicRequest{
				Topic:     "test-topic",
				EventName: "test-event",
				Event:     nil,
			},
			setupMock: func() {
				// No mock setup needed as validation should fail
			},
			expectedErr: perror.New(perror.InvalidArgument, "Event is empty"),
		},
		{
			name: "empty event name",
			req: SendEventToTopicRequest{
				Topic:     "test-topic",
				EventName: "",
				Event:     []byte("test-data"),
			},
			setupMock: func() {
				// No mock setup needed as validation should fail
			},
			expectedErr: perror.New(perror.InvalidArgument, "Event name is empty"),
		},
		{
			name: "publish error",
			req: SendEventToTopicRequest{
				Topic:     "test-topic",
				EventName: "test-event",
				Event:     []byte("test-data"),
			},
			setupMock: func() {
				mockPS.On("Publish", ctx, mock.Anything).Return(errors.New("publish failed")).Once()
			},
			expectedErr: errors.New("publish failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := svc.PublishToTopic(ctx, tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				if perr, ok := tt.expectedErr.(*perror.PError); ok {
					assert.Equal(t, perr.Code(), perror.Code(err.(*perror.PError).Code()))
					assert.Contains(t, err.Error(), perr.Error())
				} else {
					assert.Equal(t, tt.expectedErr.Error(), err.Error())
				}
			} else {
				assert.NoError(t, err)
			}

			mockPS.AssertExpectations(t)
		})
	}
}

func TestPublishToClientWithDevice(t *testing.T) {
	svc, mockPS, _, ctx := setupTest(t)

	tests := []struct {
		name    string
		req     SendEventToClientDeviceChannelRequest
		mockErr error
		wantErr bool
	}{
		{
			name: "valid request",
			req: SendEventToClientDeviceChannelRequest{
				clientID:  "test-client",
				deviceID:  "test-device",
				eventName: "test-event",
				event:     []byte("test-data"),
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name: "empty client ID",
			req: SendEventToClientDeviceChannelRequest{
				clientID:  "",
				deviceID:  "test-device",
				eventName: "test-event",
				event:     []byte("test-data"),
			},
			mockErr: nil,
			wantErr: true,
		},
		{
			name: "empty device ID",
			req: SendEventToClientDeviceChannelRequest{
				clientID:  "test-client",
				deviceID:  "",
				eventName: "test-event",
				event:     []byte("test-data"),
			},
			mockErr: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.wantErr {
				mockPS.On("Publish", ctx, mock.Anything).Return(tt.mockErr).Once()
			}

			err := svc.PublishToClientWithDevice(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				mockPS.AssertExpectations(t)
			}
		})
	}
}

// Helper function to create a test subscription
func createTestSubscription() *subscription.Subscription {
	return &subscription.Subscription{
		TopicEventChan: make(chan broker.TopicEvent, 1),
		ErrChan:        make(chan error, 1),
	}
}

func TestGetClientActiveDevices(t *testing.T) {
	svc, mockPS, mockKVStore, ctx := setupTest(t)

	_ = map[string]string{
		"device1": `{"platform": "ios", "version": "1.0"}`,
		"device2": `{"platform": "android", "version": "2.0"}`,
	}

	tests := []struct {
		name          string
		req           GetClientActiveDevicesRequest
		setupMocks    func()
		expectedCount int
		wantErr       bool
	}{
		// {
		// 	name: "valid request with devices",
		// 	req:  GetClientActiveDevicesRequest{clientID: "test-client"},
		// 	setupMocks: func() {
		// 		sub := createTestSubscription()
		// 		mockKVStore.On("Load", ctx, "test-client").Return(testDevices, nil).Once()

		// 		// First, mock AsyncSubscribe for response channels
		// 		// Use MatchedBy to handle any order of channels
		// 		mockPS.On("AsyncSubscribe", ctx, mock.MatchedBy(func(channels []string) bool {
		// 			if len(channels) != len(testDevices) {
		// 				return false
		// 			}

		// 			// Create a map of expected channels
		// 			expectedChannels := make(map[string]bool)
		// 			for deviceID := range testDevices {
		// 				expectedChannels[fmt.Sprintf("%s#%s#%s", "test-client", deviceID, "resp")] = true
		// 			}

		// 			// Check if all channels in the argument are expected
		// 			for _, ch := range channels {
		// 				if !expectedChannels[ch] {
		// 					return false
		// 				}
		// 			}

		// 			return true
		// 		})).Return(sub, nil).Once()

		// 		// Create a map to track which devices have been published to
		// 		publishedDevices := make(map[string]bool)

		// 		// Mock Publish for device validation messages with a flexible matcher
		// 		mockPS.On("Publish", ctx, mock.MatchedBy(func(req pubsub.PublishRequest) bool {
		// 			// Extract the device ID from the channel name
		// 			channelParts := strings.Split(req.Channel, "--")
		// 			if len(channelParts) != 2 || channelParts[0] != "test-client" {
		// 				return false
		// 			}

		// 			deviceID := channelParts[1]
		// 			if _, exists := testDevices[deviceID]; !exists {
		// 				return false
		// 			}

		// 			// Verify the event is a device validation message
		// 			event := &pushv1.Event{}
		// 			if err := proto.Unmarshal(req.Data, event); err != nil {
		// 				return false
		// 			}

		// 			isValid := event.Name == DeviceValidation &&
		// 					   string(event.GetData().Value) == deviceID

		// 			// Track that this device has been published to
		// 			if isValid {
		// 				publishedDevices[deviceID] = true
		// 			}

		// 			return isValid
		// 		})).Run(func(args mock.Arguments) {
		// 			// Extract the device ID from the channel name
		// 			req := args.Get(1).(pubsub.PublishRequest)
		// 			channelParts := strings.Split(req.Channel, "--")
		// 			deviceID := channelParts[1]

		// 			// Simulate device response after publish
		// 			event := &pushv1.Event{
		// 				Name: DeviceValidation,
		// 				Data: &anypb.Any{Value: []byte(deviceID)},
		// 			}
		// 			eventBytes, _ := proto.Marshal(event)
		// 			sub.TopicEventChan <- broker.TopicEvent{Event: eventBytes}
		// 		}).Return(nil).Times(len(testDevices))

		// 		// Mock Delete call for devices that don't respond
		// 		for deviceID := range testDevices {
		// 			mockKVStore.On("Delete", ctx, "test-client", []string{deviceID}).Return(nil).Maybe()
		// 		}
		// 	},
		// 	expectedCount: len(testDevices),
		// 	wantErr:       false,
		// },
		{
			name: "empty client ID",
			req:  GetClientActiveDevicesRequest{clientID: ""},
			setupMocks: func() {
				// No mocks needed as it should fail validation
			},
			expectedCount: 0,
			wantErr:       true,
		},
		// {
		// 	name: "no devices found",
		// 	req:  GetClientActiveDevicesRequest{clientID: "test-client"},
		// 	setupMocks: func() {
		// 		mockKVStore.On("Load", ctx, "test-client").Return(map[string]string{}, nil).Once()
		// 		// No need to mock AsyncSubscribe or Publish since there are no devices
		// 	},
		// 	expectedCount: 0,
		// 	wantErr:       false,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			devices, err := svc.GetClientActiveDevices(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, devices, tt.expectedCount)
				mockKVStore.AssertExpectations(t)
				mockPS.AssertExpectations(t)
			}
		})
	}
}

func TestAsyncClientSubscribe(t *testing.T) {
	svc, mockPS, mockKVStore, ctx := setupTest(t)

	testSub := createTestSubscription()
	testDevice := &Device{
		ID: "test-device",
		Attributes: map[string]string{
			"platform": "ios",
			"version":  "1.0",
		},
	}

	tests := []struct {
		name       string
		clientID   string
		device     *Device
		setupMocks func()
		wantErr    bool
	}{
		{
			name:     "valid subscription without device",
			clientID: "test-client",
			device:   nil,
			setupMocks: func() {
				mockPS.On("AsyncSubscribe", ctx, []string{"test-client"}).Return(testSub, nil).Once()
			},
			wantErr: false,
		},
		{
			name:     "valid subscription with device",
			clientID: "test-client",
			device:   testDevice,
			setupMocks: func() {
				mockPS.On("AsyncSubscribe", ctx, []string{"test-client"}).Return(testSub, nil).Once()
				mockPS.On("AddSubscription", ctx, mock.Anything, testSub).Return(nil).Once()
				attrs, _ := json.Marshal(testDevice.Attributes)
				mockKVStore.On("Store", ctx, "test-client", testDevice.ID, string(attrs)).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name:     "empty client ID",
			clientID: "",
			device:   nil,
			setupMocks: func() {
				// No mocks needed as it should fail validation
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			sub, err := svc.AsyncClientSubscribe(ctx, tt.clientID, tt.device)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, sub)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, sub)
				mockPS.AssertExpectations(t)
				if tt.device != nil {
					mockKVStore.AssertExpectations(t)
				}
			}
		})
	}
}

func TestClientUnsubscribe(t *testing.T) {
	svc, mockPS, mockKVStore, ctx := setupTest(t)

	testSub := createTestSubscription()
	testDevice := &Device{
		ID: "test-device",
		Attributes: map[string]string{
			"platform": "ios",
			"version":  "1.0",
		},
	}

	tests := []struct {
		name       string
		clientID   string
		device     *Device
		setupMocks func()
		wantErr    bool
	}{
		{
			name:     "valid unsubscribe with device",
			clientID: "test-client",
			device:   testDevice,
			setupMocks: func() {
				mockKVStore.On("Delete", ctx, "test-client", []string{testDevice.ID}).Return(nil).Once()
				mockPS.On("Unsubscribe", ctx, testSub).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name:     "valid unsubscribe without device",
			clientID: "test-client",
			device:   nil,
			setupMocks: func() {
				mockPS.On("Unsubscribe", ctx, testSub).Return(nil).Once()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			err := svc.ClientUnsubscribe(ctx, tt.clientID, testSub, tt.device)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				mockPS.AssertExpectations(t)
				if tt.device != nil {
					mockKVStore.AssertExpectations(t)
				}
			}
		})
	}
}

func TestTopicSubscribeUnsubscribe(t *testing.T) {
	svc, mockPS, _, ctx := setupTest(t)
	testSub := createTestSubscription()

	t.Run("subscribe to topic", func(t *testing.T) {
		mockPS.On("AddSubscription", ctx, "test-topic", testSub).Return(nil).Once()
		err := svc.TopicSubscribe(ctx, "test-topic", testSub)
		assert.NoError(t, err)
		mockPS.AssertExpectations(t)
	})

	t.Run("unsubscribe from topic", func(t *testing.T) {
		mockPS.On("RemoveSubscription", ctx, "test-topic", testSub).Return(nil).Once()
		err := svc.TopicUnsubscribe(ctx, "test-topic", testSub)
		assert.NoError(t, err)
		mockPS.AssertExpectations(t)
	})
}

func TestPublishToTopics(t *testing.T) {
	svc, mockPS, _, ctx := setupTest(t)

	tests := []struct {
		name       string
		req        SendEventToTopicsRequest
		setupMocks func()
		wantErr    bool
	}{
		{
			name: "valid multiple requests",
			req: SendEventToTopicsRequest{
				requests: []SendEventToTopicRequest{
					{
						Topic:     "topic1",
						EventName: "event1",
						Event:     []byte("data1"),
					},
					{
						Topic:     "topic2",
						EventName: "event2",
						Event:     []byte("data2"),
					},
				},
			},
			setupMocks: func() {
				mockPS.On("PublishBulk", ctx, mock.Anything).Return(nil).Once()
			},
			wantErr: false,
		},
		{
			name: "empty requests",
			req: SendEventToTopicsRequest{
				requests: []SendEventToTopicRequest{},
			},
			setupMocks: func() {
				mockPS.On("PublishBulk", ctx, mock.Anything).Return(nil).Once()
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			err := svc.PublishToTopics(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				mockPS.AssertExpectations(t)
			}
		})
	}
}

func TestIsDeviceValidationMessage(t *testing.T) {
	svc, _, _, _ := setupTest(t)

	tests := []struct {
		name      string
		eventName string
		expected  bool
	}{
		{
			name:      "device validation message",
			eventName: DeviceValidation,
			expected:  true,
		},
		{
			name:      "other message",
			eventName: "OTHER_EVENT",
			expected:  false,
		},
		{
			name:      "empty message",
			eventName: "",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.IsDeviceValidationMessage(tt.eventName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfirmEventReceipt(t *testing.T) {
	svc, _, _, ctx := setupTest(t)

	// This is mostly for coverage as it just increments a metric
	svc.ConfirmEventReceipt(ctx, "test-event")
}

func TestReceiveDeviceResponse(t *testing.T) {
	svc, _, _, ctx := setupTest(t)
	sub := createTestSubscription()
	ch := make(chan map[string]bool)
	expectedCount := 2

	// Test case for receiving responses
	go func() {
		// Simulate device responses
		event := &pushv1.Event{
			Name: DeviceValidation,
			Data: &anypb.Any{Value: []byte("device1")},
		}
		eventBytes, _ := proto.Marshal(event)
		sub.TopicEventChan <- broker.TopicEvent{Event: eventBytes}

		event.Data.Value = []byte("device2")
		eventBytes, _ = proto.Marshal(event)
		sub.TopicEventChan <- broker.TopicEvent{Event: eventBytes}
	}()

	go func() {
		svc.receiveDeviceResponse(ctx, sub, ch, expectedCount)
	}()

	// Wait for responses
	result := <-ch
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.True(t, result["device1"])
	assert.True(t, result["device2"])
}
