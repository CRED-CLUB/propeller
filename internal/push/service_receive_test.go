package push

import (
	"testing"
	"time"

	"github.com/CRED-CLUB/propeller/internal/pubsub/subscription"
	"github.com/CRED-CLUB/propeller/pkg/broker"
	pushv1 "github.com/CRED-CLUB/propeller/rpc/push/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestReceiveDeviceResponseComprehensive(t *testing.T) {
	svc, _, _, ctx := setupTest(t)

	t.Run("receive responses within timeout", func(t *testing.T) {
		// Create a subscription with channels
		sub := &subscription.Subscription{
			TopicEventChan: make(chan broker.TopicEvent, 2),
			ErrChan:        make(chan error, 1),
		}

		// Create a channel to receive the responses
		ch := make(chan map[string]bool)

		// Expected device IDs
		deviceIDs := []string{"device1", "device2"}
		expectedCount := len(deviceIDs)

		// Start the receiveDeviceResponse function in a goroutine
		go svc.receiveDeviceResponse(ctx, sub, ch, expectedCount)

		// Send device responses through the TopicEventChan
		for _, deviceID := range deviceIDs {
			// Create a proto event with the device ID as the value
			event := &pushv1.Event{
				Name:       DeviceValidation,
				FormatType: pushv1.Event_TYPE_JSON_UNSPECIFIED,
				Data: &anypb.Any{
					TypeUrl: "",
					Value:   []byte(deviceID),
				},
			}
			eventBytes, _ := proto.Marshal(event)

			// Send the event through the channel
			sub.TopicEventChan <- broker.TopicEvent{
				Topic: "test-topic",
				Event: eventBytes,
			}
		}

		// Wait for the response
		responses := <-ch

		// Verify the responses
		assert.Equal(t, expectedCount, len(responses))
		for _, deviceID := range deviceIDs {
			assert.True(t, responses[deviceID])
		}
	})

	t.Run("timeout before receiving all responses", func(t *testing.T) {
		// Create a subscription with channels
		sub := &subscription.Subscription{
			TopicEventChan: make(chan broker.TopicEvent, 2),
			ErrChan:        make(chan error, 1),
		}

		// Create a channel to receive the responses
		ch := make(chan map[string]bool)

		// Expected device IDs
		deviceIDs := []string{"device1", "device2"}
		expectedCount := len(deviceIDs)

		// Start the receiveDeviceResponse function in a goroutine
		go svc.receiveDeviceResponse(ctx, sub, ch, expectedCount)

		// Send only one device response through the TopicEventChan
		event := &pushv1.Event{
			Name:       DeviceValidation,
			FormatType: pushv1.Event_TYPE_JSON_UNSPECIFIED,
			Data: &anypb.Any{
				TypeUrl: "",
				Value:   []byte(deviceIDs[0]),
			},
		}
		eventBytes, _ := proto.Marshal(event)

		// Send the event through the channel
		sub.TopicEventChan <- broker.TopicEvent{
			Topic: "test-topic",
			Event: eventBytes,
		}

		// Wait for the response (should timeout after 1 second)
		responses := <-ch

		// Verify the responses
		assert.Equal(t, 1, len(responses))
		assert.True(t, responses[deviceIDs[0]])
		assert.False(t, responses[deviceIDs[1]])
	})

	t.Run("handle error from subscription", func(t *testing.T) {
		// Create a subscription with channels
		sub := &subscription.Subscription{
			TopicEventChan: make(chan broker.TopicEvent, 2),
			ErrChan:        make(chan error, 1),
		}

		// Create a channel to receive the responses
		ch := make(chan map[string]bool)

		// Expected count
		expectedCount := 1

		// Start the receiveDeviceResponse function in a goroutine
		go svc.receiveDeviceResponse(ctx, sub, ch, expectedCount)

		// Send an error through the ErrChan
		sub.ErrChan <- assert.AnError

		// Wait a short time for the error to be processed
		time.Sleep(100 * time.Millisecond)

		// Send a device response after the error
		event := &pushv1.Event{
			Name:       DeviceValidation,
			FormatType: pushv1.Event_TYPE_JSON_UNSPECIFIED,
			Data: &anypb.Any{
				TypeUrl: "",
				Value:   []byte("device1"),
			},
		}
		eventBytes, _ := proto.Marshal(event)

		// Send the event through the channel
		sub.TopicEventChan <- broker.TopicEvent{
			Topic: "test-topic",
			Event: eventBytes,
		}

		// Wait for the response
		responses := <-ch

		// Verify the responses (should still process valid events after an error)
		assert.Equal(t, 1, len(responses))
		assert.True(t, responses["device1"])
	})

	t.Run("handle invalid proto message", func(t *testing.T) {
		// Create a subscription with channels
		sub := &subscription.Subscription{
			TopicEventChan: make(chan broker.TopicEvent, 2),
			ErrChan:        make(chan error, 1),
		}

		// Create a channel to receive the responses
		ch := make(chan map[string]bool)

		// Expected count
		expectedCount := 1

		// Start the receiveDeviceResponse function in a goroutine
		go svc.receiveDeviceResponse(ctx, sub, ch, expectedCount)

		// Send an invalid proto message
		sub.TopicEventChan <- broker.TopicEvent{
			Topic: "test-topic",
			Event: []byte("invalid-proto"),
		}

		// Wait for the response (should timeout after 2 seconds)
		select {
		case responses := <-ch:
			// Verify the responses (should be empty due to invalid proto)
			assert.Equal(t, 0, len(responses))
		case <-time.After(2 * time.Second):
			t.Fatal("test timed out")
		}
	})
}
