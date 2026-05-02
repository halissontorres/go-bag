package heap

import (
	"cmp"
	"container/heap"
	"errors"
	"sync"
)

// internalHeap implements heap.Interface.
type internalHeap[T cmp.Ordered] struct {
	data []T
	less func(a, b T) bool
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
type Heap[T cmp.Ordered] struct {
	data *internalHeap[T]
}

// New cria um Heap. Padrão: Min-Heap.
func New[T cmp.Ordered](opts ...HeapOption[T]) *Heap[T] {
	h := &internalHeap[T]{
		less: func(a, b T) bool { return a < b }, // padrão Min-Heap
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

// SyncHeap é uma Heap thread-safe.
type SyncHeap[T cmp.Ordered] struct {
	mu sync.RWMutex
	h  *Heap[T]
}

func NewSync[T cmp.Ordered](opts ...HeapOption[T]) *SyncHeap[T] {
	return &SyncHeap[T]{h: New[T](opts...)}
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
