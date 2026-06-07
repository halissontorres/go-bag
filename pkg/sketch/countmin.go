// Package sketch provides probabilistic data structures for processing
// data streams with sub-linear space: CountMinSketch for frequency
// estimation and HyperLogLog for cardinality estimation.
package sketch

import (
	"math"
	"sync"
)

// countMin holds the sketch matrix and related parameters.
type countMin[T any] struct {
	matrix [][]uint64 // d × w counter matrix
	width  uint64
	depth  int
	hash   func(T) uint64
	total  uint64
}

// CountMin is a Count-Min Sketch for estimating item frequencies in a data
// stream. It uses O(w × d) space and provides estimates with an error bounded
// by ε × N with probability 1 − δ, where N is the sum of all counts.
//
// CountMin is not safe for concurrent use. Use SyncCountMin for shared access.
type CountMin[T any] struct {
	sketch *countMin[T]
}

// NewCountMin creates a Count-Min Sketch with the given error bounds.
//
// epsilon controls the relative error: the estimate for any item is at most
// ε × N above its true count (where N is the total of all increments).
//
// delta controls the error probability: P[estimate > true + ε×N] ≤ delta.
//
// The hash function must distribute items uniformly across uint64.
// A common choice is fnv.New64a or maphash.Hash.
//
// Width w = ⌈e/ε⌉ and depth d = ⌈ln(1/δ)⌉ are derived automatically.
func NewCountMin[T any](epsilon, delta float64, hash func(T) uint64, opts ...CountMinOption[T]) *CountMin[T] {
	w := uint64(math.Ceil(math.E / epsilon))
	d := int(math.Ceil(math.Log(1.0 / delta)))

	cm := &countMin[T]{
		matrix: make([][]uint64, d),
		width:  w,
		depth:  d,
		hash:   hash,
	}
	for i := range d {
		cm.matrix[i] = make([]uint64, w)
	}
	for _, o := range opts {
		o(cm)
	}
	return &CountMin[T]{sketch: cm}
}

// Increment adds 1 to the estimated frequency of item.
func (c *CountMin[T]) Increment(item T) {
	c.Update(item, 1)
}

// Update adds count to the estimated frequency of item.
func (c *CountMin[T]) Update(item T, count uint64) {
	h := c.sketch.hash(item)
	c.sketch.total += count

	for i := range c.sketch.depth {
		col := rowHash(h, i) % c.sketch.width
		c.sketch.matrix[i][col] += count
	}
}

// Estimate returns the estimated frequency of item. The true frequency is
// always ≤ the estimate (Count-Min never underestimates).
func (c *CountMin[T]) Estimate(item T) uint64 {
	h := c.sketch.hash(item)
	min := uint64(math.MaxUint64)

	for i := range c.sketch.depth {
		col := rowHash(h, i) % c.sketch.width
		v := c.sketch.matrix[i][col]
		if v < min {
			min = v
		}
	}
	return min
}

// Total returns the sum of all counts passed to Update and Increment.
func (c *CountMin[T]) Total() uint64 {
	return c.sketch.total
}

// SyncCountMin is a thread-safe Count-Min Sketch.
type SyncCountMin[T any] struct {
	mu     sync.RWMutex
	sketch *CountMin[T]
}

// NewSyncCountMin creates a thread-safe Count-Min Sketch.
func NewSyncCountMin[T any](epsilon, delta float64, hash func(T) uint64, opts ...CountMinOption[T]) *SyncCountMin[T] {
	return &SyncCountMin[T]{sketch: NewCountMin[T](epsilon, delta, hash, opts...)}
}

// Increment adds 1 atomically.
func (s *SyncCountMin[T]) Increment(item T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sketch.Increment(item)
}

// Update adds count atomically.
func (s *SyncCountMin[T]) Update(item T, count uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sketch.Update(item, count)
}

// Estimate returns the estimated frequency.
func (s *SyncCountMin[T]) Estimate(item T) uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sketch.Estimate(item)
}

// Total returns the sum of all counts.
func (s *SyncCountMin[T]) Total() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sketch.Total()
}

// rowHash derives a per-row hash value from a single input hash.
// Uses splitmix64-style mixing to decorrelate rows.
func rowHash(h uint64, row int) uint64 {
	key := h
	key ^= uint64(row+1) * 0x9e3779b97f4a7c15
	key = (key ^ (key >> 30)) * 0xbf58476d1ce4e5b9
	key = (key ^ (key >> 27)) * 0x94d049bb133111eb
	key ^= key >> 31
	return key
}
