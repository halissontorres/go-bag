package lists

import (
	"fmt"
	"strings"
	"sync"

	"github.com/halissontorres/go-bag/bag/streams"
)

// LinkedList is a generic doubly-linked list.
type LinkedList[T any] struct {
	head, tail *node[T]
	size       int
}

type node[T any] struct {
	value      T
	prev, next *node[T]
}

// NewLinkedList creates a new, empty list.
func NewLinkedList[T any]() *LinkedList[T] {
	return &LinkedList[T]{}
}

// NewLinkedListFromSlice creates a new list from the given slice elements.
func NewLinkedListFromSlice[T any](slice []T) *LinkedList[T] {
	l := NewLinkedList[T]()
	for _, v := range slice {
		l.AddLast(v)
	}
	return l
}

// Len returns the number of elements.
func (l *LinkedList[T]) Len() int { return l.size }

// IsEmpty returns whether the list is empty.
func (l *LinkedList[T]) IsEmpty() bool { return l.size == 0 }

// AddFirst inserts an element at the front.
func (l *LinkedList[T]) AddFirst(value T) {
	n := &node[T]{value: value}
	if l.head == nil {
		l.head = n
		l.tail = n
	} else {
		n.next = l.head
		l.head.prev = n
		l.head = n
	}
	l.size++
}

// PushFront is an alias for AddFirst.
func (l *LinkedList[T]) PushFront(value T) { l.AddFirst(value) }

// AddLast appends an element to the end.
func (l *LinkedList[T]) AddLast(value T) {
	n := &node[T]{value: value}
	if l.tail == nil {
		l.head = n
		l.tail = n
	} else {
		n.prev = l.tail
		l.tail.next = n
		l.tail = n
	}
	l.size++
}

// PushBack is an alias for AddLast.
func (l *LinkedList[T]) PushBack(value T) { l.AddLast(value) }

// Append is an alias for AddLast.
func (l *LinkedList[T]) Append(value T) { l.AddLast(value) }

// RemoveFirst removes and returns the first element.
func (l *LinkedList[T]) RemoveFirst() (T, bool) {
	if l.head == nil {
		var zero T
		return zero, false
	}
	val := l.head.value
	if l.head == l.tail {
		l.head = nil
		l.tail = nil
	} else {
		l.head = l.head.next
		l.head.prev = nil
	}
	l.size--
	return val, true
}

// PopFront is an alias for RemoveFirst.
func (l *LinkedList[T]) PopFront() (T, bool) { return l.RemoveFirst() }

// RemoveLast removes and returns the last element.
func (l *LinkedList[T]) RemoveLast() (T, bool) {
	if l.tail == nil {
		var zero T
		return zero, false
	}
	val := l.tail.value
	if l.head == l.tail {
		l.head = nil
		l.tail = nil
	} else {
		l.tail = l.tail.prev
		l.tail.next = nil
	}
	l.size--
	return val, true
}

// PopBack is an alias for RemoveLast.
func (l *LinkedList[T]) PopBack() (T, bool) { return l.RemoveLast() }

// First returns the first element without removing it.
func (l *LinkedList[T]) First() (T, bool) {
	if l.head == nil {
		var zero T
		return zero, false
	}
	return l.head.value, true
}

// Last returns the last element without removing it.
func (l *LinkedList[T]) Last() (T, bool) {
	if l.tail == nil {
		var zero T
		return zero, false
	}
	return l.tail.value, true
}

// Get returns the element at the given 0-based index.
// Walks from whichever end is closer, so the cost is O(min(index, n-index)).
func (l *LinkedList[T]) Get(index int) (T, bool) {
	if index < 0 || index >= l.size {
		var zero T
		return zero, false
	}
	return l.nodeAt(index).value, true
}

// InsertAt inserts value at the given index, shifting subsequent elements right.
// Walks from whichever end is closer, so the cost is O(min(index, n-index)).
func (l *LinkedList[T]) InsertAt(index int, value T) bool {
	if index < 0 || index > l.size {
		return false
	}
	if index == 0 {
		l.AddFirst(value)
		return true
	}
	if index == l.size {
		l.AddLast(value)
		return true
	}

	curr := l.nodeAt(index)
	n := &node[T]{value: value, prev: curr.prev, next: curr}
	curr.prev.next = n
	curr.prev = n
	l.size++
	return true
}

// RemoveAt removes and returns the element at the given index.
// Walks from whichever end is closer, so the cost is O(min(index, n-index)).
func (l *LinkedList[T]) RemoveAt(index int) (T, bool) {
	if index < 0 || index >= l.size {
		var zero T
		return zero, false
	}
	if index == 0 {
		return l.RemoveFirst()
	}
	if index == l.size-1 {
		return l.RemoveLast()
	}

	curr := l.nodeAt(index)
	curr.prev.next = curr.next
	curr.next.prev = curr.prev
	l.size--
	return curr.value, true
}

// nodeAt returns the node at the given index by walking from the closer end.
// The caller must ensure 0 <= index < size.
func (l *LinkedList[T]) nodeAt(index int) *node[T] {
	if index < l.size/2 {
		curr := l.head
		for i := 0; i < index; i++ {
			curr = curr.next
		}
		return curr
	}
	curr := l.tail
	for i := l.size - 1; i > index; i-- {
		curr = curr.prev
	}
	return curr
}

// ForEach applies f to each element in order.
func (l *LinkedList[T]) ForEach(f func(T)) {
	for curr := l.head; curr != nil; curr = curr.next {
		f(curr.value)
	}
}

// Elements returns a slice with all elements in order.
func (l *LinkedList[T]) Elements() []T {
	result := make([]T, 0, l.size)
	l.ForEach(func(v T) { result = append(result, v) })
	return result
}

// Clear removes all elements.
func (l *LinkedList[T]) Clear() {
	l.head = nil
	l.tail = nil
	l.size = 0
}

// String returns a human-readable representation of the list, e.g. "[1, 2, 3]".
func (l *LinkedList[T]) String() string {
	if l.size == 0 {
		return "[]"
	}
	var b strings.Builder
	b.WriteByte('[')
	for curr := l.head; curr != nil; curr = curr.next {
		_, err := fmt.Fprintf(&b, "%v", curr.value)
		if err != nil {
			return ""
		}
		if curr.next != nil {
			b.WriteString(", ")
		}
	}
	b.WriteByte(']')
	return b.String()
}

// Stream returns a stream of the list elements.
func (l *LinkedList[T]) Stream() *streams.Stream[T] {
	it := l.Iter()
	return streams.NewStream(func() (T, bool) {
		return it.Next()
	})
}

// Iterator is a generic type for traversing over elements in a sequence.
// It maintains the current position for sequential access.
type Iterator[T any] struct {
	l       *LinkedList[T]
	curr    *node[T]
	forward bool
}

// Iter returns an iterator to traverse the elements of the linked list sequentially from the beginning.
func (l *LinkedList[T]) Iter() *Iterator[T] {
	return &Iterator[T]{l: l, curr: l.head, forward: true}
}

// ReverseIter returns an iterator to traverse the elements of the linked list in reverse order.
func (l *LinkedList[T]) ReverseIter() *Iterator[T] {
	return &Iterator[T]{l: l, curr: l.tail, forward: false}
}

// Next retrieves the next element and advances the iterator. Returns false if no more elements exist.
func (it *Iterator[T]) Next() (T, bool) {
	if it.curr == nil {
		var zero T
		return zero, false
	}
	val := it.curr.value
	if it.forward {
		it.curr = it.curr.next
	} else {
		it.curr = it.curr.prev
	}
	return val, true
}

// Prev retrieves the previous element and moves the iterator back. Returns false if no more elements exist.
func (it *Iterator[T]) Prev() (T, bool) {
	var n *node[T]
	if it.curr == nil {
		if it.forward {
			n = it.l.tail
		} else {
			n = it.l.head
		}
	} else {
		if it.forward {
			n = it.curr.prev
		} else {
			n = it.curr.next
		}
	}

	if n == nil {
		var zero T
		return zero, false
	}
	it.curr = n
	return n.value, true
}

// Reset resets the iterator to its initial position.
func (it *Iterator[T]) Reset() {
	if it.forward {
		it.curr = it.l.head
	} else {
		it.curr = it.l.tail
	}
}

// SyncLinkedList is a thread-safe wrapper for LinkedList.
type SyncLinkedList[T any] struct {
	mu sync.RWMutex
	l  *LinkedList[T]
}

// NewSyncLinkedList creates a new, empty thread-safe list.
func NewSyncLinkedList[T any]() *SyncLinkedList[T] {
	return &SyncLinkedList[T]{l: NewLinkedList[T]()}
}

// AddFirst inserts an element at the front.
func (sl *SyncLinkedList[T]) AddFirst(value T) {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	sl.l.AddFirst(value)
}

// PushFront is an alias for AddFirst.
func (sl *SyncLinkedList[T]) PushFront(value T) { sl.AddFirst(value) }

// AddLast appends an element to the end.
func (sl *SyncLinkedList[T]) AddLast(value T) {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	sl.l.AddLast(value)
}

// PushBack is an alias for AddLast.
func (sl *SyncLinkedList[T]) PushBack(value T) { sl.AddLast(value) }

// Append is an alias for AddLast.
func (sl *SyncLinkedList[T]) Append(value T) { sl.AddLast(value) }

// RemoveFirst removes and returns the first element.
func (sl *SyncLinkedList[T]) RemoveFirst() (T, bool) {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	return sl.l.RemoveFirst()
}

// PopFront is an alias for RemoveFirst.
func (sl *SyncLinkedList[T]) PopFront() (T, bool) { return sl.RemoveFirst() }

// RemoveLast removes and returns the last element.
func (sl *SyncLinkedList[T]) RemoveLast() (T, bool) {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	return sl.l.RemoveLast()
}

// PopBack is an alias for RemoveLast.
func (sl *SyncLinkedList[T]) PopBack() (T, bool) { return sl.RemoveLast() }

// First returns the first element without removing it.
func (sl *SyncLinkedList[T]) First() (T, bool) {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	return sl.l.First()
}

// Last returns the last element without removing it.
func (sl *SyncLinkedList[T]) Last() (T, bool) {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	return sl.l.Last()
}

// Get returns the element at the given 0-based index.
func (sl *SyncLinkedList[T]) Get(index int) (T, bool) {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	return sl.l.Get(index)
}

// Len returns the number of elements.
func (sl *SyncLinkedList[T]) Len() int {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	return sl.l.Len()
}

// IsEmpty returns whether the list is empty.
func (sl *SyncLinkedList[T]) IsEmpty() bool {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	return sl.l.IsEmpty()
}

// InsertAt inserts value at the given index.
func (sl *SyncLinkedList[T]) InsertAt(index int, value T) bool {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	return sl.l.InsertAt(index, value)
}

// RemoveAt removes and returns the element at the given index.
func (sl *SyncLinkedList[T]) RemoveAt(index int) (T, bool) {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	return sl.l.RemoveAt(index)
}

// Elements returns a slice with all elements in order.
func (sl *SyncLinkedList[T]) Elements() []T {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	return sl.l.Elements()
}

// Clear removes all elements.
func (sl *SyncLinkedList[T]) Clear() {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	sl.l.Clear()
}

// String returns a human-readable representation of the list.
func (sl *SyncLinkedList[T]) String() string {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	return sl.l.String()
}
