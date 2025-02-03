package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

// CacheMethods defines the interface for Memcached operations.
type CacheMethods interface {
	Set(ctx context.Context, key string, value []byte, expiration int32) error
	Get(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
	Ping(ctx context.Context, maxRetries int) error
}

// MemcachedClient holds the Memcached client instance.
type MemcachedClient struct {
	client *memcache.Client
}

// NewMemcachedClient initializes a new Memcached connection.
func NewMemcachedClient(host string, port int) (*MemcachedClient, error) {
	address := fmt.Sprintf("%s:%d", host, port)
	client := memcache.New(address)

	// Simple ping to check if Memcached is available
	err := client.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Memcached: %w", err)
	}

	slog.Info("Successfully connected to Memcached")
	return &MemcachedClient{client: client}, nil
}

// Set stores a key-value pair in Memcached with an expiration time.
func (mc *MemcachedClient) Set(ctx context.Context, key string, value []byte, expiration int32) error {
	return mc.client.Set(&memcache.Item{
		Key:        key,
		Value:      value,
		Expiration: expiration, // Time in seconds (0 = never expires)
	})
}

// Get retrieves a value from Memcached by key.
func (mc *MemcachedClient) Get(ctx context.Context, key string) ([]byte, error) {
	item, err := mc.client.Get(key)
	if err != nil {
		if err == memcache.ErrCacheMiss {
			slog.Warn("Cache miss for key", "key", key)
			return nil, nil
		}
		return nil, err
	}
	return item.Value, nil
}

// Delete removes a key-value pair from Memcached.
func (mc *MemcachedClient) Delete(ctx context.Context, key string) error {
	return mc.client.Delete(key)
}

// Ping checks if the Memcached connection is alive with retries.
func (mc *MemcachedClient) Ping(ctx context.Context, maxRetries int) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err = mc.client.Ping()
		if err == nil {
			slog.Info("Successfully pinged Memcached")
			return nil
		}

		slog.Warn(
			"Memcached ping failed, retrying...",
			"attempt", i+1,
			"remainingRetries", maxRetries-i-1,
			"error", err,
		)
		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("failed to ping Memcached after %d retries: %w", maxRetries, err)
}
