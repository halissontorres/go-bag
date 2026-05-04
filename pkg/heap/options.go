package heap

import "github.com/halissontorres/go-bag/pkg/comparator"

// HeapOption configures a Heap at creation time.
type HeapOption[T any] func(*internalHeap[T])

// WithComparator overrides the comparator set at construction time.
func WithComparator[T any](cmp comparator.Comparator[T]) HeapOption[T] {
	return func(h *internalHeap[T]) {
		h.less = cmp
	}
}
