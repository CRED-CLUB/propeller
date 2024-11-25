package natspkg

import (
	"context"

	"github.com/CRED-CLUB/propeller/internal/perror"
	"github.com/CRED-CLUB/propeller/pkg/logger"
	"github.com/nats-io/nats.go/jetstream"
)

// KV struct for NATS
type KV struct {
	kv jetstream.KeyValue
}

// CreateKeyValue for NATS
func (j *JetStream) CreateKeyValue(ctx context.Context, bucket string) (*KV, error) {
	kv, err := j.js.CreateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket:       bucket,
		Description:  "",
		MaxValueSize: 0,
		History:      0,
		TTL:          0,
		MaxBytes:     0,
		Storage:      0,
		Replicas:     0,
		Placement:    nil,
		RePublish:    nil,
		Mirror:       nil,
		Sources:      nil,
	})
	if err != nil {
		pErr := perror.Newf(perror.Internal, "error creating nats key value %v", err)
		logger.Ctx(ctx).Error(pErr.Error())
		return nil, pErr
	}
	return &KV{kv: kv}, nil
}

// Put a key with value
func (kv *KV) Put(ctx context.Context, key string, value []byte) error {
	_, err := kv.kv.Put(ctx, key, value)
	return err
}

// Get value for a key
func (kv *KV) Get(ctx context.Context, key string) ([]byte, error) {
	entry, err := kv.kv.Get(ctx, key)
	if err != nil {
		pErr := perror.Newf(perror.Internal, "error getting nats key %v", err)
		return nil, pErr
	}
	return entry.Value(), nil
}

// Delete a key
func (kv *KV) Delete(ctx context.Context, key string) error {
	err := kv.kv.Delete(ctx, key)
	if err != nil {
		pErr := perror.Newf(perror.Internal, "error deleting nats key %v", err)
		logger.Ctx(ctx).Error(pErr.Error())
		return pErr
	}
	return nil
}
