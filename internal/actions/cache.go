package actions

import "sync"

// CacheStats holds statistics about cache usage.
type CacheStats struct {
	Hits   int64 // Number of cache hits
	Misses int64 // Number of cache misses (API calls made)
}

// Cache stores cached results for version lookups to avoid duplicate API calls
// when the same action appears in multiple workflows.
type Cache struct {
	mu            sync.Mutex
	constrained   map[string]VersionResult // Results for configured actions with version patterns
	unconstrained map[string]VersionResult // Results for unconfigured actions (absolute latest)
	hits          int64
	misses        int64
}

// NewCache creates a new Cache instance.
func NewCache() *Cache {
	return &Cache{
		constrained:   make(map[string]VersionResult),
		unconstrained: make(map[string]VersionResult),
	}
}

// GetConstrained retrieves a cached version result for a constrained lookup.
// Returns the cached result and true if found, or zero result and false if not cached.
func (c *Cache) GetConstrained(key VersionKey) (VersionResult, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if cached, found := c.constrained[key.String()]; found {
		c.hits++
		return cached, true
	}
	return VersionResult{}, false
}

// SetConstrained stores a version result in the cache for a constrained lookup.
func (c *Cache) SetConstrained(key VersionKey, result VersionResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.constrained[key.String()] = result
	c.misses++ // A set means we made an API call
}

// GetUnconstrained retrieves a cached version result for an unconstrained lookup.
// Returns the cached result and true if found, or zero result and false if not cached.
func (c *Cache) GetUnconstrained(key VersionKey) (VersionResult, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if cached, found := c.unconstrained[key.String()]; found {
		c.hits++
		return cached, true
	}
	return VersionResult{}, false
}

// SetUnconstrained stores a version result in the cache for an unconstrained lookup.
func (c *Cache) SetUnconstrained(key VersionKey, result VersionResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.unconstrained[key.String()] = result
	c.misses++ // A set means we made an API call
}

// Clear clears all cached version lookup results.
// This is useful for testing or when you want to force fresh API calls.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	clear(c.constrained)
	clear(c.unconstrained)
	c.hits = 0
	c.misses = 0
}

// Stats returns the current cache statistics.
func (c *Cache) Stats() CacheStats {
	c.mu.Lock()
	defer c.mu.Unlock()

	return CacheStats{
		Hits:   c.hits,
		Misses: c.misses,
	}
}
