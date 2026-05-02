package heap

import "cmp"

// HeapOption configura um Heap na criação.
type HeapOption[T cmp.Ordered] func(*internalHeap[T])

func WithMinHeap[T cmp.Ordered]() HeapOption[T] {
	return func(h *internalHeap[T]) {
		h.less = func(a, b T) bool { return a < b }
	}
}

func WithMaxHeap[T cmp.Ordered]() HeapOption[T] {
	return func(h *internalHeap[T]) {
		h.less = func(a, b T) bool { return a > b }
	}
}

func WithLessFunc[T cmp.Ordered](less func(a, b T) bool) HeapOption[T] {
	return func(h *internalHeap[T]) {
		h.less = less
	}
}
