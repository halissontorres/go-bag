package streams

import (
	"cmp"
	"slices"
)

// Filter returns a Stream containing only the elements that satisfy the predicate.
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

// Map transforms each element of the Stream using function f.
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
			// gets next source element
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

// Distinct eliminates duplicates (keeps the first occurrence). Requires comparable.
func Distinct[T comparable](s *Stream[T]) *Stream[T] {
	seen := make(map[T]struct{}, 256)
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

// Sorted collects all elements, sorts them, and emits them. Requires Ordered.
func Sorted[T cmp.Ordered](s *Stream[T]) *Stream[T] {
	collected := s.ToSlice() // already optimized
	slices.Sort(collected)   // faster than sort.Slice for ordered types
	return FromSlice(collected)
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
		if !ok {
			var zero T
			return zero, false
		}
		count++
		return val, true
	})
}

// Skip ignores the first n elements.
func Skip[T any](s *Stream[T], n int) *Stream[T] {
	skipped := 0
	return NewStream(func() (T, bool) {
		for {
			val, ok := s.next()
			if !ok {
				var zero T
				return zero, false
			}
			if skipped < n {
				skipped++
				continue
			}
			return val, true
		}
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

// ToSliceWithCapacity collects all elements of the Stream into a slice with initial capacity hint
func (s *Stream[T]) ToSliceWithCapacity(hint int) []T {
	result := make([]T, 0, hint)
	for {
		val, ok := s.next()
		if !ok {
			break
		}
		result = append(result, val)
	}
	return result
}
