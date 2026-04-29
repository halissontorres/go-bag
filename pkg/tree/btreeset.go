package tree

import "cmp"

// BTreeSet is an ordered set backed by a B-Tree.
// Elements must satisfy the Ordered constraint (i.e. support <).
type BTreeSet[T cmp.Ordered] struct {
	tree *btree[T]
}

func NewBTreeSet[T cmp.Ordered]() *BTreeSet[T] {
	return &BTreeSet[T]{tree: newBTree[T](defaultMinDegree)}
}

func NewBTreeSetWithDegree[T cmp.Ordered](minDegree int) *BTreeSet[T] {
	return &BTreeSet[T]{tree: newBTree[T](minDegree)}
}

func (s *BTreeSet[T]) Add(value T) bool {
	return s.tree.insert(value)
}

func (s *BTreeSet[T]) Contains(value T) bool {
	node, _ := s.tree.search(value)
	return node != nil
}

func (s *BTreeSet[T]) Remove(value T) bool {
	// Simple implementation: rebuild the tree without the value.
	if !s.Contains(value) {
		return false
	}
	items := s.Elements()
	s.Clear()
	removed := false
	for _, item := range items {
		if item == value && !removed {
			removed = true
			continue
		}
		s.tree.insert(item)
	}
	return true
}

func (s *BTreeSet[T]) Len() int          { return s.tree.Len() }
func (s *BTreeSet[T]) IsEmpty() bool     { return s.Len() == 0 }
func (s *BTreeSet[T]) Clear()            { s.tree = newBTree[T](s.tree.minDegree) }
func (s *BTreeSet[T]) ForEach(f func(T)) { s.tree.traverse(f) }
func (s *BTreeSet[T]) Elements() []T {
	var result []T
	s.ForEach(func(v T) { result = append(result, v) })
	return result
}

func (s *BTreeSet[T]) Min() (T, bool)              { return s.tree.Min() }
func (s *BTreeSet[T]) Max() (T, bool)              { return s.tree.Max() }
func (s *BTreeSet[T]) Range(low, high T) []T       { return s.tree.Range(low, high) }
func (s *BTreeSet[T]) Iterator() *BTreeIterator[T] { return newBTreeIterator(s.tree) }
func (s *BTreeSet[T]) IteratorFrom(key T) *BTreeIterator[T] {
	return newBTreeIteratorFromKey(s.tree, key)
}
