package streams

import "github.com/halissontorres/go-bag/bag/opts"

// Package streams provides lazy, sequential stream processing.
// Streams are NOT safe for concurrent use by multiple goroutines.
type Stream[T any] struct {
	next func() (T, bool)
}

// NewStream creates a Stream from a generator function.
func NewStream[T any](next func() (T, bool)) *Stream[T] {
	return &Stream[T]{next: next}
}

// FromSlice creates a Stream from a slice.
func FromSlice[T any](slice []T) *Stream[T] {
	i := 0
	return NewStream(func() (T, bool) {
		if i < len(slice) {
			val := slice[i]
			i++
			return val, true
		}
		var zero T
		return zero, false
	})
}

// FromChannel creates a Stream from a channel (reads until the channel is closed).
func FromChannel[T any](ch <-chan T) *Stream[T] {
	return NewStream(func() (T, bool) {
		val, ok := <-ch
		return val, ok
	})
}

// FromFunc creates a Stream from a generator function.
func FromFunc[T any](f func() (T, bool)) *Stream[T] {
	return NewStream(f)
}

// ToSlice collects all elements of the Stream into a slice.
func (s *Stream[T]) ToSlice() []T {
	result := make([]T, 0, 256)
	for {
		val, ok := s.next()
		if !ok {
			break
		}
		result = append(result, val)
	}
	return result
}

// ForEach applies a function to each element of the Stream.
func (s *Stream[T]) ForEach(f func(T)) {
	for {
		val, ok := s.next()
		if !ok {
			break
		}
		f(val)
	}
}

// Count returns the number of elements without allocating a slice.
func (s *Stream[T]) Count() int {
	count := 0
	for {
		_, ok := s.next()
		if !ok {
			break
		}
		count++
	}
	return count
}

// Any returns true if any element satisfies the predicate (short-circuit).
func (s *Stream[T]) Any(pred func(T) bool) bool {
	for {
		val, ok := s.next()
		if !ok {
			return false
		}
		if pred(val) {
			return true
		}
	}
}

// All returns true if all elements satisfy the predicate.
func (s *Stream[T]) All(pred func(T) bool) bool {
	for {
		val, ok := s.next()
		if !ok {
			return true
		}
		if !pred(val) {
			return false
		} // short-circuit!
	}
}

// Reduce combines elements using an accumulator function.
func (s *Stream[T]) Reduce(initial T, acc func(T, T) T) T {
	result := initial
	for {
		val, ok := s.next()
		if !ok {
			break
		}
		result = acc(result, val)
	}
	return result
}

// FindFirst returns the first element that satisfies the predicate as an Optional.
func FindFirst[T any](s *Stream[T], pred func(T) bool) opts.Optional[T] {
	for {
		val, ok := s.next()
		if !ok {
			return opts.Empty[T]()
		}
		if pred(val) {
			return opts.Of(val)
		}
	}
}
