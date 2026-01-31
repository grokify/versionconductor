package graph

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Cache provides caching for graph-related data.
// It uses a simple file-based cache with TTL support.
type Cache struct {
	dir    string
	ttl    time.Duration
	mu     sync.RWMutex
	memory map[string]*cacheEntry
}

type cacheEntry struct {
	data      []byte
	expiresAt time.Time
}

// CacheConfig configures the cache behavior.
type CacheConfig struct {
	// Dir is the directory for file-based cache. If empty, uses temp dir.
	Dir string

	// TTL is the time-to-live for cached entries. Default is 1 hour.
	TTL time.Duration

	// MemoryOnly disables file-based caching.
	MemoryOnly bool
}

// NewCache creates a new cache with the given configuration.
func NewCache(cfg CacheConfig) (*Cache, error) {
	if cfg.TTL == 0 {
		cfg.TTL = time.Hour
	}

	dir := cfg.Dir
	if dir == "" && !cfg.MemoryOnly {
		dir = filepath.Join(os.TempDir(), "versionconductor-cache")
	}

	if dir != "" {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, fmt.Errorf("failed to create cache directory: %w", err)
		}
	}

	return &Cache{
		dir:    dir,
		ttl:    cfg.TTL,
		memory: make(map[string]*cacheEntry),
	}, nil
}

// Get retrieves a cached value by key.
func (c *Cache) Get(ctx context.Context, key string) ([]byte, bool) {
	hash := hashKey(key)

	// Check memory cache first
	c.mu.RLock()
	if entry, ok := c.memory[hash]; ok {
		if time.Now().Before(entry.expiresAt) {
			c.mu.RUnlock()
			return entry.data, true
		}
	}
	c.mu.RUnlock()

	// Check file cache
	if c.dir != "" {
		data, err := c.getFromFile(hash)
		if err == nil {
			// Populate memory cache
			c.mu.Lock()
			c.memory[hash] = &cacheEntry{
				data:      data,
				expiresAt: time.Now().Add(c.ttl),
			}
			c.mu.Unlock()
			return data, true
		}
	}

	return nil, false
}

// Set stores a value in the cache.
func (c *Cache) Set(ctx context.Context, key string, data []byte) error {
	hash := hashKey(key)
	expiresAt := time.Now().Add(c.ttl)

	// Store in memory
	c.mu.Lock()
	c.memory[hash] = &cacheEntry{
		data:      data,
		expiresAt: expiresAt,
	}
	c.mu.Unlock()

	// Store in file
	if c.dir != "" {
		if err := c.setToFile(hash, data, expiresAt); err != nil {
			return err
		}
	}

	return nil
}

// Delete removes a value from the cache.
func (c *Cache) Delete(ctx context.Context, key string) error {
	hash := hashKey(key)

	// Remove from memory
	c.mu.Lock()
	delete(c.memory, hash)
	c.mu.Unlock()

	// Remove from file
	if c.dir != "" {
		path := filepath.Join(c.dir, hash+".json")
		_ = os.Remove(path)
		_ = os.Remove(path + ".meta")
	}

	return nil
}

// Clear removes all entries from the cache.
func (c *Cache) Clear(ctx context.Context) error {
	// Clear memory
	c.mu.Lock()
	c.memory = make(map[string]*cacheEntry)
	c.mu.Unlock()

	// Clear files
	if c.dir != "" {
		entries, err := os.ReadDir(c.dir)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".json") || strings.HasSuffix(entry.Name(), ".meta") {
				_ = os.Remove(filepath.Join(c.dir, entry.Name()))
			}
		}
	}

	return nil
}

// getFromFile retrieves a cached value from file.
func (c *Cache) getFromFile(hash string) ([]byte, error) {
	metaPath := filepath.Join(c.dir, hash+".meta")
	dataPath := filepath.Join(c.dir, hash+".json")

	// Check metadata for expiration
	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, err
	}

	var meta struct {
		ExpiresAt time.Time `json:"expiresAt"`
	}
	if err := json.Unmarshal(metaData, &meta); err != nil {
		return nil, err
	}

	if time.Now().After(meta.ExpiresAt) {
		// Expired, clean up
		_ = os.Remove(metaPath)
		_ = os.Remove(dataPath)
		return nil, fmt.Errorf("cache expired")
	}

	// Read data
	return os.ReadFile(dataPath)
}

// setToFile stores a cached value to file.
func (c *Cache) setToFile(hash string, data []byte, expiresAt time.Time) error {
	metaPath := filepath.Join(c.dir, hash+".meta")
	dataPath := filepath.Join(c.dir, hash+".json")

	// Write metadata
	meta := struct {
		ExpiresAt time.Time `json:"expiresAt"`
	}{
		ExpiresAt: expiresAt,
	}
	metaData, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	if err := os.WriteFile(metaPath, metaData, 0600); err != nil {
		return err
	}

	// Write data
	return os.WriteFile(dataPath, data, 0600)
}

// hashKey creates a hash of the cache key.
func hashKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:16]) // Use first 16 bytes for shorter filenames
}

// CachedBuilder wraps Builder with caching support.
type CachedBuilder struct {
	builder *Builder
	cache   *Cache
}

// NewCachedBuilder creates a new builder with caching.
func NewCachedBuilder(token string, cache *Cache) *CachedBuilder {
	return &CachedBuilder{
		builder: NewBuilder(token),
		cache:   cache,
	}
}

// Build constructs a dependency graph with caching.
func (cb *CachedBuilder) Build(ctx context.Context, portfolio Portfolio) (*DependencyGraph, error) {
	// Check for cached graph
	cacheKey := cb.graphCacheKey(portfolio)
	if data, ok := cb.cache.Get(ctx, cacheKey); ok {
		var snapshot GraphSnapshot
		if err := json.Unmarshal(data, &snapshot); err == nil {
			return BuildFromSnapshot(&snapshot), nil
		}
	}

	// Build fresh graph
	graph, err := cb.builder.Build(ctx, portfolio)
	if err != nil {
		return nil, err
	}

	// Cache the result
	snapshot := graph.Snapshot()
	if data, err := json.Marshal(snapshot); err == nil {
		_ = cb.cache.Set(ctx, cacheKey, data)
	}

	return graph, nil
}

// graphCacheKey creates a cache key for a portfolio graph.
func (cb *CachedBuilder) graphCacheKey(portfolio Portfolio) string {
	return fmt.Sprintf("graph:%s:%s", portfolio.Name, strings.Join(portfolio.Orgs, ","))
}

// InvalidateGraph removes the cached graph for a portfolio.
func (cb *CachedBuilder) InvalidateGraph(ctx context.Context, portfolio Portfolio) error {
	return cb.cache.Delete(ctx, cb.graphCacheKey(portfolio))
}

// GoModCache provides caching specifically for go.mod files.
type GoModCache struct {
	cache *Cache
}

// NewGoModCache creates a cache for go.mod files.
func NewGoModCache(cache *Cache) *GoModCache {
	return &GoModCache{cache: cache}
}

// Get retrieves a cached go.mod file.
func (gmc *GoModCache) Get(ctx context.Context, owner, repo, ref string) ([]byte, bool) {
	key := fmt.Sprintf("gomod:%s/%s:%s", owner, repo, ref)
	return gmc.cache.Get(ctx, key)
}

// Set stores a go.mod file in the cache.
func (gmc *GoModCache) Set(ctx context.Context, owner, repo, ref string, content []byte) error {
	key := fmt.Sprintf("gomod:%s/%s:%s", owner, repo, ref)
	return gmc.cache.Set(ctx, key, content)
}

// RepoListCache provides caching for repository listings.
type RepoListCache struct {
	cache *Cache
}

// NewRepoListCache creates a cache for repo listings.
func NewRepoListCache(cache *Cache) *RepoListCache {
	return &RepoListCache{cache: cache}
}

// Get retrieves a cached repo list.
func (rlc *RepoListCache) Get(ctx context.Context, owner string) ([]string, bool) {
	key := fmt.Sprintf("repos:%s", owner)
	data, ok := rlc.cache.Get(ctx, key)
	if !ok {
		return nil, false
	}

	var repos []string
	if err := json.Unmarshal(data, &repos); err != nil {
		return nil, false
	}
	return repos, true
}

// Set stores a repo list in the cache.
func (rlc *RepoListCache) Set(ctx context.Context, owner string, repos []string) error {
	key := fmt.Sprintf("repos:%s", owner)
	data, err := json.Marshal(repos)
	if err != nil {
		return err
	}
	return rlc.cache.Set(ctx, key, data)
}

// CacheStats provides statistics about cache usage.
type CacheStats struct {
	MemoryEntries int   `json:"memoryEntries"`
	FileEntries   int   `json:"fileEntries"`
	TotalSizeKB   int64 `json:"totalSizeKB"`
}

// Stats returns cache statistics.
func (c *Cache) Stats(ctx context.Context) CacheStats {
	c.mu.RLock()
	memCount := len(c.memory)
	c.mu.RUnlock()

	stats := CacheStats{
		MemoryEntries: memCount,
	}

	if c.dir != "" {
		entries, err := os.ReadDir(c.dir)
		if err == nil {
			var totalSize int64
			for _, entry := range entries {
				if strings.HasSuffix(entry.Name(), ".json") {
					stats.FileEntries++
					if info, err := entry.Info(); err == nil {
						totalSize += info.Size()
					}
				}
			}
			stats.TotalSizeKB = totalSize / 1024
		}
	}

	return stats
}

// Prune removes expired entries from the cache.
func (c *Cache) Prune(ctx context.Context) (int, error) {
	pruned := 0

	// Prune memory cache
	c.mu.Lock()
	now := time.Now()
	for key, entry := range c.memory {
		if now.After(entry.expiresAt) {
			delete(c.memory, key)
			pruned++
		}
	}
	c.mu.Unlock()

	// Prune file cache
	if c.dir != "" {
		entries, err := os.ReadDir(c.dir)
		if err != nil {
			return pruned, err
		}

		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".meta") {
				metaPath := filepath.Join(c.dir, entry.Name())
				metaData, err := os.ReadFile(metaPath)
				if err != nil {
					continue
				}

				var meta struct {
					ExpiresAt time.Time `json:"expiresAt"`
				}
				if err := json.Unmarshal(metaData, &meta); err != nil {
					continue
				}

				if now.After(meta.ExpiresAt) {
					hash := strings.TrimSuffix(entry.Name(), ".meta")
					_ = os.Remove(metaPath)
					_ = os.Remove(filepath.Join(c.dir, hash+".json"))
					pruned++
				}
			}
		}
	}

	return pruned, nil
}

// WithCache adds caching to a reader function.
func WithCache[T any](cache *Cache, key string, fetch func() (T, error)) (T, error) {
	ctx := context.Background()

	// Check cache
	if data, ok := cache.Get(ctx, key); ok {
		var result T
		if err := json.Unmarshal(data, &result); err == nil {
			return result, nil
		}
	}

	// Fetch fresh data
	result, err := fetch()
	if err != nil {
		var zero T
		return zero, err
	}

	// Cache result
	if data, err := json.Marshal(result); err == nil {
		_ = cache.Set(ctx, key, data)
	}

	return result, nil
}

// StreamToCache streams reader content to cache.
func StreamToCache(cache *Cache, key string, r io.Reader) ([]byte, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if err := cache.Set(context.Background(), key, data); err != nil {
		return data, err
	}

	return data, nil
}
