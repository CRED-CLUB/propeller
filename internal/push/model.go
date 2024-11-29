package push

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/CRED-CLUB/propeller/internal/dummy"
	"github.com/CRED-CLUB/propeller/internal/perror"
	"github.com/CRED-CLUB/propeller/pkg/logger"
	pushv1 "github.com/CRED-CLUB/propeller/rpc/push/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GetClientActiveDevicesRequest model
type GetClientActiveDevicesRequest struct {
	clientID string
}

// PopulateFromProto maps model from proto
func (g *GetClientActiveDevicesRequest) PopulateFromProto(ctx context.Context, proto *pushv1.GetClientActiveDevicesRequest) error {
	g.clientID = proto.GetClientId()
	return nil
}

// SendEventToClientChannelRequest model
type SendEventToClientChannelRequest struct {
	clientID  string
	eventName string
	event     []byte
}

// PopulateFromProto maps model from proto
func (smu *SendEventToClientChannelRequest) PopulateFromProto(ctx context.Context, protoRequest *pushv1.SendEventToClientChannelRequest) error {
	smu.clientID = protoRequest.GetClientId()
	if protoRequest.Event != nil {
		eventBytes, err := proto.Marshal(protoRequest.Event)
		if err != nil {
			pErr := perror.Newf(perror.Internal, "unable to marshall proto")
			logger.Ctx(ctx).Error(pErr.Error())
			return pErr
		}
		smu.event = eventBytes
		smu.eventName = protoRequest.Event.Name
	}
	return nil
}

// SendEventToClientDeviceChannelRequest model
type SendEventToClientDeviceChannelRequest struct {
	clientID  string
	deviceID  string
	eventName string
	event     []byte
}

// PopulateFromProto maps model from proto
func (smu *SendEventToClientDeviceChannelRequest) PopulateFromProto(ctx context.Context, protoRequest *pushv1.SendEventToClientDeviceChannelRequest) error {
	smu.clientID = protoRequest.GetClientId()
	smu.deviceID = protoRequest.GetDeviceId()
	if protoRequest.Event != nil {
		eventBytes, err := proto.Marshal(protoRequest.Event)
		if err != nil {
			pErr := perror.Newf(perror.Internal, "unable to marshall proto")
			logger.Ctx(ctx).Error(pErr.Error())
			return pErr
		}
		smu.event = eventBytes
		smu.eventName = protoRequest.Event.Name
	}
	return nil
}

// TopicSubscriptionRequest model
type TopicSubscriptionRequest struct {
	topicToSubscribe string
}

// TopicUnSubscriptionRequest model
type TopicUnSubscriptionRequest struct {
	topicToUnSubscribe string
}

// SendEventToTopicRequest model
type SendEventToTopicRequest struct {
	Topic     string
	EventName string
	Event     []byte
}

// PopulateFromProto maps model from proto
func (smt *SendEventToTopicRequest) PopulateFromProto(ctx context.Context, protoRequest *pushv1.SendEventToTopicRequest) error {
	smt.Topic = protoRequest.GetTopic()
	if protoRequest.Event != nil {
		eventBytes, err := proto.Marshal(protoRequest.Event)
		if err != nil {
			pErr := perror.Newf(perror.Internal, "unable to marshall proto")
			logger.Ctx(ctx).Error(pErr.Error())
			return pErr
		}
		smt.Event = eventBytes
		smt.EventName = protoRequest.Event.Name
	}
	return nil
}

// SendEventToTopicsRequest model
type SendEventToTopicsRequest struct {
	requests []SendEventToTopicRequest
}

// PopulateFromProto maps model from proto
func (sm *SendEventToTopicsRequest) PopulateFromProto(ctx context.Context, proto *pushv1.SendEventToTopicsRequest) error {
	var req SendEventToTopicRequest
	for _, v := range proto.Requests {
		req = SendEventToTopicRequest{}
		err := req.PopulateFromProto(ctx, v)
		if err != nil {
			return err
		}
		sm.requests = append(sm.requests, req)
	}
	return nil
}

func getDummySendEventToClientRequest(clientID string) SendEventToClientChannelRequest {
	req := SendEventToClientChannelRequest{
		clientID: clientID,
		event:    dummy.GetDummyEvent("dummy-client", getDummyJSONBytes()),
	}
	return req
}

func getDummySendEventToClientDeviceRequest(clientID, deviceID string) SendEventToClientDeviceChannelRequest {
	req := SendEventToClientDeviceChannelRequest{
		clientID: clientID,
		deviceID: deviceID,
		event:    dummy.GetDummyEvent("dummy-device", getDummyJSONBytes()),
	}
	return req

}

func getDummySendEventToTopicRequest(topic string) SendEventToTopicRequest {
	req := SendEventToTopicRequest{
		Topic:     topic,
		EventName: topic,
		Event:     dummy.GetDummyEvent(topic, getDummyJSONBytes()),
	}
	return req
}

func getDummyJSONBytes() []byte {
	testData := struct {
		Field string
	}{Field: fmt.Sprintf("time %s", time.Now())}

	jsonBytes, _ := json.Marshal(testData)
	return jsonBytes
}

// Device struct for device and attrs
type Device struct {
	ID         string
	Attributes map[string]string
	LoggedInAt time.Time
}

// ToProto converts to proto
func (d Device) ToProto() *pushv1.Device {
	device := &pushv1.Device{}
	device.Id = d.ID
	device.LoggedInAt = timestamppb.New(d.LoggedInAt)
	device.Attributes = d.Attributes
	return device
}
