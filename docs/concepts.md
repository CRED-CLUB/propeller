---
title: Concepts

layout: default
nav_order: 2
---
# Concepts

## Terminology

| Term              | Definition                                                                                                                                                                                                                            |
|-------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **event**         | A message payload to be exchanged                                                                                                                                                                                                     |
| **client**        | A client (generally frontend) to whom an `event` is sent (generally from backend)                                                                                                                                                     |
| **channel**       | A bidirectional connection established between the `Client` and `propeller`                                                                                                                                                           |
| **topic**         | A topic on interest between backend and frontend clients. Default `topic` for a `channel` is the value of `ClientHeader` key as defined in `propeller.toml` config. A `topic` can be custom as well to which `client`'s can subscribe |
| **device**        | A `client`'s device. A `client` can have multiple devices                                                                                                                                                                             |
| **device attributes** | Attributes of a device as defined by `DeviceAttributeHeaders` in `propeller.toml` config. Eg. `x-os`, `x-app-version` etc.                                                                                                            |

----

## APIs and Flows

### Establishing Channel

A `channel` is established when a `client` sends the `channel` request to `propeller`

```protobuf
rpc Channel(stream ChannelRequest) returns (stream ChannelResponse) {}
```
A `channel` is a bi-directional gRPC stream established between the `client` and `propeller`. Different types of messages can be sent and received on this stream by the `client` and `propeller`.

`ChannelRequest` is the request message from the `client` which can be of one of the following types:
1. `ChannelEvent`: Carries the `event` payload.
2. `ChannelEventAck`: To acknowledging the receipt of an `event`.
3. `TopicSubscriptionRequest`: To subscribe to a custom `topic`.
4. `TopicUnsubscriptionRequest`: To un-subscribe to the custom `topic`. 
```protobuf
// ChannelRequest is the channel request holder
message ChannelRequest {
  oneof request {
    // channel_event carries main payload
    ChannelEvent channel_event = 1;

    // channel_event_ack is ack of channel_event
    ChannelEventAck channel_event_ack = 2;

    // topic_subscription_request to subscribe to a topic
    TopicSubscriptionRequest topic_subscription_request = 3;

    // topic_unsubscription_request to unsubscribe to a topic
    TopicUnsubscriptionRequest topic_unsubscription_request = 4;
  }
}
```

`ChannelResponse` is the response sent by `propeller` to the client. In addition to `ChannelEvent` and `ChannelEventAck`, it can be of one of the following types:
1. `ConnectAck`: To acknowledge the `channel` connection request.
2. `TopicSubscriptionRequestAck`: To acknowledge `TopicSubscriptionRequest`.
3. `TopicUnsubscriptionRequestAck`: To acknowledge `TopicUnsubscriptionRequest`.
```
// ChannelResponse is the channel response holder
message ChannelResponse {
  oneof response {
    // connect_ack is the ack for channel connect request
    ConnectAck connect_ack = 1;

    // channel_event carries main payload
    ChannelEvent channel_event = 2;

    // channel_event_ack is ack of channel_event
    ChannelEventAck channel_event_ack = 3;

    // topic_subscription_request_ack is ack of topic_subscription_request
    TopicSubscriptionRequestAck topic_subscription_request_ack = 4;

    // topic_unsubscription_request_ack is ack of topic_unsubscription_request
    TopicUnsubscriptionRequestAck topic_unsubscription_request_ack = 5;
  }
}
```

A `client` is identified by the value of `ClientHeader` metadata passed by the client in headers.

A `device` is identified by the value of `DeviceHeader` metadata passed by the client in headers.

{: .note } 
Authentication is not handled by `propeller`. An authentication middleware or API gateway can be used to inject the defined `ClientHeader` and/or `DeviceHeader`, if required.

### Sending `event` to a `client`

A backend service can send `event` to a `client` with `SendEventToClientChannel` rpc

```protobuf
rpc SendEventToClientChannel(SendEventToClientChannelRequest) returns (SendEventToClientChannelResponse) {}
```

### Sending `event` to a particular `device` of a `client`

If `EnableDeviceSupport` config is enabled, an `event` can be sent to a particular `device` of a `client`. This is useful when a client has `channels` established from multiple `devices`

```protobuf
rpc SendEventToClientDeviceChannel(SendEventToClientDeviceChannelRequest) returns (SendEventToClientDeviceChannelResponse) {}
```

### List all active `devices` for a `client`

If `EnableDeviceSupport` config is enabled, all online devices for a `client` can be listed with their `device attributes` as defined by `DeviceAttributeHeaders` config.

```protobuf
rpc GetClientActiveDevices(GetClientActiveDevicesRequest) returns (GetClientActiveDevicesResponse) {}
```

### Sending `event` to a custom `topic`

Backend services can send `event` to a custom `topic`.
```protobuf
  rpc SendEventToTopic(SendEventToTopicRequest) returns (SendEventToTopicResponse) {}
```

There also exists a rpc to do the same in bulk to multiple `topics`.
```protobuf
  rpc SendEventToTopics(SendEventToTopicsRequest) returns (SendEventToTopicsResponse) {}
```

---
