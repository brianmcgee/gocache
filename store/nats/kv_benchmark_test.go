package nats

import (
	"context"
	"fmt"
	"math"
	"os"
	"testing"
	"time"

	lib_store "github.com/eko/gocache/lib/v4/store"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
)

const (
	_EMPTY_ = ""
)

func runBasicJetStreamServer(b *testing.B) *server.Server {
	b.Helper()
	opts := test.DefaultTestOptions
	opts.Port = -1
	opts.JetStream = true
	return test.RunServer(&opts)
}

func client(b *testing.B, s *server.Server, opts ...nats.Option) *nats.Conn {
	b.Helper()
	nc, err := nats.Connect(s.ClientURL(), opts...)
	if err != nil {
		b.Fatalf("Unexpected error: %v", err)
	}
	return nc
}

func jsClient(b *testing.B, s *server.Server, opts ...nats.Option) (*nats.Conn, nats.JetStreamContext) {
	b.Helper()
	nc := client(b, s, opts...)
	js, err := nc.JetStream(nats.MaxWait(10 * time.Second))
	if err != nil {
		b.Fatalf("Unexpected error getting JetStream context: %v", err)
	}
	return nc, js
}

func shutdownJSServerAndRemoveStorage(b *testing.B, s *server.Server) {
	b.Helper()
	var sd string
	if config := s.JetStreamConfig(); config != nil {
		sd = config.StoreDir
	}
	s.Shutdown()
	if sd != _EMPTY_ {
		if err := os.RemoveAll(sd); err != nil {
			b.Fatalf("Unable to remove storage %q: %v", sd, err)
		}
	}
	s.WaitForShutdown()
}

func BenchmarkNatsKVSet(b *testing.B) {
	ctx := context.Background()

	s := runBasicJetStreamServer(b)
	defer shutdownJSServerAndRemoveStorage(b, s)

	_, js := jsClient(b, s)
	config := &nats.KeyValueConfig{
		Bucket: "benchmark",
	}

	store, err := NewNatsKV(js, config, lib_store.WithExpiration(100*time.Second))
	assert.Nil(b, err)

	for k := 0.; k <= 10; k++ {
		n := int(math.Pow(2, k))
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N*n; i++ {
				key := fmt.Sprintf("test-%d", n)
				value := []byte(fmt.Sprintf("value-%d", n))
				err = store.Set(ctx, key, value)
				assert.Nil(b, err)
			}
		})
	}
}

func BenchmarkNatsKVGet(b *testing.B) {
	ctx := context.Background()

	s := runBasicJetStreamServer(b)
	defer shutdownJSServerAndRemoveStorage(b, s)

	_, js := jsClient(b, s)
	config := &nats.KeyValueConfig{
		Bucket: "benchmark",
	}

	store, err := NewNatsKV(js, config, lib_store.WithExpiration(100*time.Second))
	assert.Nil(b, err)

	key := "test"
	value := []byte("value")

	err = store.Set(ctx, key, value)
	assert.Nil(b, err)

	for k := 0.; k <= 10; k++ {
		n := int(math.Pow(2, k))
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N*n; i++ {
				_, _ = store.Get(ctx, key)
			}
		})
	}
}
