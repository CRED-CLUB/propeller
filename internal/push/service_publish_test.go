package push

import (
	"errors"
	"testing"

	"github.com/CRED-CLUB/propeller/internal/perror"
	"github.com/CRED-CLUB/propeller/internal/pubsub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPublishToTopicComprehensive(t *testing.T) {
	svc, mockPS, _, ctx := setupTest(t)

	tests := []struct {
		name           string
		req            SendEventToTopicRequest
		mockPublishErr error
		expectedErr    error
		validateErr    func(t *testing.T, err error)
	}{
		{
			name: "successful publish",
			req: SendEventToTopicRequest{
				Topic:     "test-topic",
				EventName: "test-event",
				Event:     []byte("test-data"),
			},
			mockPublishErr: nil,
			expectedErr:    nil,
		},
		{
			name: "empty topic",
			req: SendEventToTopicRequest{
				Topic:     "",
				EventName: "test-event",
				Event:     []byte("test-data"),
			},
			mockPublishErr: nil,
			validateErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				perr, ok := err.(*perror.PError)
				assert.True(t, ok)
				assert.Equal(t, perror.InvalidArgument, perr.Code())
				assert.Contains(t, perr.Error(), "Topic is empty")
			},
		},
		{
			name: "nil event",
			req: SendEventToTopicRequest{
				Topic:     "test-topic",
				EventName: "test-event",
				Event:     nil,
			},
			mockPublishErr: nil,
			validateErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				perr, ok := err.(*perror.PError)
				assert.True(t, ok)
				assert.Equal(t, perror.InvalidArgument, perr.Code())
				assert.Contains(t, perr.Error(), "Event is empty")
			},
		},
		{
			name: "empty event name",
			req: SendEventToTopicRequest{
				Topic:     "test-topic",
				EventName: "",
				Event:     []byte("test-data"),
			},
			mockPublishErr: nil,
			validateErr: func(t *testing.T, err error) {
				assert.Error(t, err)
				perr, ok := err.(*perror.PError)
				assert.True(t, ok)
				assert.Equal(t, perror.InvalidArgument, perr.Code())
				assert.Contains(t, perr.Error(), "Event name is empty")
			},
		},
		{
			name: "publish error",
			req: SendEventToTopicRequest{
				Topic:     "test-topic",
				EventName: "test-event",
				Event:     []byte("test-data"),
			},
			mockPublishErr: errors.New("publish failed"),
			expectedErr:    errors.New("publish failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations only for valid requests
			if tt.req.Topic != "" && tt.req.Event != nil && tt.req.EventName != "" {
				mockPS.On("Publish", ctx, mock.MatchedBy(func(req pubsub.PublishRequest) bool {
					return req.Channel == tt.req.Topic && string(req.Data) == string(tt.req.Event)
				})).Return(tt.mockPublishErr).Once()
			}

			// Call the method
			err := svc.PublishToTopic(ctx, tt.req)

			// Validate the result
			if tt.validateErr != nil {
				tt.validateErr(t, err)
			} else if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations were met
			mockPS.AssertExpectations(t)
		})
	}
}

func TestPublishToTopicWithDifferentTopicFormats(t *testing.T) {
	svc, mockPS, _, ctx := setupTest(t)

	tests := []struct {
		name  string
		topic string
	}{
		{
			name:  "simple topic",
			topic: "simple-topic",
		},
		{
			name:  "topic with dots",
			topic: "topic.with.dots",
		},
		{
			name:  "topic with special characters",
			topic: "topic-with_special#characters",
		},
		{
			name:  "topic with client and device format",
			topic: "client--device",
		},
		{
			name:  "topic with numeric values",
			topic: "topic-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := SendEventToTopicRequest{
				Topic:     tt.topic,
				EventName: "test-event",
				Event:     []byte("test-data"),
			}

			// Setup mock expectations
			mockPS.On("Publish", ctx, mock.MatchedBy(func(pubReq pubsub.PublishRequest) bool {
				return pubReq.Channel == tt.topic
			})).Return(nil).Once()

			// Call the method
			err := svc.PublishToTopic(ctx, req)

			// Validate the result
			assert.NoError(t, err)
			mockPS.AssertExpectations(t)
		})
	}
}

func TestPublishToTopicMetricsIncrement(t *testing.T) {
	svc, mockPS, _, ctx := setupTest(t)

	// Setup a valid request
	req := SendEventToTopicRequest{
		Topic:     "test-topic",
		EventName: "test-event",
		Event:     []byte("test-data"),
	}

	// Setup mock expectations
	mockPS.On("Publish", ctx, mock.Anything).Return(nil).Once()

	// Call the method
	err := svc.PublishToTopic(ctx, req)

	// Validate the result
	assert.NoError(t, err)
	mockPS.AssertExpectations(t)

	// Note: We can't directly test the metrics increment since it's a global variable
	// In a real test environment, we might use a metrics mock or a test registry
	// For now, we're just ensuring the code path that increments metrics is executed
}
