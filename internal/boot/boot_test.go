package boot

import (
	"context"
	"testing"

	"github.com/CRED-CLUB/propeller/internal/config"
	"github.com/CRED-CLUB/propeller/internal/grpcserver"
	"github.com/CRED-CLUB/propeller/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestInitLogging(t *testing.T) {
	testConfig := config.Config{
		ServiceName:            "propeller",
		SendTestPayload:        false,
		SendTestPayloadToTopic: false,
		Grpc:                   grpcserver.Config{},
		Logger: logger.Config{
			Type: "prod",
		},
	}
	err := InitLogging(testConfig)
	assert.Nil(t, err)
	assert.NotNil(t, logger.Log)
}

func TestNewContextWithNil(t *testing.T) {
	ctx := NewContext(nil)
	assert.NotNil(t, ctx)
}

func TestNewContextWithCtx(t *testing.T) {
	passedCtx := context.Background()
	ctx := NewContext(passedCtx)
	assert.NotNil(t, ctx)
	assert.Equal(t, passedCtx, ctx)
}
