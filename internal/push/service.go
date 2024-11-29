package push

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/CRED-CLUB/propeller/internal/config"
	"github.com/CRED-CLUB/propeller/internal/kv"
	"github.com/CRED-CLUB/propeller/internal/perror"
	"github.com/CRED-CLUB/propeller/internal/pubsub"
	"github.com/CRED-CLUB/propeller/internal/pubsub/subscription"
	"github.com/CRED-CLUB/propeller/pkg/logger"
	pushv1 "github.com/CRED-CLUB/propeller/rpc/push/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	// DeviceValidation is topic name for device validation messages
	DeviceValidation string = "DEVICE_VALIDATION"
)

// Service of the push service
type Service struct {
	pubSub           pubsub.IPubSub
	kv               kv.IKV
	config           config.Config
	testCancelFunc   context.CancelFunc
	sessionStartTime time.Time
}

// NewService returns a new instance of Service
func NewService(pubSub pubsub.IPubSub, kv kv.IKV, config config.Config) *Service {
	return &Service{pubSub: pubSub, kv: kv, config: config}
}

// GetClientActiveDevices ...
/*
	                                      ┌──────┐
	      2. subscribe(clnt, clnt-device) │      │◄──────────┐
	                                 ┌───►│redis │           │
	                  ┌───────────┐  │    │pubsub├──────┐    │
	┌───┐             │           ├──┘    ├───▲──┘      │    │
	│   │1. connect   │           │       │   │         │    │
	│app├────────────►│propeller  │◄──────┘   │         │    │
	│   │             │           │9.         │         │    │8.
	└───┘             │           ├───────────┘         │    │
	                  │           │10. validation resp  │    │validation
	                  └───────────┴──┐    ┌───────┐     │    │message
	                                 │    │       │     │    │published
	   3. store(clnt, device, attrs) │    │ redis │     │    │
	                                 └───►│  kv   │     │11. │
	                                      │       │     │    │
	                                      └───────┘     │    │
	                    5. devices = load(clnt)         │    │
	            13.delete stale entries   ┌──────┐      │    │
	                  ┌────────────┐      │      │      │    │
	                  │            ├─────►│redis │      │    │
	┌───┐             │            │      │ kv   │      │    │
	│   │4.getDevices │            │      └──────┘      │    │
	│ BE├────────────►│propeller   │12. validation recv │    │
	│   │             │            │◄─────┬──────┐      │    │
	│   │14. reply    │            │6.    │      │◄─────┘    │
	│   │◄────────────┤            ├─────►│redis │           │
	└───┘             │            │7.    │pubsub├───────────┘
	                  └────────────┴─────►└──────┘
	        6. subscribe(clnt-device-resp)
	   7. publishToTopic(clnt-device, validation message)
*/
func (c *Service) GetClientActiveDevices(ctx context.Context, req GetClientActiveDevicesRequest) ([]Device, error) {
	logger.Ctx(ctx).Infow("getting client devices")
	if !c.config.EnableDeviceSupport {
		return nil, perror.New(perror.FailedPrecondition, "device support disabled")
	}

	if req.clientID == "" {
		return nil, perror.New(perror.InvalidArgument, "client id required")
	}

	// load device entries from kv store
	v, err := c.kv.Load(ctx, req.clientID)
	if err != nil {
		logger.Ctx(ctx).Errorw("error in getting attr", "attr", v, "err", err)
	}

	if len(v) == 0 {
		return []Device{}, nil
	}

	// prepare to send validation pubsub message for devices found
	var foundDevices []Device
	var foundDeviceIDs []string
	for deviceID, attrs := range v {
		// TODO: fix
		logger.Ctx(ctx).Infow("found device", "deviceID", deviceID, "attrs", attrs)
		attrsMap := make(map[string]string)
		err := json.Unmarshal([]byte(attrs), &attrsMap)
		if err != nil {
			logger.Ctx(ctx).Errorw("error unmarshalling attrs", "err", err)
			return []Device{}, err
		}
		foundDevices = append(foundDevices, Device{ID: deviceID, Attributes: attrsMap})
		foundDeviceIDs = append(foundDeviceIDs, fmt.Sprintf("%s#%s#%s", req.clientID, deviceID, "resp"))
	}

	// subscribe to response channels
	s, err := c.pubSub.AsyncSubscribe(ctx, foundDeviceIDs...)
	if err != nil {
		return nil, err
	}
	ch := make(chan map[string]bool)
	go c.receiveDeviceResponse(ctx, s, ch, len(v))

	// publish validation message
	for k := range v {
		topic := fmt.Sprintf("%s--%s", req.clientID, k)

		eventToPublish := pushv1.Event{
			Name:       DeviceValidation,
			FormatType: pushv1.Event_TYPE_JSON_UNSPECIFIED,
			Data: &anypb.Any{
				TypeUrl: "",
				Value:   []byte(k),
			},
		}
		pbEvent, err := proto.Marshal(&eventToPublish)
		if err != nil {
			return nil, err
		}
		err = c.PublishToTopic(ctx, SendEventToTopicRequest{
			Topic: topic,
			Event: pbEvent,
		})
		if err != nil {
		}
	}
	responses := <-ch

	var result []Device
	// prepare the result, delete stray device entries
	for _, attrs := range foundDevices {
		_, ok := responses[attrs.ID]
		if ok {
			result = append(result, Device{
				ID:         attrs.ID,
				Attributes: attrs.Attributes,
			})
		} else {
			_ = c.kv.Delete(ctx, req.clientID, attrs.ID)
		}
	}
	return result, nil
}

func (c *Service) receiveDeviceResponse(ctx context.Context, s *subscription.Subscription, ch chan map[string]bool, expectedCount int) {
	responses := make(map[string]bool)
	timeout := time.NewTimer(1 * time.Second)
	receivedSoFar := 0
	for {
		select {
		case topicEvent := <-s.TopicEventChan:
			receivedSoFar++
			protoEvent := &pushv1.Event{}
			msg := topicEvent.Event
			err := proto.Unmarshal(msg, protoEvent)
			if err != nil {
				return
			}
			responses[string(protoEvent.GetData().Value)] = true
			if receivedSoFar == expectedCount {
				ch <- responses
				return
			}
		case err := <-s.ErrChan:
			logger.Ctx(ctx).Errorw("error in subscriber", "error", err.Error())
		case <-timeout.C:
			ch <- responses
			return
		}
	}
}

// PublishToClient publishes to the client
func (c *Service) PublishToClient(ctx context.Context, req SendEventToClientChannelRequest) error {
	logger.Ctx(ctx).Infow("publishing to client")

	err := c.validateSendEventToClientChannelRequest(req)
	if err != nil {
		return perror.New(perror.InvalidArgument, err.Error())
	}

	publishReq := pubsub.PublishRequest{Channel: req.clientID, Data: req.event}

	messagesSent.WithLabelValues(req.eventName).Inc()

	return c.pubSub.Publish(ctx, publishReq)
}

func (c *Service) validateSendEventToClientChannelRequest(req SendEventToClientChannelRequest) error {
	if req.clientID == "" {
		return fmt.Errorf("client ID is empty")
	}
	if req.event == nil {
		return fmt.Errorf("event is empty")
	}
	if req.eventName == "" {
		return fmt.Errorf("event name is empty")
	}
	return nil
}

// PublishToClientWithDevice publishes to the client with device
func (c *Service) PublishToClientWithDevice(ctx context.Context, req SendEventToClientDeviceChannelRequest) error {
	logger.Ctx(ctx).Infow("publishing to client with device")
	if !c.config.EnableDeviceSupport {
		return perror.New(perror.FailedPrecondition, "device support disabled")
	}

	err := c.validateSendEventToClientDeviceChannelRequest(req)
	if err != nil {
		return perror.New(perror.InvalidArgument, err.Error())
	}

	publishReq := pubsub.PublishRequest{Channel: fmt.Sprintf("%s--%s", req.clientID, req.deviceID), Data: req.event}

	messagesSent.WithLabelValues(req.eventName).Inc()

	return c.pubSub.Publish(ctx, publishReq)
}

func (c *Service) validateSendEventToClientDeviceChannelRequest(req SendEventToClientDeviceChannelRequest) error {
	if req.clientID == "" {
		return fmt.Errorf("client ID is empty")
	}
	if req.event == nil {
		return fmt.Errorf("event is empty")
	}
	if req.eventName == "" {
		return fmt.Errorf("event name is empty")
	}
	if req.deviceID == "" {
		return fmt.Errorf("device ID is empty")
	}
	return nil
}

// PublishToTopic publishes to the topic
func (c *Service) PublishToTopic(ctx context.Context, req SendEventToTopicRequest) error {
	// TODO: add device id support
	logger.Ctx(ctx).Infow("publishing to Topic", "Topic", req.Topic)

	err := c.validateSendEventToTopicRequest(req)
	if err != nil {
		return perror.New(perror.InvalidArgument, err.Error())
	}

	publishReq := pubsub.PublishRequest{Channel: req.Topic, Data: req.Event}

	messagesSent.WithLabelValues(req.EventName).Inc()

	return c.pubSub.Publish(ctx, publishReq)
}

func (c *Service) validateSendEventToTopicRequest(req SendEventToTopicRequest) error {
	if req.Topic == "" {
		return perror.New(perror.InvalidArgument, "Topic is empty")
	}
	if req.Event == nil {
		return perror.New(perror.InvalidArgument, "Event is empty")
	}
	if req.EventName == "" {
		return perror.New(perror.InvalidArgument, "Event name is empty")
	}
	return nil
}

// PublishToTopics publishes to multiple topics in bulk
func (c *Service) PublishToTopics(ctx context.Context, req SendEventToTopicsRequest) error {
	var publishReqList []pubsub.PublishRequest

	logger.Ctx(ctx).Infow("publishing to topics")

	for _, v := range req.requests {
		err := c.validateSendEventToTopicRequest(v)
		if err != nil {
			return perror.New(perror.InvalidArgument, err.Error())
		}
		publishReq := pubsub.PublishRequest{Data: v.Event, Channel: v.Topic}
		publishReqList = append(publishReqList, publishReq)
		messagesSent.WithLabelValues(v.EventName).Inc()
	}

	return c.pubSub.PublishBulk(ctx, publishReqList)
}

// AsyncClientSubscribe to the client
func (c *Service) AsyncClientSubscribe(ctx context.Context, clientID string, device *Device) (*subscription.Subscription, error) {
	logger.Ctx(ctx).Infow("subscribing to client", "clientID", clientID)
	var clientSubscription *subscription.Subscription
	var err error

	if clientID == "" {
		return nil, perror.New(perror.InvalidArgument, "client ID is empty")

	}

	clientSubscription, err = c.pubSub.AsyncSubscribe(ctx, clientID)
	if err != nil {
		return nil, err
	}

	if c.config.EnableDeviceSupport && device != nil {
		err = c.pubSub.AddSubscription(ctx, fmt.Sprintf("%s--%s", clientID, device.ID), clientSubscription)
		if err != nil {
			return nil, err
		}
		v, err := json.Marshal(device.Attributes)
		err = c.kv.Store(ctx, clientID, device.ID, string(v))
		if err != nil {
			logger.Ctx(ctx).Errorf("error in storing device details %+v", err)
		}
	}

	if c.config.SendTestPayload {
		go c.triggerTestPayloadToClient(ctx, clientID)
		if device != nil {
			go c.triggerTestPayloadToClientWithDevice(ctx, clientID, device.ID)
		}
	}

	connectedClients.Inc()
	c.sessionStartTime = time.Now()
	return clientSubscription, nil
}

// TopicSubscribe to the topic
func (c *Service) TopicSubscribe(ctx context.Context, topic string, clientSubscription *subscription.Subscription) error {
	logger.Ctx(ctx).Infow("subscribing", "topic", topic)
	err := c.pubSub.AddSubscription(ctx, topic, clientSubscription)
	if err != nil {
		return err
	}
	return nil
}

// TopicUnsubscribe to unsubscribe from a topic
func (c *Service) TopicUnsubscribe(ctx context.Context, topic string, clientSubscription *subscription.Subscription) error {
	logger.Ctx(ctx).Debugw("un-subscribing", "topic", topic)
	err := c.pubSub.RemoveSubscription(ctx, topic, clientSubscription)
	if err != nil {
		return err
	}
	return nil
}

// ClientUnsubscribe unsubscribes a client
func (c *Service) ClientUnsubscribe(ctx context.Context, clientID string, subscription *subscription.Subscription, device *Device) error {
	if c.config.EnableDeviceSupport {
		err := c.kv.Delete(ctx, clientID, device.ID)
		if err != nil {
			logger.Ctx(ctx).Errorf("error in deleting device details %+v", err.Error())
		}
	}
	connectedClients.Dec()
	sessionDuration.Observe(time.Since(c.sessionStartTime).Seconds())
	return c.pubSub.Unsubscribe(ctx, subscription)
}

func (c *Service) triggerTestPayloadToClient(ctx context.Context, clientID string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_ = c.PublishToClient(ctx, getDummySendEventToClientRequest(clientID))
		}
		time.Sleep(10 * time.Second)
	}
}

func (c *Service) triggerTestPayloadToClientWithDevice(ctx context.Context, clientID string, deviceID string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_ = c.PublishToClientWithDevice(ctx, getDummySendEventToClientDeviceRequest(clientID, deviceID))
		}
		time.Sleep(10 * time.Second)
	}
}

// TriggerTestPayloadToTopic triggers test payload to topic
func (c *Service) TriggerTestPayloadToTopic(ctx context.Context, topic string) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_ = c.PublishToTopic(ctx, getDummySendEventToTopicRequest(topic))
		}
		time.Sleep(10 * time.Second)
	}
}

// IsDeviceValidationMessage checks if the message is device validation message
func (c *Service) IsDeviceValidationMessage(eventName string) bool {
	if eventName == DeviceValidation {
		return true
	}
	return false
}

// ConfirmEventReceipt is just used for instrumentation
func (c *Service) ConfirmEventReceipt(ctx context.Context, eventName string) {
	messagesReceived.WithLabelValues(eventName).Inc()
}
