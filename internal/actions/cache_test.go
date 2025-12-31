package actions

import (
	"errors"
	"sync"
	"testing"
)

func TestNewCache(t *testing.T) {
	cache := NewCache()
	if cache == nil {
		t.Fatal("NewCache() returned nil")
	}
	if cache.constrained == nil {
		t.Error("NewCache().constrained is nil")
	}
	if cache.unconstrained == nil {
		t.Error("NewCache().unconstrained is nil")
	}
}

func TestCache_ConstrainedGetSet(t *testing.T) {
	cache := NewCache()
	key := NewConstrainedKey("owner", "repo", "v1.0.0", "^1.0.0")

	// Test cache miss
	_, ok := cache.GetConstrained(key)
	if ok {
		t.Error("GetConstrained returned ok=true for empty cache")
	}

	// Test cache set and get
	cache.SetConstrained(key, NewVersionResult("v2.0.0", "abc123", nil))

	result, ok := cache.GetConstrained(key)
	if !ok {
		t.Fatal("GetConstrained returned ok=false after Set")
	}
	if result.Tag != "v2.0.0" {
		t.Errorf("GetConstrained tag = %q, want %q", result.Tag, "v2.0.0")
	}
	if result.Hash != "abc123" {
		t.Errorf("GetConstrained hash = %q, want %q", result.Hash, "abc123")
	}
	if result.Err != nil {
		t.Errorf("GetConstrained err = %v, want nil", result.Err)
	}

	// Test different key returns miss
	differentKey := NewConstrainedKey("owner", "repo", "v1.0.0", "^2.0.0")
	_, ok = cache.GetConstrained(differentKey)
	if ok {
		t.Error("GetConstrained returned ok=true for different pattern")
	}
}

func TestCache_ConstrainedWithError(t *testing.T) {
	cache := NewCache()
	key := NewConstrainedKey("owner", "repo", "v1.0.0", "^1.0.0")

	expectedErr := errors.New("test error")
	cache.SetConstrained(key, NewVersionResult("", "", expectedErr))

	result, ok := cache.GetConstrained(key)
	if !ok {
		t.Fatal("GetConstrained returned ok=false after Set with error")
	}
	if !errors.Is(result.Err, expectedErr) {
		t.Errorf("GetConstrained err = %v, want %v", result.Err, expectedErr)
	}
}

func TestCache_UnconstrainedGetSet(t *testing.T) {
	cache := NewCache()
	key := NewUnconstrainedKey("owner", "repo")

	// Test cache miss
	_, ok := cache.GetUnconstrained(key)
	if ok {
		t.Error("GetUnconstrained returned ok=true for empty cache")
	}

	// Test cache set and get
	cache.SetUnconstrained(key, NewVersionResult("v3.0.0", "def456", nil))

	result, ok := cache.GetUnconstrained(key)
	if !ok {
		t.Fatal("GetUnconstrained returned ok=false after Set")
	}
	if result.Tag != "v3.0.0" {
		t.Errorf("GetUnconstrained tag = %q, want %q", result.Tag, "v3.0.0")
	}
	if result.Hash != "def456" {
		t.Errorf("GetUnconstrained hash = %q, want %q", result.Hash, "def456")
	}
	if result.Err != nil {
		t.Errorf("GetUnconstrained err = %v, want nil", result.Err)
	}

	// Test different key returns miss
	differentKey := NewUnconstrainedKey("owner", "other-repo")
	_, ok = cache.GetUnconstrained(differentKey)
	if ok {
		t.Error("GetUnconstrained returned ok=true for different repo")
	}
}

func TestCache_UnconstrainedWithError(t *testing.T) {
	cache := NewCache()
	key := NewUnconstrainedKey("owner", "repo")

	expectedErr := errors.New("network error")
	cache.SetUnconstrained(key, NewVersionResult("", "", expectedErr))

	result, ok := cache.GetUnconstrained(key)
	if !ok {
		t.Fatal("GetUnconstrained returned ok=false after Set with error")
	}
	if !errors.Is(result.Err, expectedErr) {
		t.Errorf("GetUnconstrained err = %v, want %v", result.Err, expectedErr)
	}
}

func TestCache_Clear(t *testing.T) {
	cache := NewCache()

	// Add some entries
	constrainedKey := NewConstrainedKey("owner", "repo", "v1.0.0", "^1.0.0")
	unconstrainedKey := NewUnconstrainedKey("owner", "repo")

	cache.SetConstrained(constrainedKey, NewVersionResult("v2.0.0", "abc123", nil))
	cache.SetUnconstrained(unconstrainedKey, NewVersionResult("v3.0.0", "def456", nil))

	// Verify they exist
	_, ok := cache.GetConstrained(constrainedKey)
	if !ok {
		t.Fatal("Entry should exist before Clear")
	}
	_, ok = cache.GetUnconstrained(unconstrainedKey)
	if !ok {
		t.Fatal("Entry should exist before Clear")
	}

	// Clear the cache
	cache.Clear()

	// Verify they're gone
	_, ok = cache.GetConstrained(constrainedKey)
	if ok {
		t.Error("Constrained entry should not exist after Clear")
	}
	_, ok = cache.GetUnconstrained(unconstrainedKey)
	if ok {
		t.Error("Unconstrained entry should not exist after Clear")
	}
}

func TestCache_ConcurrentAccess(_ *testing.T) {
	cache := NewCache()
	var wg sync.WaitGroup

	constrainedKey := NewConstrainedKey("owner", "repo", "v1.0.0", "^1.0.0")
	unconstrainedKey := NewUnconstrainedKey("owner", "repo")

	// Run concurrent reads and writes
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				cache.SetConstrained(constrainedKey, NewVersionResult("v2.0.0", "abc123", nil))
				_, _ = cache.GetConstrained(constrainedKey)
				cache.SetUnconstrained(unconstrainedKey, NewVersionResult("v3.0.0", "def456", nil))
				_, _ = cache.GetUnconstrained(unconstrainedKey)
			}
		}()
	}

	wg.Wait()
}

func TestVersionKey_String(t *testing.T) {
	tests := []struct {
		name     string
		key      VersionKey
		expected string
	}{
		{
			name:     "unconstrained key",
			key:      NewUnconstrainedKey("actions", "checkout"),
			expected: "actions/checkout",
		},
		{
			name:     "constrained key",
			key:      NewConstrainedKey("actions", "checkout", "v2.0.0", "^2.0.0"),
			expected: "actions/checkout:v2.0.0:^2.0.0",
		},
		{
			name:     "constrained key empty version",
			key:      NewConstrainedKey("owner", "repo", "", "^1.0.0"),
			expected: "owner/repo::^1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.key.String(); got != tt.expected {
				t.Errorf("VersionKey.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestVersionKey_IsConstrained(t *testing.T) {
	tests := []struct {
		name     string
		key      VersionKey
		expected bool
	}{
		{
			name:     "unconstrained key",
			key:      NewUnconstrainedKey("actions", "checkout"),
			expected: false,
		},
		{
			name:     "constrained with both",
			key:      NewConstrainedKey("actions", "checkout", "v2.0.0", "^2.0.0"),
			expected: true,
		},
		{
			name:     "constrained with pattern only",
			key:      NewConstrainedKey("owner", "repo", "", "^1.0.0"),
			expected: true,
		},
		{
			name:     "constrained with version only",
			key:      NewConstrainedKey("owner", "repo", "v1.0.0", ""),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.key.IsConstrained(); got != tt.expected {
				t.Errorf("VersionKey.IsConstrained() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestCache_Stats(t *testing.T) {
	cache := NewCache()

	// Initial stats are zero
	stats := cache.Stats()
	if stats.Hits != 0 || stats.Misses != 0 {
		t.Errorf("Initial stats = {Hits: %d, Misses: %d}, want {0, 0}", stats.Hits, stats.Misses)
	}

	// Create a miss and check the stats
	key := NewConstrainedKey("owner", "repo", "v1.0.0", "^1.0.0")
	cache.SetConstrained(key, NewVersionResult("v2.0.0", "abc123", nil))

	stats = cache.Stats()
	if stats.Hits != 0 || stats.Misses != 1 {
		t.Errorf("After Set: stats = {Hits: %d, Misses: %d}, want {0, 1}", stats.Hits, stats.Misses)
	}

	// Get a hit and check the stats
	_, ok := cache.GetConstrained(key)
	if !ok {
		t.Fatal("GetConstrained returned ok=false, expected true")
	}

	stats = cache.Stats()
	if stats.Hits != 1 || stats.Misses != 1 {
		t.Errorf("After Get: stats = {Hits: %d, Misses: %d}, want {1, 1}", stats.Hits, stats.Misses)
	}

	// Clear the cache and check the stats
	cache.Clear()

	stats = cache.Stats()
	if stats.Hits != 0 || stats.Misses != 0 {
		t.Errorf("After Clear: stats = {Hits: %d, Misses: %d}, want {0, 0}", stats.Hits, stats.Misses)
	}
}

func TestVersionResult(t *testing.T) {
	result := NewVersionResult("v1.0.0", "abc123", nil)
	if result.Tag != "v1.0.0" {
		t.Errorf("Tag = %q, want %q", result.Tag, "v1.0.0")
	}
	if result.Hash != "abc123" {
		t.Errorf("Hash = %q, want %q", result.Hash, "abc123")
	}
	if result.IsError() {
		t.Error("IsError() = true, want false")
	}

	errResult := NewVersionResult("", "", errors.New("test"))
	if !errResult.IsError() {
		t.Error("IsError() = false, want true")
	}
}
