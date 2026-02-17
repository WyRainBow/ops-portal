package cache

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/WyRainBow/ops-portal/internal/ai/errors"
)

// Cache is the interface for cache implementations.
type Cache interface {
	Get(ctx context.Context, key string, dest any) error
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
	Exists(ctx context.Context, key string) (bool, error)
}

// MemoryCache is an in-memory cache implementation.
type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]*cacheItem
}

type cacheItem struct {
	value      []byte
	expiration time.Time
}

// NewMemoryCache creates a new in-memory cache.
func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		items: make(map[string]*cacheItem),
	}
	// Start expiration goroutine
	go cache.cleanupLoop()
	return cache
}

// Get retrieves a value from the cache.
func (c *MemoryCache) Get(ctx context.Context, key string, dest any) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.items[key]
	if !ok {
		return ErrCacheKeyNotFound
	}

	// Check expiration
	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		return ErrCacheKeyExpired
	}

	// Deserialize
	if err := json.Unmarshal(item.value, dest); err != nil {
		return err
	}

	return nil
}

// Set stores a value in the cache.
func (c *MemoryCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Serialize
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	var expiration time.Time
	if ttl > 0 {
		expiration = time.Now().Add(ttl)
	}

	c.items[key] = &cacheItem{
		value:      data,
		expiration: expiration,
	}

	return nil
}

// Delete removes a value from the cache.
func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
	return nil
}

// Clear removes all items from the cache.
func (c *MemoryCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*cacheItem)
	return nil
}

// Exists checks if a key exists in the cache.
func (c *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, ok := c.items[key]
	if !ok {
		return false, nil
	}

	// Check expiration
	if !item.expiration.IsZero() && time.Now().After(item.expiration) {
		return false, nil
	}

	return true, nil
}

// cleanupLoop periodically removes expired items.
func (c *MemoryCache) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.items {
			if !item.expiration.IsZero() && now.After(item.expiration) {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

// Errors
var (
	ErrCacheKeyNotFound = &CacheError{Code: "KEY_NOT_FOUND", Message: "cache key not found"}
	ErrCacheKeyExpired  = &CacheError{Code: "KEY_EXPIRED", Message: "cache key expired"}
)

// CacheError represents a cache error.
type CacheError struct {
	Code    string
	Message string
}

func (e *CacheError) Error() string {
	return e.Message
}

// Global cache instance.
var globalCache Cache = NewMemoryCache()

// Global returns the global cache instance.
func Global() Cache {
	return globalCache
}

// SetGlobal sets a custom global cache (e.g., Redis).
func SetGlobal(cache Cache) {
	globalCache = cache
}

// GetOrSet is a helper that gets a value or executes the function to set it.
func GetOrSet[T any](ctx context.Context, cache Cache, key string, ttl time.Duration, fn func() (T, error)) (T, error) {
	var zero T

	// Try to get from cache
	var cached T
	err := cache.Get(ctx, key, &cached)
	if err == nil {
		return cached, nil
	}

	// Not in cache, execute function
	result, err := fn()
	if err != nil {
		return zero, err
	}

	// Store in cache
	if cacheErr := cache.Set(ctx, key, result, ttl); cacheErr != nil {
		errors.Warn("cache", "failed to set cache key: "+key)
	}

	return result, nil
}

// LLMCache provides semantic caching for LLM responses.
type LLMCache struct {
	cache               Cache
	similarityThreshold float64
}

// NewLLMCache creates a new LLM cache.
func NewLLMCache(cache Cache, threshold float64) *LLMCache {
	return &LLMCache{
		cache:               cache,
		similarityThreshold: threshold,
	}
}

// GetKey generates a cache key for LLM requests.
func (c *LLMCache) GetKey(model, prompt string) string {
	// Simple hash-based key generation
	// In production, use a proper hash function
	return "llm:" + model + ":" + prompt
}

// ToolCache provides caching for tool results.
type ToolCache struct {
	cache      Cache
	defaultTTL time.Duration
}

// NewToolCache creates a new tool cache.
func NewToolCache(cache Cache, defaultTTL time.Duration) *ToolCache {
	return &ToolCache{
		cache:      cache,
		defaultTTL: defaultTTL,
	}
}

// GetKey generates a cache key for tool calls.
func (c *ToolCache) GetKey(toolName string, input any) string {
	// Serialize input to create key
	data, _ := json.Marshal(input)
	return "tool:" + toolName + ":" + string(data)
}

// Get retrieves a cached tool result.
func (c *ToolCache) Get(ctx context.Context, toolName string, input any, result any) error {
	key := c.GetKey(toolName, input)
	return c.cache.Get(ctx, key, result)
}

// Set stores a tool result in cache.
func (c *ToolCache) Set(ctx context.Context, toolName string, input any, result any, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = c.defaultTTL
	}
	key := c.GetKey(toolName, input)
	return c.cache.Set(ctx, key, result, ttl)
}

// Default TTL values.
const (
	DefaultLLMCacheTTL  = 1 * time.Hour
	DefaultToolCacheTTL = 5 * time.Minute
	DefaultRAGCacheTTL  = 10 * time.Minute
)
