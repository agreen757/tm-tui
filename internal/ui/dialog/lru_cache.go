package dialog

import (
	"container/list"
	"sync"
)

// LRUCache implements a thread-safe Least Recently Used cache with a maximum size
type LRUCache struct {
	maxSize int
	items   map[string]*list.Element
	lruList *list.List
	mu      sync.RWMutex
}

// cacheEntry holds a key-value pair for the LRU list
type cacheEntry struct {
	key   string
	value interface{}
}

// NewLRUCache creates a new LRU cache with the specified maximum size
func NewLRUCache(maxSize int) *LRUCache {
	return &LRUCache{
		maxSize: maxSize,
		items:   make(map[string]*list.Element),
		lruList: list.New(),
	}
}

// Get retrieves a value from the cache
// Returns the value and true if found, nil and false otherwise
func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, found := c.items[key]; found {
		// Move to front (most recently used)
		c.lruList.MoveToFront(element)
		return element.Value.(*cacheEntry).value, true
	}

	return nil, false
}

// Put adds or updates a value in the cache
func (c *LRUCache) Put(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If key exists, update value and move to front
	if element, found := c.items[key]; found {
		c.lruList.MoveToFront(element)
		element.Value.(*cacheEntry).value = value
		return
	}

	// Add new entry
	entry := &cacheEntry{key: key, value: value}
	element := c.lruList.PushFront(entry)
	c.items[key] = element

	// Evict oldest if cache is full
	if c.lruList.Len() > c.maxSize {
		c.evictOldest()
	}
}

// evictOldest removes the least recently used item from the cache
func (c *LRUCache) evictOldest() {
	oldest := c.lruList.Back()
	if oldest != nil {
		c.lruList.Remove(oldest)
		entry := oldest.Value.(*cacheEntry)
		delete(c.items, entry.key)
	}
}

// Delete removes a key from the cache
func (c *LRUCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if element, found := c.items[key]; found {
		c.lruList.Remove(element)
		delete(c.items, key)
	}
}

// Clear removes all entries from the cache
func (c *LRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*list.Element)
	c.lruList.Init()
}

// Len returns the number of items in the cache
func (c *LRUCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.lruList.Len()
}

// Keys returns all keys in the cache (from most to least recently used)
func (c *LRUCache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, c.lruList.Len())
	for e := c.lruList.Front(); e != nil; e = e.Next() {
		keys = append(keys, e.Value.(*cacheEntry).key)
	}

	return keys
}
