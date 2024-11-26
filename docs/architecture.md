---
title: Architecture

layout: default
nav_order: 5
---
# Overview

`Propeller` acts as a proxy to send events from the backend services to the frontend clients. It is specifically designed to solve for non gRPC based services.

![](https://pic.surf/uda)

# High Level Design

![](https://pic.surf/vy4)

## Components

- propeller server
   - written in golang.
   - supports gRPC bidi streaming and websockets.
  
- broker support
    - redis
        - pubsub
        - streams
    - NATS
        - pubsub
        - jetstream

---
