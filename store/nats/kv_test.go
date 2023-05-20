package nats

import (
	"context"
	"testing"
	"time"

	lib_store "github.com/eko/gocache/lib/v4/store"
	"github.com/golang/mock/gomock"
	lib_nats "github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
)

func TestNewNatsKV(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	js := NewMockNatsJetStreamContextInterface(ctrl)
	kv := NewMockNatsKeyValueInterface(ctrl)

	config := &lib_nats.KeyValueConfig{
		Bucket: "test",
	}

	js.EXPECT().CreateKeyValue(config).Return(kv, nil)

	// When
	store, err := NewNatsKV(js, config, lib_store.WithExpiration(3*time.Second))

	// Then
	assert.Nil(t, err)
	assert.IsType(t, new(NatsKVStore), store)
	assert.Equal(t, kv, store.kv)
	assert.Equal(t, &lib_store.Options{Expiration: 3 * time.Second}, store.options)
}

func TestNatsKVGet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	entry := NewMockNatsKeyValueEntryInterface(ctrl)
	entry.EXPECT().Value().Return([]byte("my-value"))

	kv := NewMockNatsKeyValueInterface(ctrl)
	kv.EXPECT().Get("my-key").Return(entry, nil)

	store := newNatsKV(kv)

	// When
	value, err := store.Get(ctx, "my-key")

	// Then
	assert.Nil(t, err)
	assert.Equal(t, []byte("my-value"), value)
}
