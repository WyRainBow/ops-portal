package cache

import (
	"context"
	"testing"
	"time"
)

func TestMemoryCacheSetAndGet(t *testing.T) {
	c := NewMemoryCache()
	ctx := context.Background()

	key := "test_key"
	value := "test_value"

	// Set value
	err := c.Set(ctx, key, value, 5*time.Minute)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get value
	var result string
	err = c.Get(ctx, key, &result)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if result != value {
		t.Errorf("Expected value '%s', got '%s'", value, result)
	}
}

func TestMemoryCacheGetNonExistent(t *testing.T) {
	c := NewMemoryCache()
	ctx := context.Background()

	var result string
	err := c.Get(ctx, "non_existent_key", &result)

	if err != ErrCacheKeyNotFound {
		t.Errorf("Expected ErrCacheKeyNotFound, got %v", err)
	}
}

func TestMemoryCacheDelete(t *testing.T) {
	c := NewMemoryCache()
	ctx := context.Background()

	key := "delete_test"
	value := "delete_value"

	// Set value
	c.Set(ctx, key, value, 5*time.Minute)

	// Delete value
	err := c.Delete(ctx, key)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deleted
	var result string
	err = c.Get(ctx, key, &result)
	if err != ErrCacheKeyNotFound {
		t.Errorf("Expected ErrCacheKeyNotFound after delete, got %v", err)
	}
}

func TestMemoryCacheExists(t *testing.T) {
	c := NewMemoryCache()
	ctx := context.Background()

	key := "exists_test"

	// Check non-existent key
	exists, err := c.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Error("Expected exists to be false for non-existent key")
	}

	// Set value
	c.Set(ctx, key, "value", 5*time.Minute)

	// Check existing key
	exists, err = c.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("Expected exists to be true for existing key")
	}
}

func TestMemoryCacheExpiration(t *testing.T) {
	c := NewMemoryCache()
	ctx := context.Background()

	key := "expire_test"
	value := "expire_value"

	// Set with very short TTL
	err := c.Set(ctx, key, value, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Wait for expiration
	time.Sleep(50 * time.Millisecond)

	// Should be expired - returns either ErrCacheKeyExpired or ErrCacheKeyNotFound
	var result string
	err = c.Get(ctx, key, &result)
	if err != ErrCacheKeyNotFound && err != ErrCacheKeyExpired {
		t.Errorf("Expected ErrCacheKeyNotFound or ErrCacheKeyExpired after expiration, got %v", err)
	}
}

func TestMemoryCacheClear(t *testing.T) {
	c := NewMemoryCache()
	ctx := context.Background()

	// Set multiple values
	c.Set(ctx, "key1", "value1", 5*time.Minute)
	c.Set(ctx, "key2", "value2", 5*time.Minute)
	c.Set(ctx, "key3", "value3", 5*time.Minute)

	// Clear all
	err := c.Clear(ctx)
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	// Verify all cleared
	keys := []string{"key1", "key2", "key3"}
	for _, key := range keys {
		var result string
		err = c.Get(ctx, key, &result)
		if err != ErrCacheKeyNotFound {
			t.Errorf("Expected ErrCacheKeyNotFound for key %s after clear, got %v", key, err)
		}
	}
}

func TestGlobalCache(t *testing.T) {
	// Get global cache - should be singleton
	c1 := Global()
	c2 := Global()

	if c1 != c2 {
		t.Error("Global() should return the same instance")
	}
}
