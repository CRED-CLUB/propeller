package grpcserver

// Config holds grpc server config
type Config struct {
	Address                  string
	PingIntervalInSec        int
	PingResponseTimeoutInSec int
    EnableReflection         bool
}
