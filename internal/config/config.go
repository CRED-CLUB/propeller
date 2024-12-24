package config

import (
	"github.com/CRED-CLUB/propeller/internal/broker"
	"github.com/CRED-CLUB/propeller/internal/feature"
	"github.com/CRED-CLUB/propeller/internal/grpcserver"
	"github.com/CRED-CLUB/propeller/internal/httpserver"
	"github.com/CRED-CLUB/propeller/pkg/logger"
)

// Config stores application config
type Config struct {
	ServiceName             string
	SendTestPayload         bool
	SendTestPayloadToTopic  bool
	Broker                  broker.Config
	Grpc                    grpcserver.Config
	Features                feature.Config
	Logger                  logger.Config
	DeviceAttributeHeaders  []string
	ClientHeader            string
	DeviceHeader            string
	EnableDeviceSupport     bool
	HTTP                    httpserver.HTTPConfig
	EnableProfilingHandlers bool
}
