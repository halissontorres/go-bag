package stack

import "sync"

// Stack is a generic LIFO stack.
// The zero value is not usable; use NewStack.
type Stack[T any] struct {
	items []T
}

// NewStack creates a new, empty stack.
func NewStack[T any](opts ...Option) *Stack[T] {
	c := applyStackOptions(opts)
	return &Stack[T]{
		items: make([]T, 0, c.initialCap),
	}
}

// Push pushes an element onto the top of the stack.
func (s *Stack[T]) Push(val T) {
	s.items = append(s.items, val)
}

// Pop removes and returns the top element.
// The second return value is false if the stack is empty.
func (s *Stack[T]) Pop() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}
	index := len(s.items) - 1
	val := s.items[index]
	var zero T
	s.items[index] = zero // help the GC reclaim referenced memory
	s.items = s.items[:index]
	return val, true
}

// Peek returns the top element without removing it.
func (s *Stack[T]) Peek() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}
	return s.items[len(s.items)-1], true
}

// Len returns the number of elements in the stack.
func (s *Stack[T]) Len() int {
	return len(s.items)
}

// IsEmpty reports whether the stack is empty.
func (s *Stack[T]) IsEmpty() bool {
	return len(s.items) == 0
}

// Clear empties the stack.
func (s *Stack[T]) Clear() {
	s.items = s.items[:0]
}

// Elements returns a slice with all elements in stack order (bottom to top).
func (s *Stack[T]) Elements() []T {
	if len(s.items) == 0 {
		return nil
	}
	result := make([]T, len(s.items))
	copy(result, s.items)
	return result
}

// SyncStack is a thread-safe Stack.
type SyncStack[T any] struct {
	mu sync.RWMutex
	s  *Stack[T]
}

// NewSyncStack creates a new, empty thread-safe stack.
func NewSyncStack[T any](opts ...Option) *SyncStack[T] {
	return &SyncStack[T]{s: NewStack[T](opts...)}
}

func (ss *SyncStack[T]) Push(val T) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.s.Push(val)
}

func (ss *SyncStack[T]) Pop() (T, bool) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	return ss.s.Pop()
}

func (ss *SyncStack[T]) Peek() (T, bool) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.s.Peek()
}

func (ss *SyncStack[T]) Len() int {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.s.Len()
}

func (ss *SyncStack[T]) IsEmpty() bool {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.s.IsEmpty()
}
