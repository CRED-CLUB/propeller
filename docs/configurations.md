---
title: Configurations

layout: default
nav_order: 7
---
# Configuration Reference

| Config                          | Values          | Description                                                                                                                |
|---------------------------------|-----------------|----------------------------------------------------------------------------------------------------------------------------|
| SendTestPayload                 | true/false      | Periodically sends a test payload on the channel to the client. Useful for testing.                                        |
| SendTestPayloadToTopic          | true/false      | Periodically sends a test payload when a client subscribes to a topic. Useful for testing.                                 |
| ClientHeader                    | string          | The metadata header key which is used to identify a client.                                                                |
| EnableDeviceSupport             | true/false      | If enabled, a client can create a channel through multiple devices. Backend can send events targetting specific devices of a client. |
| DeviceHeader                    | string          | The metadata header key which is used to identify a device of a client.                                                    |
| DeviceAttributeHeaders          | list of strings | (Optional) metadata header keys for attributes of a devices. They are listed when active devices for a client are fetched from the backend. |
| broker.broker                   | redis/nats      | choice of broker                                                                                                           |
| broker.persistence              | true/false      | If the broker should persist events in case the client is not connected and deliver them later whe the client connects     |
| broker.nats                     | string          | NATS address.                                                                                                              |
| broker.nats.EmbeddedServer      | true/false      | If enabled, an embedded NATS server is started. Useful for testing                                                         |
| broker.redis                    | string          | Redis address.                                                                                                             |
| broker.redis.Password           | string          | Redis password for authentication.                                                                                         |
| broker.redis.TLSEnabled         | true/false      | If TLS should be enabled while connecting to Redis.                                                                        |
| broker.redis.ClusterModeEnabled | true/false      | If cluster mode is enabled on Redis or not. Help with scalability by sharding keys.                                        |
| grpc.Address                    | string          | gRPC server port to bind.                                                                                                  |
| grpc.PingIntervalInSec          | integer         | gRPC keepalive configuration. [Reference](https://grpc.io/docs/guides/keepalive/)                                          |
| grpc.PingResponseTimeoutInSec   | integer         | gRPC keepalive configuration. [Reference](https://grpc.io/docs/guides/keepalive/)                                          |
| Logger.type                     | dev/prod        | prod logger prints logs in JSON while dev logger prints in human-friendly format.                                          |
| Http.Port                       | integer         | HTTP port to bind for websockets and prometheus metrics endpoint                                                           |
| Features.<name>                 | string          | Feature flag for new features.                                                                                             |
| Features.<name>.Enabled         | true/false      | If a feature should be enabled or not.                                                                                     |
| Features.<name>.RolloutPercentage | integer (0-100) | Percentage rollout of the feature.                                                                                         |

# Environment Variable Support

`Propeller` also supports reading configuration parameters like passwords etc. from environment variables. The environment variables have to be prefixed with `PROPELLER_` followed by the `toml` config fields concatenated by `_`. For example, `broker.redis.Password` in `propeller.toml` can be overridden by `PROPELLER_BROKER_REDIS_PASSWORD` environment variable.  

---
