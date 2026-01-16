package dialog

import (
	"fmt"
	"sync"
	"testing"
)

// TestLRUCacheBasicOperations tests basic get/put operations
func TestLRUCacheBasicOperations(t *testing.T) {
	cache := NewLRUCache(3)

	// Test Put and Get
	cache.Put("key1", "value1")
	value, found := cache.Get("key1")
	if !found {
		t.Error("Expected to find key1")
	}
	if value != "value1" {
		t.Errorf("Expected value1, got %v", value)
	}

	// Test not found
	_, found = cache.Get("nonexistent")
	if found {
		t.Error("Expected not to find nonexistent key")
	}
}

// TestLRUCacheEviction tests that cache evicts oldest items when full
func TestLRUCacheEviction(t *testing.T) {
	cache := NewLRUCache(3)

	// Fill cache to capacity
	cache.Put("key1", "value1")
	cache.Put("key2", "value2")
	cache.Put("key3", "value3")

	if cache.Len() != 3 {
		t.Errorf("Expected cache size 3, got %d", cache.Len())
	}

	// Add one more - should evict key1 (oldest)
	cache.Put("key4", "value4")

	if cache.Len() != 3 {
		t.Errorf("Expected cache size 3 after eviction, got %d", cache.Len())
	}

	// key1 should be evicted
	_, found := cache.Get("key1")
	if found {
		t.Error("Expected key1 to be evicted")
	}

	// key4 should be present
	_, found = cache.Get("key4")
	if !found {
		t.Error("Expected key4 to be present")
	}
}

// TestLRUCacheMoveToFront tests that accessing an item moves it to front
func TestLRUCacheMoveToFront(t *testing.T) {
	cache := NewLRUCache(3)

	// Fill cache
	cache.Put("key1", "value1")
	cache.Put("key2", "value2")
	cache.Put("key3", "value3")

	// Access key1 (moves to front)
	cache.Get("key1")

	// Add key4 - should evict key2 (now oldest)
	cache.Put("key4", "value4")

	// key2 should be evicted (not key1)
	_, found := cache.Get("key2")
	if found {
		t.Error("Expected key2 to be evicted")
	}

	// key1 should still be present
	_, found = cache.Get("key1")
	if !found {
		t.Error("Expected key1 to still be present")
	}
}

// TestLRUCacheUpdate tests updating existing keys
func TestLRUCacheUpdate(t *testing.T) {
	cache := NewLRUCache(3)

	cache.Put("key1", "value1")
	cache.Put("key1", "value1_updated")

	// Should have only one entry
	if cache.Len() != 1 {
		t.Errorf("Expected cache size 1, got %d", cache.Len())
	}

	// Should have updated value
	value, found := cache.Get("key1")
	if !found {
		t.Error("Expected to find key1")
	}
	if value != "value1_updated" {
		t.Errorf("Expected value1_updated, got %v", value)
	}
}

// TestLRUCacheDelete tests deleting items
func TestLRUCacheDelete(t *testing.T) {
	cache := NewLRUCache(3)

	cache.Put("key1", "value1")
	cache.Put("key2", "value2")

	if cache.Len() != 2 {
		t.Errorf("Expected cache size 2, got %d", cache.Len())
	}

	// Delete key1
	cache.Delete("key1")

	if cache.Len() != 1 {
		t.Errorf("Expected cache size 1 after delete, got %d", cache.Len())
	}

	_, found := cache.Get("key1")
	if found {
		t.Error("Expected key1 to be deleted")
	}

	// key2 should still be present
	_, found = cache.Get("key2")
	if !found {
		t.Error("Expected key2 to still be present")
	}
}

// TestLRUCacheClear tests clearing the cache
func TestLRUCacheClear(t *testing.T) {
	cache := NewLRUCache(3)

	cache.Put("key1", "value1")
	cache.Put("key2", "value2")
	cache.Put("key3", "value3")

	if cache.Len() != 3 {
		t.Errorf("Expected cache size 3, got %d", cache.Len())
	}

	cache.Clear()

	if cache.Len() != 0 {
		t.Errorf("Expected cache size 0 after clear, got %d", cache.Len())
	}

	_, found := cache.Get("key1")
	if found {
		t.Error("Expected cache to be empty after clear")
	}
}

// TestLRUCacheKeys tests retrieving all keys
func TestLRUCacheKeys(t *testing.T) {
	cache := NewLRUCache(5)

	cache.Put("key1", "value1")
	cache.Put("key2", "value2")
	cache.Put("key3", "value3")

	keys := cache.Keys()

	if len(keys) != 3 {
		t.Errorf("Expected 3 keys, got %d", len(keys))
	}

	// Keys should be in MRU order (key3, key2, key1)
	expectedKeys := []string{"key3", "key2", "key1"}
	for i, key := range keys {
		if key != expectedKeys[i] {
			t.Errorf("Expected key[%d] to be %s, got %s", i, expectedKeys[i], key)
		}
	}
}

// TestLRUCacheConcurrency tests thread-safe operations
func TestLRUCacheConcurrency(t *testing.T) {
	cache := NewLRUCache(100)
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", n)
			value := fmt.Sprintf("value%d", n)
			cache.Put(key, value)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", n)
			cache.Get(key)
		}(i)
	}

	wg.Wait()

	// Cache should have items (exact count depends on eviction)
	if cache.Len() == 0 {
		t.Error("Expected cache to have items after concurrent operations")
	}
}

// TestLRUCacheWithStructValues tests storing complex types
func TestLRUCacheWithStructValues(t *testing.T) {
	cache := NewLRUCache(3)

	type testStruct struct {
		ID   int
		Name string
	}

	obj := testStruct{ID: 1, Name: "test"}
	cache.Put("obj1", obj)

	value, found := cache.Get("obj1")
	if !found {
		t.Error("Expected to find obj1")
	}

	retrieved, ok := value.(testStruct)
	if !ok {
		t.Error("Expected value to be testStruct type")
	}

	if retrieved.ID != 1 || retrieved.Name != "test" {
		t.Errorf("Expected {1, test}, got {%d, %s}", retrieved.ID, retrieved.Name)
	}
}

// TestLRUCacheZeroSize tests cache with size 0 (edge case)
func TestLRUCacheZeroSize(t *testing.T) {
	cache := NewLRUCache(0)

	cache.Put("key1", "value1")

	// Should not store anything
	if cache.Len() != 0 {
		t.Errorf("Expected cache size 0, got %d", cache.Len())
	}

	_, found := cache.Get("key1")
	if found {
		t.Error("Expected cache with size 0 to not store items")
	}
}

// BenchmarkLRUCachePut benchmarks Put operations
func BenchmarkLRUCachePut(b *testing.B) {
	cache := NewLRUCache(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i%1000)
		value := fmt.Sprintf("value%d", i)
		cache.Put(key, value)
	}
}

// BenchmarkLRUCacheGet benchmarks Get operations
func BenchmarkLRUCacheGet(b *testing.B) {
	cache := NewLRUCache(1000)

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		value := fmt.Sprintf("value%d", i)
		cache.Put(key, value)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i%1000)
		cache.Get(key)
	}
}

// BenchmarkLRUCacheMixed benchmarks mixed operations
func BenchmarkLRUCacheMixed(b *testing.B) {
	cache := NewLRUCache(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i%200)
		value := fmt.Sprintf("value%d", i)

		if i%2 == 0 {
			cache.Put(key, value)
		} else {
			cache.Get(key)
		}
	}
}
