package heap

import (
	"cmp"
	"container/heap"
	"errors"
	"sync"
)

// internalHeap is the structure that implements heap.Interface using Generics.
// The 'cmp.Ordered' constraint allows using < and > operators.
type internalHeap[T cmp.Ordered] []T

func (h internalHeap[T]) Len() int           { return len(h) }
func (h internalHeap[T]) Less(i, j int) bool { return h[i] < h[j] } // Min-Heap
func (h internalHeap[T]) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *internalHeap[T]) Push(x any) {
	*h = append(*h, x.(T))
}

func (h *internalHeap[T]) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// --- Generic Public Structure ---

// Heap represents a generic Priority Queue.
type Heap[T cmp.Ordered] struct {
	data *internalHeap[T]
}

// New creates a new instance of a generic Min-Heap.
func New[T cmp.Ordered]() *Heap[T] {
	h := &internalHeap[T]{}
	heap.Init(h)
	return &Heap[T]{data: h}
}

// Push adds a value of type T to the heap.
func (h *Heap[T]) Push(v T) {
	heap.Push(h.data, v)
}

// Pop removes and returns the smallest element (Min-Heap).
func (h *Heap[T]) Pop() (T, error) {
	var zero T
	if h.Len() == 0 {
		return zero, errors.New("empty heap")
	}
	val := heap.Pop(h.data).(T)
	return val, nil
}

// Peek returns the smallest element without removing it.
func (h *Heap[T]) Peek() (T, error) {
	var zero T
	if h.Len() == 0 {
		return zero, errors.New("empty heap")
	}
	return (*h.data)[0], nil
}

func (h *Heap[T]) Len() int {
	return h.data.Len()
}

// IsEmpty reports whether the heap is empty.
func (h *Heap[T]) IsEmpty() bool {
	return h.data.Len() == 0
}

// SyncHeap is a thread-safe Min-Heap.
type SyncHeap[T cmp.Ordered] struct {
	mu sync.RWMutex
	h  *Heap[T]
}

// NewSync creates a new thread-safe Min-Heap.
func NewSync[T cmp.Ordered]() *SyncHeap[T] {
	return &SyncHeap[T]{h: New[T]()}
}

func (sh *SyncHeap[T]) Push(v T) {
	sh.mu.Lock()
	defer sh.mu.Unlock()
	sh.h.Push(v)
}

func (sh *SyncHeap[T]) Pop() (T, error) {
	sh.mu.Lock()
	defer sh.mu.Unlock()
	return sh.h.Pop()
}

func (sh *SyncHeap[T]) Peek() (T, error) {
	sh.mu.RLock()
	defer sh.mu.RUnlock()
	return sh.h.Peek()
}

func (sh *SyncHeap[T]) Len() int {
	sh.mu.RLock()
	defer sh.mu.RUnlock()
	return sh.h.Len()
}

func (sh *SyncHeap[T]) IsEmpty() bool {
	sh.mu.RLock()
	defer sh.mu.RUnlock()
	return sh.h.IsEmpty()
}
