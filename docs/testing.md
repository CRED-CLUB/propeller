---
title: Testing the APIs

layout: default
nav_order: 4
---
# Client Setup

Any gRPC client can be used to test `propeller` (eg [`Postman`](https://www.postman.com/)). `Propeller` doesn't support server reflection (yet). The [`api.proto`](https://github.com/CRED-CLUB/propeller/blob/main/proto/push/v1/api.proto) file can be imported in a client (like Postman).

## `Channel` request
Create a new `gRPC` request.

Set the following metadata headers:
- `ClientHeader`, as defined in `propeller.toml` config. Eg, `x-user-id` here.
- `DeviceHeader`, as defined in `propeller.toml` config. Eg, `x-device-id` here.
- Optionally, `DeviceAttributeHeaders` metadata headers can be set as defined in `propeller.toml`.

On `Invoke`, a `connect_ack` event would be received indicating successful connection with `propeller` backend.

![](https://i.ibb.co/hLm9y8B/Screenshot-2024-11-25-at-8-38-03-AM.png)

## `SendEventToClientChannel` request

Set the `client_id` to the value of `ClientHeader` metadata as also defined in the `Channel` request.

![](https://i.ibb.co/cwZt3dk/Screenshot-2024-11-25-at-8-39-04-AM.png)
The event is received on the `Channel`.

![](https://i.ibb.co/8Ppc6bz/Screenshot-2024-11-25-at-8-40-07-AM.png)
## Custom topic Subscription from the client

Clients can subscribe to custom topics (eg, `test-topic` here)

![](https://i.ibb.co/xLNBgYH/Screenshot-2024-11-25-at-9-12-46-AM.png)

## `SendEventToTopic` request

Events can be sent to custom topics from the backend.

![](https://i.ibb.co/vPVTp38/Screenshot-2024-11-25-at-8-42-14-AM.png)

The sent event is received on the `channel`.

![](https://i.ibb.co/F3j43Kn/Screenshot-2024-11-25-at-8-42-44-AM.png)

----

