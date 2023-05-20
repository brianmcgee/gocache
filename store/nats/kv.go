package nats

import (
	"context"
	"time"

	libstore "github.com/eko/gocache/lib/v4/store"
	libnats "github.com/nats-io/nats.go"
)

// NatsJetStreamContextInterface represents a nats-io/nats.go JetStreamContext interface
type NatsJetStreamContextInterface interface {
	libnats.JetStreamContext
}

// NatsKeyValueInterface represents a nats-io/nats.go KeyValue interface
type NatsKeyValueInterface interface {
	libnats.KeyValue
}

type NatsKeyValueEntryInterface interface {
	libnats.KeyValueEntry
}

const (
	// NatsKvType represents the storage type as a string value
	NatsKvType = "nats-kv"
)

type NatsKVStore struct {
	kv      libnats.KeyValue
	options *libstore.Options
}

// NewNatsKV creates a new store
func NewNatsKV(
	js libnats.JetStreamContext,
	config *libnats.KeyValueConfig,
	options ...libstore.Option,
) (*NatsKVStore, error) {
	opts := libstore.ApplyOptions(options...)
	config.TTL = opts.Expiration

	kv, err := js.CreateKeyValue(config)
	if err != nil {
		return nil, err
	}
	return &NatsKVStore{
		kv:      kv,
		options: opts,
	}, nil
}

func newNatsKV(kv NatsKeyValueInterface, options ...libstore.Option) *NatsKVStore {
	return &NatsKVStore{
		kv:      kv,
		options: libstore.ApplyOptions(options...),
	}
}

func (s *NatsKVStore) Get(_ context.Context, key any) (any, error) {
	object, err := s.kv.Get(key.(string))
	if err == libnats.ErrKeyNotFound {
		return nil, libstore.NotFoundWithCause(err)
	}
	return object.Value(), err
}

// GetWithTTL returns data stored from a given key and its corresponding TTL
func (s *NatsKVStore) GetWithTTL(_ context.Context, key any) (any, time.Duration, error) {
	object, err := s.kv.Get(key.(string))
	if err == libnats.ErrKeyNotFound {
		return nil, 0, libstore.NotFoundWithCause(err)
	}

	expiresAt := object.Created().Add(s.options.Expiration)
	ttl := expiresAt.Sub(object.Created())

	return object, ttl, err
}

// Set defines data in Nats for given key identifier
func (s *NatsKVStore) Set(ctx context.Context, key any, value any, options ...libstore.Option) error {
	opts := libstore.ApplyOptionsWithDefault(s.options, options...)

	if len(opts.Tags) > 0 {
		return ErrTagsNotSupported
	}

	if opts.Expiration != s.options.Expiration {
		return ErrPerKeyExpirationNotSupported
	}

	_, err := s.kv.Put(key.(string), value.([]byte))
	if err != nil {
		return err
	}

	return nil
}

// Delete removes data from Nats for given key identifier
func (s *NatsKVStore) Delete(_ context.Context, key any) error {
	return s.kv.Delete(key.(string))
}

// Invalidate invalidates some cache data in Nats for given options
func (s *NatsKVStore) Invalidate(_ context.Context, options ...libstore.InvalidateOption) error {
	opts := libstore.ApplyInvalidateOptions(options...)

	if len(opts.Tags) > 0 {
		return ErrTagsNotSupported
	}

	return nil
}

// GetType returns the store type
func (s *NatsKVStore) GetType() string {
	return NatsKvType
}

// Clear resets all data in the store
func (s *NatsKVStore) Clear(_ context.Context) error {
	// TODO is there a more efficient way of doing this?
	keys, err := s.kv.Keys()
	if err != nil {
		return err
	}

	for _, key := range keys {
		err = s.kv.Delete(key)
		if err != nil {
			return err
		}
	}

	return nil
}
