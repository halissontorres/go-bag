package stream

import (
	"slices"

	"github.com/halissontorres/go-bag/pkg/comparator"
)

// Filter returns a Stream containing only elements that satisfy the predicate.
func Filter[T any](s *Stream[T], pred func(T) bool) *Stream[T] {
	return NewStream(func() (T, bool) {
		for {
			val, ok := s.next()
			if !ok {
				var zero T
				return zero, false
			}
			if pred(val) {
				return val, true
			}
		}
	})
}

// Map transforms each element using function f.
func Map[T, U any](s *Stream[T], f func(T) U) *Stream[U] {
	return NewStream(func() (U, bool) {
		val, ok := s.next()
		if !ok {
			var zero U
			return zero, false
		}
		return f(val), true
	})
}

// FlatMap applies f to each element and concatenates the resulting slices.
func FlatMap[T, U any](s *Stream[T], f func(T) []U) *Stream[U] {
	var cur []U
	idx := 0
	return NewStream(func() (U, bool) {
		for {
			if idx < len(cur) {
				val := cur[idx]
				idx++
				return val, true
			}
			elem, ok := s.next()
			if !ok {
				var zero U
				return zero, false
			}
			cur = f(elem)
			idx = 0
		}
	})
}

// Distinct eliminates duplicates (keeps first occurrence). Requires comparable.
// Accepts an optional WithInitialCap to pre-size the internal seen-map.
// Default initial capacity: 256.
func Distinct[T comparable](s *Stream[T], opts ...Option) *Stream[T] {
	c := applyStreamOptions(opts)
	seen := make(map[T]struct{}, c.initialCap)
	return NewStream(func() (T, bool) {
		for {
			val, ok := s.next()
			if !ok {
				var zero T
				return zero, false
			}
			if _, exists := seen[val]; !exists {
				seen[val] = struct{}{}
				return val, true
			}
		}
	})
}

// SortedBy collects all elements, sorts them using the provided Comparator,
// and emits them in order. Works with any type T.
func SortedBy[T any](s *Stream[T], less comparator.Comparator[T], opts ...Option) *Stream[T] {
	collected := s.ToSlice(opts...)
	slices.SortFunc(collected, func(a, b T) int {
		if less(a, b) {
			return -1
		}
		if less(b, a) {
			return 1
		}
		return 0
	})
	return FromSlice(collected)
}

// DistinctBy eliminates duplicates using the Comparator for equality.
// Two elements are considered equal when neither less(a,b) nor less(b,a) is true.
// Use when T is not comparable or when equality is defined by ordering.
func DistinctBy[T any](s *Stream[T], less comparator.Comparator[T], opts ...Option) *Stream[T] {
	c := applyStreamOptions(opts)
	seen := make([]T, 0, c.initialCap)
	return NewStream(func() (T, bool) {
		for {
			val, ok := s.next()
			if !ok {
				var zero T
				return zero, false
			}
			found := false
			for _, v := range seen {
				if !less(val, v) && !less(v, val) {
					found = true
					break
				}
			}
			if !found {
				seen = append(seen, val)
				return val, true
			}
		}
	})
}

// Limit limits the number of elements emitted.
func Limit[T any](s *Stream[T], n int) *Stream[T] {
	count := 0
	return NewStream(func() (T, bool) {
		if count >= n {
			var zero T
			return zero, false
		}
		val, ok := s.next()
		if ok {
			count++
		}
		return val, ok
	})
}

// Skip ignores the first n elements.
func Skip[T any](s *Stream[T], n int) *Stream[T] {
	ready := false
	return NewStream(func() (T, bool) {
		if !ready {
			ready = true
			for i := 0; i < n; i++ {
				if _, ok := s.next(); !ok {
					var zero T
					return zero, false
				}
			}
		}
		return s.next()
	})
}

// Concat concatenates multiple Streams.
func Concat[T any](streams ...*Stream[T]) *Stream[T] {
	idx := 0
	return NewStream(func() (T, bool) {
		for idx < len(streams) {
			val, ok := streams[idx].next()
			if ok {
				return val, true
			}
			idx++
		}
		var zero T
		return zero, false
	})
}

// Peek applies an action to each element without consuming it (useful for debugging).
func Peek[T any](s *Stream[T], action func(T)) *Stream[T] {
	return NewStream(func() (T, bool) {
		val, ok := s.next()
		if ok {
			action(val)
		}
		return val, ok
	})
}
