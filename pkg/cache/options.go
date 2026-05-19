package cache

// CacheOption configures an LRUCache at creation time.
type CacheOption[K comparable, V any] func(*internalCache[K, V])

// WithCapacity sets the maximum number of entries the cache can hold.
// This is mandatory — New will panic if a positive capacity is not provided.
func WithCapacity[K comparable, V any](n int) CacheOption[K, V] {
	return func(c *internalCache[K, V]) {
		if n <= 0 {
			panic("cache: WithCapacity must be > 0")
		}
		c.capacity = n
	}
}

// WithOnEvict registers a callback that is invoked when an entry is evicted
// from the cache (by overflow, Remove, Clear, or Resize).
func WithOnEvict[K comparable, V any](fn func(K, V)) CacheOption[K, V] {
	return func(c *internalCache[K, V]) {
		c.onEvict = fn
	}
}
