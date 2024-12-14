package push

import (
	"context"
	"testing"
	"time"

	pushv1 "github.com/CRED-CLUB/propeller/rpc/push/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestGetClientActiveDevicesRequest_PopulateFromProto(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		proto   *pushv1.GetClientActiveDevicesRequest
		want    string
		wantErr bool
	}{
		{
			name:    "valid client ID",
			proto:   &pushv1.GetClientActiveDevicesRequest{ClientId: "test-client"},
			want:    "test-client",
			wantErr: false,
		},
		{
			name:    "empty client ID",
			proto:   &pushv1.GetClientActiveDevicesRequest{ClientId: ""},
			want:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &GetClientActiveDevicesRequest{}
			err := g.PopulateFromProto(ctx, tt.proto)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, g.clientID)
		})
	}
}

func TestSendEventToClientChannelRequest_PopulateFromProto(t *testing.T) {
	ctx := context.Background()
	testData, _ := anypb.New(&timestamppb.Timestamp{Seconds: 1234})
	tests := []struct {
		name    string
		proto   *pushv1.SendEventToClientChannelRequest
		wantErr bool
	}{
		{
			name: "valid request",
			proto: &pushv1.SendEventToClientChannelRequest{
				ClientId: "test-client",
				Event: &pushv1.Event{
					Name: "test-event",
					Data: testData,
				},
			},
			wantErr: false,
		},
		{
			name: "nil event",
			proto: &pushv1.SendEventToClientChannelRequest{
				ClientId: "test-client",
				Event:    nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			smu := &SendEventToClientChannelRequest{}
			err := smu.PopulateFromProto(ctx, tt.proto)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.proto.GetClientId(), smu.clientID)
			if tt.proto.Event != nil {
				assert.Equal(t, tt.proto.Event.Name, smu.eventName)
				assert.NotNil(t, smu.event)
			}
		})
	}
}

func TestSendEventToClientDeviceChannelRequest_PopulateFromProto(t *testing.T) {
	ctx := context.Background()
	testData, _ := anypb.New(&timestamppb.Timestamp{Seconds: 1234})
	tests := []struct {
		name    string
		proto   *pushv1.SendEventToClientDeviceChannelRequest
		wantErr bool
	}{
		{
			name: "valid request",
			proto: &pushv1.SendEventToClientDeviceChannelRequest{
				ClientId: "test-client",
				DeviceId: "test-device",
				Event: &pushv1.Event{
					Name: "test-event",
					Data: testData,
				},
			},
			wantErr: false,
		},
		{
			name: "nil event",
			proto: &pushv1.SendEventToClientDeviceChannelRequest{
				ClientId: "test-client",
				DeviceId: "test-device",
				Event:    nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			smu := &SendEventToClientDeviceChannelRequest{}
			err := smu.PopulateFromProto(ctx, tt.proto)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.proto.GetClientId(), smu.clientID)
			assert.Equal(t, tt.proto.GetDeviceId(), smu.deviceID)
			if tt.proto.Event != nil {
				assert.Equal(t, tt.proto.Event.Name, smu.eventName)
				assert.NotNil(t, smu.event)
			}
		})
	}
}

func TestSendEventToTopicRequest_PopulateFromProto(t *testing.T) {
	ctx := context.Background()
	testData, _ := anypb.New(&timestamppb.Timestamp{Seconds: 1234})
	tests := []struct {
		name    string
		proto   *pushv1.SendEventToTopicRequest
		wantErr bool
	}{
		{
			name: "valid request",
			proto: &pushv1.SendEventToTopicRequest{
				Topic: "test-topic",
				Event: &pushv1.Event{
					Name: "test-event",
					Data: testData,
				},
			},
			wantErr: false,
		},
		{
			name: "nil event",
			proto: &pushv1.SendEventToTopicRequest{
				Topic: "test-topic",
				Event: nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			smt := &SendEventToTopicRequest{}
			err := smt.PopulateFromProto(ctx, tt.proto)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.proto.GetTopic(), smt.Topic)
			if tt.proto.Event != nil {
				assert.Equal(t, tt.proto.Event.Name, smt.EventName)
				assert.NotNil(t, smt.Event)
			}
		})
	}
}

func TestSendEventToTopicsRequest_PopulateFromProto(t *testing.T) {
	ctx := context.Background()
	testData, _ := anypb.New(&timestamppb.Timestamp{Seconds: 1234})
	tests := []struct {
		name    string
		proto   *pushv1.SendEventToTopicsRequest
		wantLen int
		wantErr bool
	}{
		{
			name: "valid multiple requests",
			proto: &pushv1.SendEventToTopicsRequest{
				Requests: []*pushv1.SendEventToTopicRequest{
					{
						Topic: "topic1",
						Event: &pushv1.Event{Name: "event1", Data: testData},
					},
					{
						Topic: "topic2",
						Event: &pushv1.Event{Name: "event2", Data: testData},
					},
				},
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "empty requests",
			proto: &pushv1.SendEventToTopicsRequest{
				Requests: []*pushv1.SendEventToTopicRequest{},
			},
			wantLen: 0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := &SendEventToTopicsRequest{}
			err := sm.PopulateFromProto(ctx, tt.proto)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Len(t, sm.requests, tt.wantLen)
		})
	}
}

func TestDevice_ToProto(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name string
		dev  Device
		want *pushv1.Device
	}{
		{
			name: "valid device",
			dev: Device{
				ID: "test-device",
				Attributes: map[string]string{
					"attr1": "value1",
					"attr2": "value2",
				},
				LoggedInAt: now,
			},
			want: &pushv1.Device{
				Id: "test-device",
				Attributes: map[string]string{
					"attr1": "value1",
					"attr2": "value2",
				},
				LoggedInAt: timestamppb.New(now),
			},
		},
		{
			name: "empty attributes",
			dev: Device{
				ID:         "test-device",
				Attributes: map[string]string{},
				LoggedInAt: now,
			},
			want: &pushv1.Device{
				Id:         "test-device",
				Attributes: map[string]string{},
				LoggedInAt: timestamppb.New(now),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.dev.ToProto()
			assert.Equal(t, tt.want.Id, got.Id)
			assert.Equal(t, tt.want.Attributes, got.Attributes)
			assert.Equal(t, tt.want.LoggedInAt.AsTime().Unix(), got.LoggedInAt.AsTime().Unix())
		})
	}
}

func TestGetDummyHelperFunctions(t *testing.T) {
	t.Run("getDummySendEventToClientRequest", func(t *testing.T) {
		req := getDummySendEventToClientRequest("test-client")
		assert.Equal(t, "test-client", req.clientID)
		assert.NotNil(t, req.event)
	})

	t.Run("getDummySendEventToClientDeviceRequest", func(t *testing.T) {
		req := getDummySendEventToClientDeviceRequest("test-client", "test-device")
		assert.Equal(t, "test-client", req.clientID)
		assert.Equal(t, "test-device", req.deviceID)
		assert.NotNil(t, req.event)
	})

	t.Run("getDummySendEventToTopicRequest", func(t *testing.T) {
		req := getDummySendEventToTopicRequest("test-topic")
		assert.Equal(t, "test-topic", req.Topic)
		assert.Equal(t, "test-topic", req.EventName)
		assert.NotNil(t, req.Event)
	})

	t.Run("getDummyJSONBytes", func(t *testing.T) {
		jsonBytes := getDummyJSONBytes()
		assert.NotNil(t, jsonBytes)
		assert.True(t, len(jsonBytes) > 0)
	})
}
