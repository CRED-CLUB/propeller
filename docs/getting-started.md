---
title: Getting Started
layout: default
nav_order: 2
---
# Getting Started Guide

The easiest way to get started is by pulling and running the docker image. It has the following pre-requisites.

## Create a config file

`/tmp/propeller.yaml`
```yaml
ServiceName = "propeller"
SendTestPayload = false
SendTestPayloadToTopic = false
ClientHeader = "x-user-id"
EnableDeviceSupport = true
DeviceHeader = "x-device-id"
DeviceAttributeHeaders = ["x-os", "x-os-version"]

[broker]
    broker = "redis" # broker = "nats" or "redis"
    persistence = false
    [broker.nats]
        URL = "http://localhost:4222"
        EmbeddedServer = true
    [broker.redis]
        Address = "localhost:6379"
        Password = ""
        TLSEnabled = false
        ClusterModeEnabled = false

[grpc]
    Address = "0.0.0.0:5011"
    PingIntervalInSec = 20
    PingResponseTimeoutInSec = 10

[Logger]
    Type = "dev" # prod or dev

[Features]
    [Features.SampleFeature]
        Enabled = true
        RolloutPercentage = 100
[Http]
    Port = "8081"

```
## Run redis locally
```bash
docker run -d --name redis-stack-server -p 6379:6379 redis/redis-stack-server:latest
```

## Run the docker image

```bash
docker run -it -e PROPELLER_CONFIG_FILE_PATH=/etc -e PROPELLER_BROKER_REDIS_ADDRESS=localhost:6379 -v $(pwd)/tmp:/etc --network="host" quay.io/abhishekvrshny/propeller
```

## Connect the sample client


----

