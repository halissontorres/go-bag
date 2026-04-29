package set

import "sync"

// Set is a mutable set of unique elements of type T.
// The zero value is a usable empty set (do not take a pointer to it).
type Set[T comparable] struct {
	m map[T]struct{}
}

// NewSet creates a new, empty Set.
func NewSet[T comparable]() *Set[T] {
	return &Set[T]{
		m: make(map[T]struct{}),
	}
}

// Add inserts one or more elements into the set.
func (s *Set[T]) Add(values ...T) {
	if s.m == nil {
		s.m = make(map[T]struct{})
	}
	for _, v := range values {
		s.m[v] = struct{}{}
	}
}

// Remove deletes one or more elements from the set.
func (s *Set[T]) Remove(values ...T) {
	if s.m == nil {
		return
	}
	for _, v := range values {
		delete(s.m, v)
	}
}

// Contains reports whether the element is present in the set.
func (s *Set[T]) Contains(value T) bool {
	if s.m == nil {
		return false
	}
	_, ok := s.m[value]
	return ok
}

// Len returns the number of elements.
func (s *Set[T]) Len() int {
	if s.m == nil {
		return 0
	}
	return len(s.m)
}

// Clear removes all elements.
func (s *Set[T]) Clear() {
	if s.m != nil {
		s.m = make(map[T]struct{})
	}
}

// Elements returns a slice with all elements of the set.
// Iteration order is not guaranteed.
func (s *Set[T]) Elements() []T {
	if s.m == nil {
		return nil
	}
	elements := make([]T, 0, len(s.m))
	for v := range s.m {
		elements = append(elements, v)
	}
	return elements
}

// Union returns a new Set containing the union of a and b.
func Union[T comparable](a, b *Set[T]) *Set[T] {
	result := NewSet[T]()
	if a != nil {
		for v := range a.m {
			result.Add(v)
		}
	}
	if b != nil {
		for v := range b.m {
			result.Add(v)
		}
	}
	return result
}

// Intersection returns a new Set containing the intersection of a and b.
func Intersection[T comparable](a, b *Set[T]) *Set[T] {
	result := NewSet[T]()
	if a == nil || b == nil {
		return result
	}
	for v := range a.m {
		if b.Contains(v) {
			result.Add(v)
		}
	}
	return result
}

// Difference returns a new Set containing the elements of a that are not in b.
func Difference[T comparable](a, b *Set[T]) *Set[T] {
	result := NewSet[T]()
	if a == nil {
		return result
	}
	for v := range a.m {
		if b == nil || !b.Contains(v) {
			result.Add(v)
		}
	}
	return result
}

// SubsetOf reports whether a is a subset of b.
func SubsetOf[T comparable](a, b *Set[T]) bool {
	if a == nil {
		return true
	}
	if b == nil {
		return a.Len() == 0
	}
	for v := range a.m {
		if !b.Contains(v) {
			return false
		}
	}
	return true
}

// Equal reports whether two sets contain the same elements.
func Equal[T comparable](a, b *Set[T]) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if a.Len() != b.Len() {
		return false
	}
	for v := range a.m {
		if !b.Contains(v) {
			return false
		}
	}
	return true
}

// ===== Functional helpers =====

// ForEach calls f for every element in the set.
// Iteration order is not guaranteed.
func (s *Set[T]) ForEach(f func(T)) {
	if s.m == nil {
		return
	}
	for v := range s.m {
		f(v)
	}
}

// Filter returns a new set containing only the elements that satisfy the predicate.
func (s *Set[T]) Filter(pred func(T) bool) *Set[T] {
	result := NewSet[T]()
	if s.m == nil {
		return result
	}
	for v := range s.m {
		if pred(v) {
			result.Add(v)
		}
	}
	return result
}

// Any reports whether at least one element satisfies the predicate.
func (s *Set[T]) Any(pred func(T) bool) bool {
	if s.m == nil {
		return false
	}
	for v := range s.m {
		if pred(v) {
			return true
		}
	}
	return false
}

// All reports whether every element satisfies the predicate.
func (s *Set[T]) All(pred func(T) bool) bool {
	if s.m == nil {
		return true // empty set: vacuously true.
	}
	for v := range s.m {
		if !pred(v) {
			return false
		}
	}
	return true
}

// Clone returns a shallow copy of the set.
func (s *Set[T]) Clone() *Set[T] {
	newSet := NewSet[T]()
	if s.m == nil {
		return newSet
	}
	for v := range s.m {
		newSet.Add(v)
	}
	return newSet
}

// MapSet applies f to every element and returns a new set with the results.
// The target type U must be comparable.
func MapSet[T comparable, U comparable](s *Set[T], f func(T) U) *Set[U] {
	result := NewSet[U]()
	if s == nil || s.m == nil {
		return result
	}
	for v := range s.m {
		result.Add(f(v))
	}
	return result
}

// ReduceSet reduces the set to a single value using the aggregation function f.
// The initial accumulator is initial, and f is called for each element as f(acc, elem).
func ReduceSet[T comparable, R any](s *Set[T], initial R, f func(acc R, elem T) R) R {
	acc := initial
	if s == nil || s.m == nil {
		return acc
	}
	for v := range s.m {
		acc = f(acc, v)
	}
	return acc
}

// SyncSet is a thread-safe Set.
type SyncSet[T comparable] struct {
	mu sync.RWMutex
	s  *Set[T]
}

func NewSyncSet[T comparable]() *SyncSet[T] {
	return &SyncSet[T]{s: NewSet[T]()}
}

func (ss *SyncSet[T]) Add(values ...T) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.s.Add(values...)
}

func (ss *SyncSet[T]) Remove(values ...T) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.s.Remove(values...)
}

func (ss *SyncSet[T]) Contains(value T) bool {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.s.Contains(value)
}

func (ss *SyncSet[T]) Len() int {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.s.Len()
}

func (ss *SyncSet[T]) Elements() []T {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.s.Elements()
}
