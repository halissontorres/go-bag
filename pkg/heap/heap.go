package heap

import (
	"container/heap"
	"errors"
	"sync"

	"github.com/halissontorres/go-bag/pkg/comparator"
)

// internalHeap implements heap.Interface.
type internalHeap[T any] struct {
	data []T
	less comparator.Comparator[T]
}

func (h *internalHeap[T]) Len() int           { return len(h.data) }
func (h *internalHeap[T]) Less(i, j int) bool { return h.less(h.data[i], h.data[j]) }
func (h *internalHeap[T]) Swap(i, j int)      { h.data[i], h.data[j] = h.data[j], h.data[i] }
func (h *internalHeap[T]) Push(x any)         { h.data = append(h.data, x.(T)) }
func (h *internalHeap[T]) Pop() any {
	old := h.data
	n := len(old)
	x := old[n-1]
	var zero T
	old[n-1] = zero // ajuda o GC
	h.data = old[:n-1]
	return x
}

// Heap represents a generic Priority Queue.
type Heap[T any] struct {
	data *internalHeap[T]
}

// New creates a Heap ordered by the provided Comparator.
// The comparator is mandatory — there is no default ordering.
func New[T any](cmp comparator.Comparator[T], opts ...HeapOption[T]) *Heap[T] {
	h := &internalHeap[T]{
		less: cmp,
	}
	for _, o := range opts {
		o(h)
	}
	heap.Init(h)
	return &Heap[T]{data: h}
}

func (h *Heap[T]) Push(v T) {
	heap.Push(h.data, v)
}

func (h *Heap[T]) Pop() (T, error) {
	var zero T
	if h.Len() == 0 {
		return zero, errors.New("empty heap")
	}
	return heap.Pop(h.data).(T), nil
}

func (h *Heap[T]) Peek() (T, error) {
	var zero T
	if h.Len() == 0 {
		return zero, errors.New("empty heap")
	}
	return h.data.data[0], nil
}

func (h *Heap[T]) Len() int      { return h.data.Len() }
func (h *Heap[T]) IsEmpty() bool { return h.data.Len() == 0 }

// SyncHeap is a thread-safe Heap.
type SyncHeap[T any] struct {
	mu sync.RWMutex
	h  *Heap[T]
}

// NewSync creates a thread-safe Heap ordered by the provided Comparator.
func NewSync[T any](cmp comparator.Comparator[T], opts ...HeapOption[T]) *SyncHeap[T] {
	return &SyncHeap[T]{h: New[T](cmp, opts...)}
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
