package kv

import (
	"bytes"
	"context"
	"encoding/gob"

	"github.com/CRED-CLUB/propeller/internal/perror"
	"github.com/CRED-CLUB/propeller/pkg/logger"
)

// NatsKVClient defines the interface for NATS KV operations
type NatsKVClient interface {
	Put(ctx context.Context, key string, value []byte) error
	Get(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
}

// Nats ...
type Nats struct {
	kvClient NatsKVClient
}

// NewNats returns a new Nats instance that implements IKV
func NewNats(kvClient NatsKVClient) (IKV, error) {
	if kvClient == nil {
		return nil, perror.New(perror.Internal, "kvClient is required")
	}
	return &Nats{kvClient}, nil
}

// Store key with values
func (n *Nats) Store(ctx context.Context, key string, field string, attrs string) error {
	// TODO: fix overwriting of values here
	mapToStore := map[string]string{field: attrs}
	existingMap, _ := n.Load(ctx, key)

	mergedMap := make(map[string]string)

	// Iterate over existingMap and add all key-value pairs to the mergedMap
	for key, value := range existingMap {
		mergedMap[key] = value
	}

	// Iterate over mapToStore and update or add key-value pairs to the mergedMap
	for key, value := range mapToStore {
		mergedMap[key] = value // Overwrite existing values or add new ones
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(mergedMap)
	if err != nil {
		pErr := perror.Newf(perror.Internal, "error encoding key: %v", err)
		logger.Ctx(ctx).Error(pErr.Error())
		return pErr
	}
	return n.kvClient.Put(ctx, key, buf.Bytes())
}

// Load values for a key
func (n *Nats) Load(ctx context.Context, key string) (map[string]string, error) {
	b, _ := n.kvClient.Get(ctx, key)
	buffer := bytes.NewBuffer(b)
	var attrs map[string]string
	decoder := gob.NewDecoder(buffer)
	err := decoder.Decode(&attrs)
	if err != nil {
		return nil, err
	}
	return attrs, nil
}

// Delete values for a key
func (n *Nats) Delete(ctx context.Context, key string, fields ...string) error {
	existingMap, _ := n.Load(ctx, key)
	for _, field := range fields {
		delete(existingMap, field)
	}
	err := n.kvClient.Delete(ctx, key)
	if err != nil {
		return err
	}
	for k, v := range existingMap {
		err = n.Store(ctx, key, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}
