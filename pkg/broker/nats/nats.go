package natspkg

import (
	"context"

	"github.com/CRED-CLUB/propeller/pkg/broker"

	"github.com/CRED-CLUB/propeller/internal/perror"
	"github.com/CRED-CLUB/propeller/pkg/logger"
	"github.com/nats-io/nats.go"
)

// INats is an interface over nats pubsub and jetstream
type INats interface {
	Publish(ctx context.Context, publishRequest PublishRequest) error
	Subscribe(ctx context.Context, channel string) (broker.ISubscription, error)
	UnSubscribe(ctx context.Context, s broker.ISubscription) error
}

// Client holds the connection to NATS
type Client struct {
	conn *nats.Conn
}

// NewClient returns a new client with the connection
func NewClient(ctx context.Context, config Config) (*Client, error) {
	nc, err := nats.Connect(config.URL)
	if err != nil {
		pErr := perror.Newf(perror.Internal, "unable to connect to nats %w", err)
		logger.Ctx(ctx).Error(pErr.Error())
		return nil, pErr
	}
	return &Client{nc}, err
}
