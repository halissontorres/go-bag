package queue

import "sync"

// Queue is a generic FIFO queue.
// The zero value is not usable; use NewQueue.
type Queue[T any] struct {
	items []T // underlying slice
	head  int // index of the first element
	tail  int // index of the last element
	size  int // number of elements
}

// NewQueue creates a new, empty queue.
func NewQueue[T any](opts ...Option) *Queue[T] {
	c := &options{initialCap: defaultInitialCap}
	for _, opt := range opts {
		opt(c)
	}
	return &Queue[T]{
		items: make([]T, 0, c.initialCap),
	}
}

// Enqueue appends an element to the back of the queue.
func (q *Queue[T]) Enqueue(val T) {
	q.items = append(q.items, val)
	q.tail++
	q.size++
}

// Dequeue removes and returns the front element.
// The second return value is false if the queue is empty.
func (q *Queue[T]) Dequeue() (T, bool) {
	if q.size == 0 {
		var zero T
		return zero, false
	}
	val := q.items[q.head]
	q.head++
	q.size--
	// Compact the underlying slice once the head has drifted far enough.
	if q.head > len(q.items)/2 {
		q.compact()
	}
	return val, true
}

// Peek returns the front element without removing it.
func (q *Queue[T]) Peek() (T, bool) {
	if q.size == 0 {
		var zero T
		return zero, false
	}
	return q.items[q.head], true
}

// Len returns the number of elements in the queue.
func (q *Queue[T]) Len() int {
	return q.size
}

// IsEmpty reports whether the queue is empty.
func (q *Queue[T]) IsEmpty() bool {
	return q.size == 0
}

// Clear empties the queue.
func (q *Queue[T]) Clear() {
	q.items = q.items[:0]
	q.head = 0
	q.tail = 0
	q.size = 0
}

// Elements returns a slice with all elements in queue order
// (oldest to newest).
func (q *Queue[T]) Elements() []T {
	if q.size == 0 {
		return nil
	}
	result := make([]T, q.size)
	copy(result, q.items[q.head:q.tail])
	return result
}

// compact shrinks the slice once the head is far from the start.
func (q *Queue[T]) compact() {
	if q.size == 0 {
		q.items = q.items[:0]
		q.head = 0
		q.tail = 0
		return
	}
	newItems := make([]T, q.size)
	copy(newItems, q.items[q.head:q.tail])
	q.items = newItems
	q.head = 0
	q.tail = q.size
}

// SyncQueue is a thread-safe Queue.
type SyncQueue[T any] struct {
	mu sync.RWMutex
	q  *Queue[T]
}

func NewSyncQueue[T any](opts ...Option) *SyncQueue[T] {
	return &SyncQueue[T]{q: NewQueue[T](opts...)}
}

func (sq *SyncQueue[T]) Enqueue(val T) {
	sq.mu.Lock()
	defer sq.mu.Unlock()
	sq.q.Enqueue(val)
}

func (sq *SyncQueue[T]) Dequeue() (T, bool) {
	sq.mu.Lock()
	defer sq.mu.Unlock()
	return sq.q.Dequeue()
}

func (sq *SyncQueue[T]) Peek() (T, bool) {
	sq.mu.RLock()
	defer sq.mu.RUnlock()
	return sq.q.Peek()
}

func (sq *SyncQueue[T]) Len() int {
	sq.mu.RLock()
	defer sq.mu.RUnlock()
	return sq.q.Len()
}

func (sq *SyncQueue[T]) IsEmpty() bool {
	sq.mu.RLock()
	defer sq.mu.RUnlock()
	return sq.q.IsEmpty()
}
