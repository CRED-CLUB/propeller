---
title: Client Testing

layout: default
nav_order: 3
---
# Client Setup

Any gRPC client can be used to test `propeller` (eg [`Postmman`](https://www.postman.com/)). `propeller` doesn't support server reflection. [`api.proto`](https://github.com/CRED-CLUB/propeller/blob/main/proto/push/v1/api.proto) file can be imported in the client (like Postman).

## `Channel` request
Create a new `gRPC` request.

Set the following metadata:
- `x-user-id`: This should match the `ClientHeader` defined in `propeller.yaml` config.
- `x-device-id`: This should match the `DeviceHeader` defined in `propeller.yaml` config.
- Optionally, metadata headers defined in `DeviceAttributeHeaders` config can be supplied.

On `Invoke`, a `connect_ack` event would be received indicating successful connection with `propeller` backend.

![](https://i.ibb.co/hLm9y8B/Screenshot-2024-11-25-at-8-38-03-AM.png)

## `SendEventToClientChannel` request

Set the `client_id` to the value of `ClientHeader` metadata defined in the `Channel` request.

![](https://i.ibb.co/cwZt3dk/Screenshot-2024-11-25-at-8-39-04-AM.png)
The event is received on the `Channel`

![](https://i.ibb.co/8Ppc6bz/Screenshot-2024-11-25-at-8-40-07-AM.png)
## Topic Subscription from Client

Clients can subscribe to custom topic (eg, `test-topic` hereP)

![](https://i.ibb.co/xLNBgYH/Screenshot-2024-11-25-at-9-12-46-AM.png)

## `SendEventToTopic` request

Events can be sent to custom topics

![](https://i.ibb.co/vPVTp38/Screenshot-2024-11-25-at-8-42-14-AM.png)

The sent even is received on the `Channel`

![](https://i.ibb.co/F3j43Kn/Screenshot-2024-11-25-at-8-42-44-AM.png)

----

