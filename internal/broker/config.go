package broker

import (
	natspkg "github.com/CRED-CLUB/propeller/pkg/broker/nats"
	redispkg "github.com/CRED-CLUB/propeller/pkg/broker/redis"
)

// Config for broker
type Config struct {
	Broker      string
	Persistence bool
	Nats        natspkg.Config
	Redis       redispkg.Config
}
