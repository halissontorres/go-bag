package lists

import "sync"

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

// Len returns the number of elements.
func (l *LinkedList[T]) Len() int { return l.size }

// IsEmpty returns whether the list is empty.
func (l *LinkedList[T]) IsEmpty() bool { return l.size == 0 }

// AddFirst inserts an element at the front (equivalent to PushFront).
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

// AddLast appends an element to the end (equivalent to Append/PushBack).
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

type SyncLinkedList[T any] struct {
	mu sync.RWMutex
	l  *LinkedList[T]
}

func NewSyncLinkedList[T any]() *SyncLinkedList[T] {
	return &SyncLinkedList[T]{l: NewLinkedList[T]()}
}

func (sl *SyncLinkedList[T]) AddFirst(value T) {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	sl.l.AddFirst(value)
}

func (sl *SyncLinkedList[T]) AddLast(value T) {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	sl.l.AddLast(value)
}

func (sl *SyncLinkedList[T]) RemoveFirst() (T, bool) {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	return sl.l.RemoveFirst()
}

func (sl *SyncLinkedList[T]) RemoveLast() (T, bool) {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	return sl.l.RemoveLast()
}

func (sl *SyncLinkedList[T]) First() (T, bool) {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	return sl.l.First()
}

func (sl *SyncLinkedList[T]) Last() (T, bool) {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	return sl.l.Last()
}

func (sl *SyncLinkedList[T]) Get(index int) (T, bool) {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	return sl.l.Get(index)
}

func (sl *SyncLinkedList[T]) Len() int {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	return sl.l.Len()
}

func (sl *SyncLinkedList[T]) IsEmpty() bool {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	return sl.l.IsEmpty()
}

func (sl *SyncLinkedList[T]) InsertAt(index int, value T) bool {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	return sl.l.InsertAt(index, value)
}

func (sl *SyncLinkedList[T]) RemoveAt(index int) (T, bool) {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	return sl.l.RemoveAt(index)
}

func (sl *SyncLinkedList[T]) Elements() []T {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	return sl.l.Elements()
}

func (sl *SyncLinkedList[T]) Clear() {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	sl.l.Clear()
}
