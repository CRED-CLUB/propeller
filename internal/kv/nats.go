package kv

import (
	"context"

	natsclient "github.com/CRED-CLUB/propeller/pkg/broker/nats"
)

// Nats ...
type Nats struct {
	conn *natsclient.Client
}

// NewNats returns NATS kv client
func NewNats(conn *natsclient.Client) IKV {
	return &Nats{conn}
}

// Store key with values
func (n *Nats) Store(ctx context.Context, key string, field string, attrs map[string]string) error {
	//TODO: Implement
	return nil
}

// Load values for a key
func (n *Nats) Load(ctx context.Context, key string) (map[string]string, error) {
	//TODO: Implement
	return nil, nil
}

// Delete values for a key
func (n *Nats) Delete(ctx context.Context, key string, fields ...string) error {
	//TODO: Implement
	return nil
}
