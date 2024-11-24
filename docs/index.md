---
title: Home
layout: home
nav_order: 1
---
# Propeller 𖣘
Propeller is an *opinionated* platform that enabled realtime and bidirectional communication between frontend and backend clients.

## The Need
While polling and REST-style request-response have been the dominant paradigms for client-server communication, they inherently come with limitations that can impact performance, scalability, and user experience.
1. **Increased Latency**: Polling intervals often dictate the minimum latency for receiving updates, limiting real-time responsiveness.
2. **Resource Consumption**: Each polling request requires server resources to process, even if there's no new data.
3. **Retry Storms**: In case of widespread failures or degradations, a large number of clients may initiate retries simultaneously, overwhelming the server.
4. **Scalability Constraints**: As the number of clients and polling frequency increase, server load can become a significant bottleneck.

## Features
1. Frontend Client can create a persistent channel with the backend.
2. Backend services can send events to the frontend clients.
3. Support for multiple devices for a client.
4. Support for custom topics between frontend and backend.
5. Easy integration with legacy REST based services.

## Building Blocks
Propeller is built on top of the following battle-tested technologies to power realtime experiences:
1. **bi-directional gRPC**: Propeller uses bi-directional streaming gRPC to establish a stream between the client and the server.
2. **Redis and NATS**: Propeller supports Redis and NATS as brokers for the communication.
3. **Golang**: Propeller uses the power of Golang to achieve high number of concurrent clients being connected.

----

