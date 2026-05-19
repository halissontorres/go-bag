package cache

import "sync"

// LRUCache is a generic cache with least-recently-used eviction.
// It is not safe for concurrent use; see SyncLRUCache for a thread-safe wrapper.
type LRUCache[K comparable, V any] struct {
	data *internalCache[K, V]
}

type internalCache[K comparable, V any] struct {
	capacity int
	onEvict  func(K, V)
	items    map[K]*entry[K, V]
	head     *entry[K, V]
	tail     *entry[K, V]
}

type entry[K comparable, V any] struct {
	key        K
	value      V
	prev, next *entry[K, V]
}

// New creates an LRUCache configured by the provided options.
// WithCapacity is mandatory.
func New[K comparable, V any](opts ...CacheOption[K, V]) *LRUCache[K, V] {
	c := &internalCache[K, V]{
		items: make(map[K]*entry[K, V]),
	}
	for _, o := range opts {
		o(c)
	}
	if c.capacity <= 0 {
		panic("cache: New requires WithCapacity > 0")
	}
	return &LRUCache[K, V]{data: c}
}

// Get returns the value for key, marking it as most recently used.
func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	e, ok := c.data.items[key]
	if !ok {
		var zero V
		return zero, false
	}
	c.data.moveToFront(e)
	return e.value, true
}

// Peek returns the value for key without updating its access order.
func (c *LRUCache[K, V]) Peek(key K) (V, bool) {
	e, ok := c.data.items[key]
	if !ok {
		var zero V
		return zero, false
	}
	return e.value, true
}

// Put inserts or updates the value for key. If the key is new and the cache
// is at capacity, the least recently used entry is evicted.
func (c *LRUCache[K, V]) Put(key K, value V) {
	if e, ok := c.data.items[key]; ok {
		e.value = value
		c.data.moveToFront(e)
		return
	}
	if len(c.data.items) == c.data.capacity {
		c.data.evictLRU()
	}
	e := &entry[K, V]{key: key, value: value}
	c.data.items[key] = e
	c.data.pushFront(e)
}

// Remove removes the entry for key and returns its value.
func (c *LRUCache[K, V]) Remove(key K) (V, bool) {
	e, ok := c.data.items[key]
	if !ok {
		var zero V
		return zero, false
	}
	c.data.remove(e)
	delete(c.data.items, key)
	if c.data.onEvict != nil {
		c.data.onEvict(e.key, e.value)
	}
	return e.value, true
}

// Contains reports whether key is present in the cache.
func (c *LRUCache[K, V]) Contains(key K) bool {
	_, ok := c.data.items[key]
	return ok
}

// Len returns the current number of entries in the cache.
func (c *LRUCache[K, V]) Len() int {
	return len(c.data.items)
}

// Cap returns the configured maximum capacity.
func (c *LRUCache[K, V]) Cap() int {
	return c.data.capacity
}

// Keys returns the cache keys from most to least recently used.
func (c *LRUCache[K, V]) Keys() []K {
	keys := make([]K, 0, len(c.data.items))
	for e := c.data.head; e != nil; e = e.next {
		keys = append(keys, e.key)
	}
	return keys
}

// Values returns the cache values from most to least recently used.
func (c *LRUCache[K, V]) Values() []V {
	values := make([]V, 0, len(c.data.items))
	for e := c.data.head; e != nil; e = e.next {
		values = append(values, e.value)
	}
	return values
}

// Clear removes all entries from the cache.
func (c *LRUCache[K, V]) Clear() {
	if c.data.onEvict != nil {
		for e := c.data.head; e != nil; e = e.next {
			c.data.onEvict(e.key, e.value)
		}
	}
	c.data.items = make(map[K]*entry[K, V])
	c.data.head = nil
	c.data.tail = nil
}

// Resize changes the cache capacity. If the new capacity is smaller than the
// current size, the least recently used entries are evicted until the cache
// fits within the new capacity.
func (c *LRUCache[K, V]) Resize(capacity int) {
	if capacity <= 0 {
		panic("cache: Resize capacity must be > 0")
	}
	c.data.capacity = capacity
	for len(c.data.items) > c.data.capacity {
		c.data.evictLRU()
	}
}

// ---------- internal list operations ----------

func (c *internalCache[K, V]) pushFront(e *entry[K, V]) {
	e.prev = nil
	e.next = c.head
	if c.head != nil {
		c.head.prev = e
	}
	c.head = e
	if c.tail == nil {
		c.tail = e
	}
}

func (c *internalCache[K, V]) remove(e *entry[K, V]) {
	if e.prev != nil {
		e.prev.next = e.next
	} else {
		c.head = e.next
	}
	if e.next != nil {
		e.next.prev = e.prev
	} else {
		c.tail = e.prev
	}
}

func (c *internalCache[K, V]) moveToFront(e *entry[K, V]) {
	if c.head == e {
		return
	}
	c.remove(e)
	c.pushFront(e)
}

func (c *internalCache[K, V]) evictLRU() {
	if c.tail == nil {
		return
	}
	e := c.tail
	c.remove(e)
	delete(c.items, e.key)
	if c.onEvict != nil {
		c.onEvict(e.key, e.value)
	}
}

// ---------- Sync variant ----------

// SyncLRUCache is a thread-safe wrapper for LRUCache.
type SyncLRUCache[K comparable, V any] struct {
	mu sync.RWMutex
	c  *LRUCache[K, V]
}

// NewSync creates a thread-safe LRUCache configured by the provided options.
func NewSync[K comparable, V any](opts ...CacheOption[K, V]) *SyncLRUCache[K, V] {
	return &SyncLRUCache[K, V]{c: New[K, V](opts...)}
}

func (sc *SyncLRUCache[K, V]) Get(key K) (V, bool) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.c.Get(key)
}

func (sc *SyncLRUCache[K, V]) Peek(key K) (V, bool) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.c.Peek(key)
}

func (sc *SyncLRUCache[K, V]) Put(key K, value V) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.c.Put(key, value)
}

func (sc *SyncLRUCache[K, V]) Remove(key K) (V, bool) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.c.Remove(key)
}

func (sc *SyncLRUCache[K, V]) Contains(key K) bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.c.Contains(key)
}

func (sc *SyncLRUCache[K, V]) Len() int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.c.Len()
}

func (sc *SyncLRUCache[K, V]) Cap() int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.c.Cap()
}

func (sc *SyncLRUCache[K, V]) Keys() []K {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.c.Keys()
}

func (sc *SyncLRUCache[K, V]) Values() []V {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.c.Values()
}

func (sc *SyncLRUCache[K, V]) Clear() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.c.Clear()
}

func (sc *SyncLRUCache[K, V]) Resize(capacity int) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.c.Resize(capacity)
}
