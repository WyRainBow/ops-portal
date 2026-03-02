package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/WyRainBow/ops-portal/internal/ai/errors"
	"github.com/redis/go-redis/v9"
)

// RedisCache is a Redis-based cache implementation.
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache.
func NewRedisCache(addr, password string, db int) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisCache{client: client}, nil
}

// Get retrieves a value from Redis.
func (c *RedisCache) Get(ctx context.Context, key string, dest any) error {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheKeyNotFound
		}
		return err
	}

	// Deserialize
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return err
	}

	return nil
}

// Set stores a value in Redis.
func (c *RedisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	// Serialize
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, ttl).Err()
}

// Delete removes a value from Redis.
func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// Clear removes all keys with the cache prefix.
func (c *RedisCache) Clear(ctx context.Context) error {
	// In production, you might want to use SCAN to avoid blocking
	iter := c.client.Scan(ctx, 0, "*", 0).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			errors.Warn("cache", "failed to delete key during clear: "+iter.Val())
		}
	}
	return nil
}

// Exists checks if a key exists in Redis.
func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.client.Exists(ctx, key).Result()
	return n > 0, err
}

// Close closes the Redis connection.
func (c *RedisCache) Close() error {
	return c.client.Close()
}
