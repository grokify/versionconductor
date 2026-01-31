package graph

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCache_MemoryOnly(t *testing.T) {
	cache, err := NewCache(CacheConfig{
		MemoryOnly: true,
		TTL:        time.Hour,
	})
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}

	ctx := context.Background()

	// Test set and get
	key := "test-key"
	data := []byte("test-data")

	err = cache.Set(ctx, key, data)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	got, ok := cache.Get(ctx, key)
	if !ok {
		t.Fatal("expected to find cached value")
	}

	if string(got) != string(data) {
		t.Errorf("expected %s, got %s", string(data), string(got))
	}
}

func TestCache_FileBackend(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "versionconductor-test-cache")
	defer os.RemoveAll(dir)

	cache, err := NewCache(CacheConfig{
		Dir: dir,
		TTL: time.Hour,
	})
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}

	ctx := context.Background()

	// Test set and get
	key := "test-file-key"
	data := []byte(`{"test": "data"}`)

	err = cache.Set(ctx, key, data)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Clear memory cache to force file read
	cache.mu.Lock()
	cache.memory = make(map[string]*cacheEntry)
	cache.mu.Unlock()

	got, ok := cache.Get(ctx, key)
	if !ok {
		t.Fatal("expected to find cached value from file")
	}

	if string(got) != string(data) {
		t.Errorf("expected %s, got %s", string(data), string(got))
	}
}

func TestCache_Expiration(t *testing.T) {
	cache, err := NewCache(CacheConfig{
		MemoryOnly: true,
		TTL:        50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}

	ctx := context.Background()
	key := "expire-test"
	data := []byte("expires soon")

	err = cache.Set(ctx, key, data)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Should be found immediately
	_, ok := cache.Get(ctx, key)
	if !ok {
		t.Fatal("expected to find cached value")
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should be expired
	_, ok = cache.Get(ctx, key)
	if ok {
		t.Error("expected cached value to be expired")
	}
}

func TestCache_Delete(t *testing.T) {
	cache, err := NewCache(CacheConfig{
		MemoryOnly: true,
		TTL:        time.Hour,
	})
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}

	ctx := context.Background()
	key := "delete-test"
	data := []byte("to be deleted")

	_ = cache.Set(ctx, key, data)
	_ = cache.Delete(ctx, key)

	_, ok := cache.Get(ctx, key)
	if ok {
		t.Error("expected cached value to be deleted")
	}
}

func TestCache_Clear(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "versionconductor-test-clear")
	defer os.RemoveAll(dir)

	cache, err := NewCache(CacheConfig{
		Dir: dir,
		TTL: time.Hour,
	})
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}

	ctx := context.Background()

	// Add multiple entries
	_ = cache.Set(ctx, "key1", []byte("data1"))
	_ = cache.Set(ctx, "key2", []byte("data2"))
	_ = cache.Set(ctx, "key3", []byte("data3"))

	// Clear cache
	err = cache.Clear(ctx)
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	// Verify all cleared
	for _, key := range []string{"key1", "key2", "key3"} {
		_, ok := cache.Get(ctx, key)
		if ok {
			t.Errorf("expected %s to be cleared", key)
		}
	}
}

func TestCache_Stats(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "versionconductor-test-stats")
	defer os.RemoveAll(dir)

	cache, err := NewCache(CacheConfig{
		Dir: dir,
		TTL: time.Hour,
	})
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}

	ctx := context.Background()

	// Add entries
	_ = cache.Set(ctx, "stat1", []byte("data1"))
	_ = cache.Set(ctx, "stat2", []byte("data2"))

	stats := cache.Stats(ctx)

	if stats.MemoryEntries != 2 {
		t.Errorf("expected 2 memory entries, got %d", stats.MemoryEntries)
	}

	if stats.FileEntries != 2 {
		t.Errorf("expected 2 file entries, got %d", stats.FileEntries)
	}
}

func TestCache_Prune(t *testing.T) {
	cache, err := NewCache(CacheConfig{
		MemoryOnly: true,
		TTL:        50 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}

	ctx := context.Background()

	// Add entries
	_ = cache.Set(ctx, "prune1", []byte("data1"))
	_ = cache.Set(ctx, "prune2", []byte("data2"))

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Prune expired entries
	pruned, err := cache.Prune(ctx)
	if err != nil {
		t.Fatalf("Prune failed: %v", err)
	}

	if pruned != 2 {
		t.Errorf("expected 2 pruned entries, got %d", pruned)
	}

	// Verify memory is empty
	cache.mu.RLock()
	memCount := len(cache.memory)
	cache.mu.RUnlock()

	if memCount != 0 {
		t.Errorf("expected 0 memory entries after prune, got %d", memCount)
	}
}

func TestWithCache(t *testing.T) {
	cache, err := NewCache(CacheConfig{
		MemoryOnly: true,
		TTL:        time.Hour,
	})
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}

	callCount := 0
	fetch := func() (string, error) {
		callCount++
		return "fetched-value", nil
	}

	// First call should fetch
	result, err := WithCache(cache, "with-cache-test", fetch)
	if err != nil {
		t.Fatalf("WithCache failed: %v", err)
	}
	if result != "fetched-value" {
		t.Errorf("expected fetched-value, got %s", result)
	}
	if callCount != 1 {
		t.Errorf("expected 1 fetch call, got %d", callCount)
	}

	// Second call should use cache
	result, err = WithCache(cache, "with-cache-test", fetch)
	if err != nil {
		t.Fatalf("WithCache failed: %v", err)
	}
	if result != "fetched-value" {
		t.Errorf("expected fetched-value, got %s", result)
	}
	if callCount != 1 {
		t.Errorf("expected 1 fetch call (cached), got %d", callCount)
	}
}

func TestHashKey(t *testing.T) {
	// Same input should produce same hash
	hash1 := hashKey("test-key")
	hash2 := hashKey("test-key")

	if hash1 != hash2 {
		t.Error("expected same hash for same input")
	}

	// Different input should produce different hash
	hash3 := hashKey("different-key")
	if hash1 == hash3 {
		t.Error("expected different hash for different input")
	}

	// Hash should be 32 chars (16 bytes hex encoded)
	if len(hash1) != 32 {
		t.Errorf("expected hash length 32, got %d", len(hash1))
	}
}
