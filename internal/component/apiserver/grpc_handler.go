package apiserver

import (
	"context"
	"fmt"
	"io"

	"github.com/CRED-CLUB/propeller/internal/config"
	"github.com/CRED-CLUB/propeller/internal/perror"
	"github.com/CRED-CLUB/propeller/internal/push"
	"github.com/CRED-CLUB/propeller/pkg/logger"
	pushv1 "github.com/CRED-CLUB/propeller/rpc/push/v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// NewPushServer returns the web server for push
func NewPushServer(svc *push.Service, config config.Config) *PushServer {
	return &PushServer{svc: svc, conf: config}
}

// PushServer implements push web component
type PushServer struct {
	pushv1.UnimplementedPushServiceServer
	svc  *push.Service
	conf config.Config
}

// Channel to the push server and receive streaming response
func (ps *PushServer) Channel(srv pushv1.PushService_ChannelServer) error {

	clientID, device, pErr := ps.validateAndGetClientAndDeviceDetails(srv.Context())
	if pErr != nil {
		return perror.ToGRPCError(pErr)
	}

	// prepare contextual logger with  fields
	derivedCtx := context.WithValue(srv.Context(), logger.CtxKeyType("meta"), map[string]string{
		"clientID": clientID,
	})
	loggerCtx := context.WithValue(derivedCtx, logger.CtxKey, logger.WithContext(derivedCtx, []logger.CtxKeyType{"meta"}))

	clientSubscription, err := ps.svc.AsyncClientSubscribe(loggerCtx, clientID, device)
	if err != nil {
		logger.Ctx(loggerCtx).Info("error in subscribing to client", "error", err.Error())
		return perror.ToGRPCError(err)
	}

	rc := make(chan *pushv1.ChannelRequest)
	go receiveLoop(loggerCtx, rc, srv)

	err = srv.Send(&pushv1.ChannelResponse{Response: &pushv1.ChannelResponse_ConnectAck{
		ConnectAck: &pushv1.ConnectAck{
			Status: &pushv1.ResponseStatus{
				Success:   true,
				ErrorCode: "",
				Message:   nil,
				ErrorType: "",
			},
		},
	}})
	if err != nil {
		logger.Ctx(loggerCtx).Errorw("error in sending response ack", "error", err.Error())
	}

	for {
		select {
		case <-srv.Context().Done():
			logger.Ctx(loggerCtx).Debugw("client closed connection")
			// create new context as input one is already cancelled
			_ = ps.svc.ClientUnsubscribe(context.WithoutCancel(loggerCtx), clientID, clientSubscription, device)
			return nil
		case err := <-clientSubscription.ErrChan:
			logger.Ctx(loggerCtx).Errorw("error in subscriber", "error", err.Error())
		case topicEventReceived := <-clientSubscription.TopicEventChan:
			protoEvent := &pushv1.Event{}
			err := proto.Unmarshal(topicEventReceived.Event, protoEvent)
			if err != nil {
				logger.Ctx(loggerCtx).Errorf("error in converting to proto %v", err)
				break
			}
			if ps.svc.IsDeviceValidationMessage(protoEvent.Name) {
				e := pushv1.Event{
					Name:       fmt.Sprintf("%s#%s#%s", clientID, protoEvent.GetData().Value, "resp"),
					FormatType: 0,
					Data: &anypb.Any{
						TypeUrl: "",
						Value:   protoEvent.GetData().Value,
					},
				}
				eventBytes, err := proto.Marshal(&e)
				if err != nil {
					logger.Ctx(loggerCtx).Errorw("error in marshaling event", "error", err.Error())
					break
				}
				err = ps.svc.PublishToTopic(loggerCtx, push.SendEventToTopicRequest{
					Topic:     fmt.Sprintf("%s#%s#%s", clientID, protoEvent.GetData().Value, "resp"),
					EventName: protoEvent.Name,
					Event:     eventBytes,
				})
				if err != nil {
					logger.Ctx(loggerCtx).Errorw("error in publish device validation resp", "error", err.Error())
				}
				break
			}
			err = srv.Send(&pushv1.ChannelResponse{Response: &pushv1.ChannelResponse_ChannelEvent{
				ChannelEvent: &pushv1.ChannelEvent{Event: protoEvent, Topic: topicEventReceived.Topic},
			}})
			if err != nil {
				logger.Ctx(loggerCtx).Errorw("error in send", "error", err.Error())
			}
			ps.svc.ConfirmEventReceipt(loggerCtx, protoEvent.Name)
			logger.Ctx(loggerCtx).Debugw("sent event", "eventName", protoEvent.GetName())
		case req := <-rc:
			logger.Ctx(loggerCtx).Infow("received from client", "req", req)
			ps.svc.HandleReceivedPayload(loggerCtx, req, clientSubscription)
		}
	}
}

// SendEventToClientChannel sends event to a client
func (ps *PushServer) SendEventToClientChannel(ctx context.Context, req *pushv1.SendEventToClientChannelRequest) (*pushv1.SendEventToClientChannelResponse, error) {
	// prepare contextual logger with  fields
	derivedCtx := context.WithValue(ctx, logger.CtxKeyType("meta"), map[string]string{
		"clientId":  req.ClientId,
		"eventName": req.GetEvent().GetName(),
	})
	loggerCtx := context.WithValue(derivedCtx, logger.CtxKey, logger.WithContext(derivedCtx, []logger.CtxKeyType{"meta"}))

	reqModel := push.SendEventToClientChannelRequest{}

	err := reqModel.PopulateFromProto(loggerCtx, req)
	if err != nil {
		return nil, perror.ToGRPCError(err)
	}

	err = ps.svc.PublishToClient(loggerCtx, reqModel)
	if err != nil {
		return nil, perror.ToGRPCError(err)
	}

	return &pushv1.SendEventToClientChannelResponse{}, nil
}

// SendEventToClientDeviceChannel sends event to a client with device
func (ps *PushServer) SendEventToClientDeviceChannel(ctx context.Context, req *pushv1.SendEventToClientDeviceChannelRequest) (*pushv1.SendEventToClientDeviceChannelResponse, error) {
	// prepare contextual logger with  fields
	derivedCtx := context.WithValue(ctx, logger.CtxKeyType("meta"), map[string]string{
		"clientId":  req.ClientId,
		"deviceId":  req.DeviceId,
		"eventName": req.GetEvent().GetName(),
	})
	loggerCtx := context.WithValue(derivedCtx, logger.CtxKey, logger.WithContext(derivedCtx, []logger.CtxKeyType{"meta"}))

	reqModel := push.SendEventToClientDeviceChannelRequest{}

	err := reqModel.PopulateFromProto(loggerCtx, req)
	if err != nil {
		return nil, perror.ToGRPCError(err)
	}

	err = ps.svc.PublishToClientWithDevice(loggerCtx, reqModel)
	if err != nil {
		return nil, perror.ToGRPCError(err)
	}

	return &pushv1.SendEventToClientDeviceChannelResponse{}, nil

}

// SendEventToTopic sends event to a topic
func (ps *PushServer) SendEventToTopic(ctx context.Context, req *pushv1.SendEventToTopicRequest) (*pushv1.SendEventToTopicResponse, error) {
	// prepare contextual logger with  fields
	derivedCtx := context.WithValue(ctx, logger.CtxKeyType("meta"), map[string]string{
		"topic":     req.GetTopic(),
		"eventName": req.GetEvent().GetName(),
	})
	loggerCtx := context.WithValue(derivedCtx, logger.CtxKey, logger.WithContext(derivedCtx, []logger.CtxKeyType{"meta"}))

	reqModel := push.SendEventToTopicRequest{}

	err := reqModel.PopulateFromProto(loggerCtx, req)
	if err != nil {
		return nil, perror.ToGRPCError(err)
	}

	err = ps.svc.PublishToTopic(loggerCtx, reqModel)
	if err != nil {
		return nil, perror.ToGRPCError(err)
	}

	return &pushv1.SendEventToTopicResponse{}, nil
}

// SendEventToTopics sends event to multiple topics
func (ps *PushServer) SendEventToTopics(ctx context.Context, req *pushv1.SendEventToTopicsRequest) (*pushv1.SendEventToTopicsResponse, error) {
	// prepare contextual logger with  fields
	var eventNames, topics []string
	for _, v := range req.Requests {
		topics = append(topics, v.Topic)
		eventNames = append(eventNames, v.GetEvent().GetName())
	}
	derivedCtx := context.WithValue(ctx, logger.CtxKeyType("meta"), map[string][]string{
		"topics":     topics,
		"eventNames": eventNames,
	})
	loggerCtx := context.WithValue(derivedCtx, logger.CtxKey, logger.WithContext(derivedCtx, []logger.CtxKeyType{"meta"}))

	reqModel := push.SendEventToTopicsRequest{}

	err := reqModel.PopulateFromProto(loggerCtx, req)
	if err != nil {
		return nil, perror.ToGRPCError(err)
	}

	err = ps.svc.PublishToTopics(loggerCtx, reqModel)
	if err != nil {
		return nil, perror.ToGRPCError(err)
	}

	return &pushv1.SendEventToTopicsResponse{}, nil
}

// GetClientActiveDevices returns currently online devices for a client
func (ps *PushServer) GetClientActiveDevices(ctx context.Context, req *pushv1.GetClientActiveDevicesRequest) (*pushv1.GetClientActiveDevicesResponse, error) {
	// prepare contextual logger with  fields
	derivedCtx := context.WithValue(ctx, logger.CtxKeyType("meta"), map[string]string{
		"clientId": req.ClientId,
	})
	loggerCtx := context.WithValue(derivedCtx, logger.CtxKey, logger.WithContext(derivedCtx, []logger.CtxKeyType{"meta"}))

	reqModel := push.GetClientActiveDevicesRequest{}

	err := reqModel.PopulateFromProto(loggerCtx, req)
	if err != nil {
		return nil, perror.ToGRPCError(err)
	}

	r, err := ps.svc.GetClientActiveDevices(loggerCtx, reqModel)
	if err != nil {
		return nil, perror.ToGRPCError(err)
	}
	var devices []*pushv1.Device
	var isClientOnline bool
	for i := range r {
		devices = append(devices, r[i].ToProto())
		isClientOnline = true
	}

	return &pushv1.GetClientActiveDevicesResponse{
		Status: &pushv1.ResponseStatus{
			Success:   true,
			ErrorCode: "",
			Message:   nil,
			ErrorType: "",
		},
		IsClientOnline: isClientOnline,
		Devices:        devices,
	}, nil
}

func receiveLoop(ctx context.Context, rc chan *pushv1.ChannelRequest, srv pushv1.PushService_ChannelServer) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			req, err := srv.Recv()
			if err != nil {
				if err == io.EOF {
					continue
				}
				logger.Ctx(ctx).Debugw("error in recv", "error", err.Error())
				return
			}
			rc <- req
		}
	}
}

func (ps *PushServer) validateAndGetClientAndDeviceDetails(ctx context.Context) (clientID string, device *push.Device, pErr error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		pErr = perror.New(perror.InvalidArgument, "couldn't parse incoming context metadata")
		logger.Ctx(ctx).Error(pErr)
		return "", nil, pErr
	}

	if len(md.Get(ps.conf.ClientHeader)) <= 0 {
		pErr = perror.New(perror.InvalidArgument, "couldn't parse client header")
		logger.Ctx(ctx).Error(pErr)
		return "", nil, pErr
	}

	clientID = md.Get(ps.conf.ClientHeader)[0]
	device = &push.Device{}

	if ps.conf.EnableDeviceSupport {
		attrs := make(map[string]string)
		for i := range ps.conf.DeviceAttributeHeaders {
			headerKey := ps.conf.DeviceAttributeHeaders[i]
			headerValue := md.Get(headerKey)
			if len(headerValue) > 0 {
				attrs[headerKey] = headerValue[0]
			}
		}
		if len(md.Get(ps.conf.DeviceHeader)) <= 0 {
			pErr = perror.New(perror.InvalidArgument, "couldn't parse device header")
			logger.Ctx(ctx).Error(pErr)
			return "", nil, pErr
		}
		device.ID = md.Get(ps.conf.DeviceHeader)[0]
		device.Attributes = attrs
	}

	return clientID, device, nil
}
