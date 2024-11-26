package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	pushv1 "github.com/CRED-CLUB/propeller/rpc/push/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/anypb"
)

func main() {
	clientID := flag.String("clientID", "test", "client id")
	deviceID := flag.String("deviceID", "device", "device id")
	topic := flag.String("topic", "", "additional topic to subscribe")
	endpoint := flag.String("endpoint", "localhost:5011", "endpoint")
	action := flag.String("action", "send-event", "action to perform: send-event or connect")
	flag.Parse()

	var transportCredentials credentials.TransportCredentials
	transportCredentials = insecure.NewCredentials()

	// connect to the grpc streaming grpcserver
	conn, err := grpc.Dial(*endpoint, grpc.WithTransportCredentials(
		transportCredentials,
	))
	if err != nil {
		log.Fatalf("unable to start grpc grpcserver %s", err.Error())
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {

		}
	}(conn)

	c := pushv1.NewPushServiceClient(conn)
	// Injecting X-USER-ID is not required with auth plugin, as the plugin would inject this
	ctx := metadata.AppendToOutgoingContext(context.Background(), "X-USER-ID", *clientID)
	ctx = metadata.AppendToOutgoingContext(ctx, "x-device-id", *deviceID)
	ctx = metadata.AppendToOutgoingContext(ctx, "x-os", "android")
	ctx = metadata.AppendToOutgoingContext(ctx, "x-os-version", "v1")

	switch *action {
	case "send-event":
		sendMessage(ctx, *clientID, *deviceID, c, *topic)
	case "connect":
		connect(ctx, *clientID, c, *topic)
	case "list-client-devices":
		listClientDevices(ctx, *clientID, c)
	}
}

func sendMessage(ctx context.Context, userID string, deviceID string, c pushv1.PushServiceClient, subTopic string) {
	testData := struct {
		Field string
	}{Field: fmt.Sprintf("time %s", time.Now())}

	jsonBytes, err := json.Marshal(testData)
	if err != nil {
		log.Fatalf("error in encoding %s", err)
	}

	// sleep to make sure client connection is established
	time.Sleep(1 * time.Second)

	if subTopic == "" {
		r, err := c.SendEventToClientChannel(ctx, &pushv1.SendEventToClientChannelRequest{
			ClientId: userID,
			Event: &pushv1.Event{
				Name:       "PAYMENT_SUCCESS",
				FormatType: pushv1.Event_TYPE_JSON_UNSPECIFIED,
				Data: &anypb.Any{
					TypeUrl: "",
					Value:   jsonBytes,
				},
			},
		})
		if err != nil {
			log.Fatalf("error in SendEventToClientChannel %s with response %v", err.Error(), r)
		}
		if deviceID != "" {
			_, err := c.SendEventToClientDeviceChannel(ctx, &pushv1.SendEventToClientDeviceChannelRequest{
				ClientId: userID,
				DeviceId: deviceID,
				Event: &pushv1.Event{
					Name:       "CHECKOUT_PAYMENT_STATUS_DEVICE",
					FormatType: pushv1.Event_TYPE_JSON_UNSPECIFIED,
					Data: &anypb.Any{
						TypeUrl: "",
						Value:   jsonBytes,
					},
				},
			})
			if err != nil {
				log.Fatalf("error in SendEventToClientDeviceChannel %s with response %v", err.Error(), r)
			}
		}
	} else {
		_, err = c.SendEventToTopic(ctx, &pushv1.SendEventToTopicRequest{
			Topic: subTopic,
			Event: &pushv1.Event{
				Name:       subTopic,
				FormatType: pushv1.Event_TYPE_JSON_UNSPECIFIED,
				Data: &anypb.Any{
					TypeUrl: "",
					Value:   jsonBytes,
				},
			},
		})
		if err != nil {
			log.Fatalf("error in SendEventToTopic %s", err.Error())
		}
		_, err = c.SendEventToTopics(ctx, &pushv1.SendEventToTopicsRequest{
			Requests: []*pushv1.SendEventToTopicRequest{
				{
					Topic: subTopic,
					Event: &pushv1.Event{
						Name:       subTopic,
						FormatType: pushv1.Event_TYPE_JSON_UNSPECIFIED,
						Data: &anypb.Any{
							TypeUrl: "",
							Value:   jsonBytes,
						},
					},
				},
				{
					Topic: subTopic,
					Event: &pushv1.Event{
						Name:       subTopic,
						FormatType: pushv1.Event_TYPE_JSON_UNSPECIFIED,
						Data: &anypb.Any{
							TypeUrl: "",
							Value:   jsonBytes,
						},
					},
				},
			},
		})
		if err != nil {
			log.Fatalf("error in SendEventToTopics %s", err.Error())
		}
	}

	log.Printf("sent event request")
}

func connect(ctx context.Context, userID string, c pushv1.PushServiceClient, subTopic string) {
	stream, err := c.Channel(ctx)
	if err != nil {
		log.Fatalf("unable to connect %s", err.Error())
	}

	waitc := make(chan struct{})

	go func() {
		for {
			in, err := stream.Recv()
			log.Printf("sent recv request")
			if err == io.EOF {
				// read done.
				log.Printf("close")
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive a note : %v", err)
			}
			switch in.Response.(type) {
			case *pushv1.ChannelResponse_ChannelEvent:
				log.Printf("event name %s, data %v", in.GetChannelEvent().GetEvent().GetName(), in.GetChannelEvent().GetEvent().GetData())
				req := &pushv1.ChannelRequest{Request: &pushv1.ChannelRequest_ChannelEventAck{ChannelEventAck: &pushv1.ChannelEventAck{}}}
				if err := stream.Send(req); err != nil {
					log.Fatalf("can not send %v", err)
				}
			case *pushv1.ChannelResponse_ConnectAck:
				log.Printf("connect ack received")
			}
		}
	}()
	if subTopic != "" {
		time.Sleep(10 * time.Second)

		log.Printf("subscribing to %s", subTopic)

		req := &pushv1.ChannelRequest{Request: &pushv1.ChannelRequest_TopicSubscriptionRequest{TopicSubscriptionRequest: &pushv1.TopicSubscriptionRequest{Topic: subTopic}}}
		if err := stream.Send(req); err != nil {
			log.Fatalf("cannot subscribe %v", err)
		}

		// sleep before un-subscribing
		time.Sleep(time.Duration(10) * time.Second)

		req = &pushv1.ChannelRequest{Request: &pushv1.ChannelRequest_TopicUnsubscriptionRequest{TopicUnsubscriptionRequest: &pushv1.TopicUnsubscriptionRequest{Topic: subTopic}}}
		if err := stream.Send(req); err != nil {
			log.Fatalf("cannot unsubscribe %v", err)
		}

		log.Printf("un-subscribing")

	}
	<-waitc
	err = stream.CloseSend()
	if err != nil {
		return
	}
}

func listClientDevices(ctx context.Context, userID string, c pushv1.PushServiceClient) {
	res, err := c.GetClientActiveDevices(ctx, &pushv1.GetClientActiveDevicesRequest{ClientId: userID})
	if err != nil {
		log.Fatalf("can not list user devices %v", err)
	}
	log.Printf("user devices %+v", res)
}
