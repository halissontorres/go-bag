package queue

import (
	"fmt"
	"sync"
)

// Deque is a generic double-ended queue.
// It is backed by a circular slice for consistent performance.
// The zero value is not usable; use NewDeque.
type Deque[T any] struct {
	buf      []T
	head     int // index of the first element
	tail     int // index just past the last element
	size     int
	capacity int
}

// NewDeque creates a new, empty Deque with an initial capacity of 16.
func NewDeque[T any]() *Deque[T] {
	return NewDequeWithCap[T](16)
}

// NewDequeWithCap creates an empty Deque with the given initial capacity.
func NewDequeWithCap[T any](cap int) *Deque[T] {
	if cap < 1 {
		cap = 1
	}
	return &Deque[T]{
		buf:      make([]T, cap),
		capacity: cap,
	}
}

// resize reallocates the underlying buffer when capacity is exceeded.
func (d *Deque[T]) resize(newCap int) {
	newBuf := make([]T, newCap)
	if d.size > 0 {
		if d.head < d.tail {
			copy(newBuf, d.buf[d.head:d.tail])
		} else {
			n := copy(newBuf, d.buf[d.head:])
			copy(newBuf[n:], d.buf[:d.tail])
		}
	}
	d.buf = newBuf
	d.head = 0
	d.tail = d.size
	d.capacity = newCap
}

// grow doubles the capacity if needed.
func (d *Deque[T]) grow() {
	if d.size == d.capacity {
		d.resize(d.capacity * 2)
	}
}

// shrink halves the capacity when usage drops to one quarter.
func (d *Deque[T]) shrink() {
	if d.capacity > 16 && d.size <= d.capacity/4 {
		d.resize(d.capacity / 2)
	}
}

// Len returns the number of elements.
func (d *Deque[T]) Len() int { return d.size }

// IsEmpty reports whether the deque is empty.
func (d *Deque[T]) IsEmpty() bool { return d.size == 0 }

// PushFront inserts an element at the front.
func (d *Deque[T]) PushFront(value T) {
	d.grow()
	d.head = (d.head - 1 + d.capacity) % d.capacity
	d.buf[d.head] = value
	d.size++
}

// PushBack inserts an element at the back.
func (d *Deque[T]) PushBack(value T) {
	d.grow()
	d.buf[d.tail] = value
	d.tail = (d.tail + 1) % d.capacity
	d.size++
}

// PopFront removes and returns the first element.
func (d *Deque[T]) PopFront() (T, bool) {
	if d.size == 0 {
		var zero T
		return zero, false
	}
	value := d.buf[d.head]
	var zero T
	d.buf[d.head] = zero // help the GC reclaim referenced memory.
	d.head = (d.head + 1) % d.capacity
	d.size--
	d.shrink()
	return value, true
}

// PopBack removes and returns the last element.
func (d *Deque[T]) PopBack() (T, bool) {
	if d.size == 0 {
		var zero T
		return zero, false
	}
	d.tail = (d.tail - 1 + d.capacity) % d.capacity
	value := d.buf[d.tail]
	var zero T
	d.buf[d.tail] = zero
	d.size--
	d.shrink()
	return value, true
}

// PeekFront returns the first element without removing it.
func (d *Deque[T]) PeekFront() (T, bool) {
	if d.size == 0 {
		var zero T
		return zero, false
	}
	return d.buf[d.head], true
}

// PeekBack returns the last element without removing it.
func (d *Deque[T]) PeekBack() (T, bool) {
	if d.size == 0 {
		var zero T
		return zero, false
	}
	idx := (d.tail - 1 + d.capacity) % d.capacity
	return d.buf[idx], true
}

// Clear empties the deque.
func (d *Deque[T]) Clear() {
	for i := 0; i < d.size; i++ {
		var zero T
		d.buf[(d.head+i)%d.capacity] = zero
	}
	d.head = 0
	d.tail = 0
	d.size = 0
}

// Elements returns a slice with all elements from front to back.
func (d *Deque[T]) Elements() []T {
	if d.size == 0 {
		return nil
	}
	result := make([]T, d.size)
	if d.head < d.tail {
		copy(result, d.buf[d.head:d.tail])
	} else {
		n := copy(result, d.buf[d.head:])
		copy(result[n:], d.buf[:d.tail])
	}
	return result
}

// String returns a textual representation of the deque.
func (d *Deque[T]) String() string {
	return fmt.Sprintf("%v", d.Elements())
}

// SyncDeque is a thread-safe Deque.
type SyncDeque[T any] struct {
	mu sync.RWMutex
	d  *Deque[T]
}

func NewSyncDeque[T any]() *SyncDeque[T] {
	return &SyncDeque[T]{d: NewDeque[T]()}
}

func (sd *SyncDeque[T]) PushFront(val T) { sd.mu.Lock(); defer sd.mu.Unlock(); sd.d.PushFront(val) }
func (sd *SyncDeque[T]) PushBack(val T)  { sd.mu.Lock(); defer sd.mu.Unlock(); sd.d.PushBack(val) }
func (sd *SyncDeque[T]) PopFront() (T, bool) {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	return sd.d.PopFront()
}
func (sd *SyncDeque[T]) PopBack() (T, bool) {
	sd.mu.Lock()
	defer sd.mu.Unlock()
	return sd.d.PopBack()
}
func (sd *SyncDeque[T]) PeekFront() (T, bool) {
	sd.mu.RLock()
	defer sd.mu.RUnlock()
	return sd.d.PeekFront()
}
func (sd *SyncDeque[T]) PeekBack() (T, bool) {
	sd.mu.RLock()
	defer sd.mu.RUnlock()
	return sd.d.PeekBack()
}
func (sd *SyncDeque[T]) Len() int      { sd.mu.RLock(); defer sd.mu.RUnlock(); return sd.d.Len() }
func (sd *SyncDeque[T]) IsEmpty() bool { sd.mu.RLock(); defer sd.mu.RUnlock(); return sd.d.IsEmpty() }
