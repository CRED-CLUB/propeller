package push

import (
	"errors"
	"testing"

	"github.com/CRED-CLUB/propeller/internal/perror"
	"github.com/CRED-CLUB/propeller/internal/pubsub/subscription"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAsyncClientSubscribeComprehensive(t *testing.T) {
	svc, mockPS, mockKV, ctx := setupTest(t)

	// Create a test subscription
	testSub := &subscription.Subscription{
		ID: uuid.New(),
	}

	tests := []struct {
		name           string
		clientID       string
		device         *Device
		setupMocks     func()
		expectedErr    error
		validateErr    func(t *testing.T, err error)
		validateResult func(t *testing.T, sub *subscription.Subscription)
	}{
		{
			name:     "successful subscription without device",
			clientID: "test-client",
			device:   nil,
			setupMocks: func() {
				mockPS.On("AsyncSubscribe", ctx, []string{"test-client"}).Return(testSub, nil).Once()
			},
			validateResult: func(t *testing.T, sub *subscription.Subscription) {
				assert.Equal(t, testSub, sub)
			},
		},
		{
			name:     "successful subscription with device",
			clientID: "test-client",
			device: &Device{
				ID: "test-device",
				Attributes: map[string]string{
					"platform": "ios",
					"version":  "1.0",
				},
			},
			setupMocks: func() {
				mockPS.On("AsyncSubscribe", ctx, []string{"test-client"}).Return(testSub, nil).Once()
				mockPS.On("AddSubscription", ctx, "test-client--test-device", testSub).Return(nil).Once()
				mockKV.On("Store", ctx, "test-client", "test-device", mock.AnythingOfType("string")).Return(nil).Once()
			},
			validateResult: func(t *testing.T, sub *subscription.Subscription) {
				assert.Equal(t, testSub, sub)
			},
		},
		{
			name:     "empty client ID",
			clientID: "",
			device:   nil,
			validateErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				perr, ok := err.(*perror.PError)
				assert.True(t, ok)
				assert.Equal(t, perror.InvalidArgument, perr.Code())
				assert.Contains(t, perr.Error(), "client ID is empty")
			},
		},
		{
			name:     "AsyncSubscribe error",
			clientID: "test-client",
			device:   nil,
			setupMocks: func() {
				mockPS.On("AsyncSubscribe", ctx, []string{"test-client"}).Return(&subscription.Subscription{}, errors.New("subscription error")).Once()
			},
			expectedErr: errors.New("subscription error"),
		},
		{
			name:     "AddSubscription error with device",
			clientID: "test-client",
			device: &Device{
				ID: "test-device",
				Attributes: map[string]string{
					"platform": "ios",
					"version":  "1.0",
				},
			},
			setupMocks: func() {
				mockPS.On("AsyncSubscribe", ctx, []string{"test-client"}).Return(testSub, nil).Once()
				mockPS.On("AddSubscription", ctx, "test-client--test-device", testSub).Return(errors.New("add subscription error")).Once()
			},
			expectedErr: errors.New("add subscription error"),
		},
		{
			name:     "Store error with device",
			clientID: "test-client",
			device: &Device{
				ID: "test-device",
				Attributes: map[string]string{
					"platform": "ios",
					"version":  "1.0",
				},
			},
			setupMocks: func() {
				mockPS.On("AsyncSubscribe", ctx, []string{"test-client"}).Return(testSub, nil).Once()
				mockPS.On("AddSubscription", ctx, "test-client--test-device", testSub).Return(nil).Once()
				mockKV.On("Store", ctx, "test-client", "test-device", mock.AnythingOfType("string")).Return(errors.New("store error")).Once()
			},
			expectedErr: errors.New("store error"),
		},
		{
			name:     "device with empty ID",
			clientID: "test-client",
			device: &Device{
				ID: "",
				Attributes: map[string]string{
					"platform": "ios",
					"version":  "1.0",
				},
			},
			validateErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				perr, ok := err.(*perror.PError)
				assert.True(t, ok)
				assert.Equal(t, perror.InvalidArgument, perr.Code())
				assert.Contains(t, perr.Error(), "Device ID is empty")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks if provided
			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			// Call the method
			sub, err := svc.AsyncClientSubscribe(ctx, tt.clientID, tt.device)

			// Validate the result
			if tt.validateErr != nil {
				tt.validateErr(t, err)
				assert.Nil(t, sub)
			} else if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
				assert.Nil(t, sub)
			} else {
				assert.NoError(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, sub)
				}
			}

			// Verify all expectations were met
			mockPS.AssertExpectations(t)
			mockKV.AssertExpectations(t)
		})
	}
}

func TestClientUnsubscribeComprehensive(t *testing.T) {
	svc, mockPS, mockKV, ctx := setupTest(t)

	// Create a test subscription
	testSub := &subscription.Subscription{
		ID: uuid.New(),
	}

	tests := []struct {
		name        string
		clientID    string
		device      *Device
		setupMocks  func()
		expectedErr error
	}{
		{
			name:     "successful unsubscribe without device",
			clientID: "test-client",
			device:   nil,
			setupMocks: func() {
				mockPS.On("Unsubscribe", ctx, testSub).Return(nil).Once()
			},
		},
		{
			name:     "successful unsubscribe with device",
			clientID: "test-client",
			device: &Device{
				ID: "test-device",
				Attributes: map[string]string{
					"platform": "ios",
					"version":  "1.0",
				},
			},
			setupMocks: func() {
				mockKV.On("Delete", ctx, "test-client", []string{"test-device"}).Return(nil).Once()
				mockPS.On("Unsubscribe", ctx, testSub).Return(nil).Once()
			},
		},
		{
			name:     "Delete error with device",
			clientID: "test-client",
			device: &Device{
				ID: "test-device",
				Attributes: map[string]string{
					"platform": "ios",
					"version":  "1.0",
				},
			},
			setupMocks: func() {
				mockKV.On("Delete", ctx, "test-client", []string{"test-device"}).Return(errors.New("delete error")).Once()
				// The implementation logs the error but continues with Unsubscribe
				mockPS.On("Unsubscribe", ctx, testSub).Return(nil).Once()
			},
			// The implementation doesn't return the Delete error, it just logs it
			expectedErr: nil,
		},
		{
			name:     "Unsubscribe error",
			clientID: "test-client",
			device:   nil,
			setupMocks: func() {
				mockPS.On("Unsubscribe", ctx, testSub).Return(errors.New("unsubscribe error")).Once()
			},
			expectedErr: errors.New("unsubscribe error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks if provided
			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			// Call the method
			err := svc.ClientUnsubscribe(ctx, tt.clientID, testSub, tt.device)

			// Validate the result
			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations were met
			mockPS.AssertExpectations(t)
			mockKV.AssertExpectations(t)
		})
	}
}

func TestTopicSubscribeUnsubscribeComprehensive(t *testing.T) {
	svc, mockPS, _, ctx := setupTest(t)

	// Create a test subscription
	testSub := &subscription.Subscription{
		ID: uuid.New(),
	}

	t.Run("topic subscribe success", func(t *testing.T) {
		mockPS.On("AddSubscription", ctx, "test-topic", testSub).Return(nil).Once()
		err := svc.TopicSubscribe(ctx, "test-topic", testSub)
		assert.NoError(t, err)
		mockPS.AssertExpectations(t)
	})

	t.Run("topic subscribe error", func(t *testing.T) {
		mockPS.On("AddSubscription", ctx, "test-topic", testSub).Return(errors.New("add subscription error")).Once()
		err := svc.TopicSubscribe(ctx, "test-topic", testSub)
		assert.Error(t, err)
		assert.Equal(t, "add subscription error", err.Error())
		mockPS.AssertExpectations(t)
	})

	t.Run("topic unsubscribe success", func(t *testing.T) {
		mockPS.On("RemoveSubscription", ctx, "test-topic", testSub).Return(nil).Once()
		err := svc.TopicUnsubscribe(ctx, "test-topic", testSub)
		assert.NoError(t, err)
		mockPS.AssertExpectations(t)
	})

	t.Run("topic unsubscribe error", func(t *testing.T) {
		mockPS.On("RemoveSubscription", ctx, "test-topic", testSub).Return(errors.New("remove subscription error")).Once()
		err := svc.TopicUnsubscribe(ctx, "test-topic", testSub)
		assert.Error(t, err)
		assert.Equal(t, "remove subscription error", err.Error())
		mockPS.AssertExpectations(t)
	})

	t.Run("empty topic name", func(t *testing.T) {
		err := svc.TopicSubscribe(ctx, "", testSub)
		assert.Error(t, err)
		perr, ok := err.(*perror.PError)
		assert.True(t, ok)
		assert.Equal(t, perror.InvalidArgument, perr.Code())
		assert.Contains(t, perr.Error(), "Topic is empty")

		err = svc.TopicUnsubscribe(ctx, "", testSub)
		assert.Error(t, err)
		perr, ok = err.(*perror.PError)
		assert.True(t, ok)
		assert.Equal(t, perror.InvalidArgument, perr.Code())
		assert.Contains(t, perr.Error(), "Topic is empty")
	})

	t.Run("nil subscription", func(t *testing.T) {
		err := svc.TopicSubscribe(ctx, "test-topic", nil)
		assert.Error(t, err)
		perr, ok := err.(*perror.PError)
		assert.True(t, ok)
		assert.Equal(t, perror.InvalidArgument, perr.Code())
		assert.Contains(t, perr.Error(), "Subscription is nil")

		err = svc.TopicUnsubscribe(ctx, "test-topic", nil)
		assert.Error(t, err)
		perr, ok = err.(*perror.PError)
		assert.True(t, ok)
		assert.Equal(t, perror.InvalidArgument, perr.Code())
		assert.Contains(t, perr.Error(), "Subscription is nil")
	})
}
