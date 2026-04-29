package tree

import (
	"slices"
	"testing"
)

func TestBTreeIterator_EmptyTree(t *testing.T) {
	t.Parallel()

	s := NewBTreeSet[int]()
	it := s.Iterator()
	if _, ok := it.Next(); ok {
		t.Fatalf("Next on empty iterator should return false")
	}
}

func TestBTreeIterator_InOrderTraversal(t *testing.T) {
	t.Parallel()

	s := NewBTreeSetWithDegree[int](3)
	values := []int{50, 10, 80, 30, 20, 60, 90, 40, 70, 5, 15, 25, 35, 45, 55, 65, 75, 85, 95}
	for _, v := range values {
		s.Add(v)
	}
	want := slices.Clone(values)
	slices.Sort(want)

	it := s.Iterator()
	var got []int
	for {
		v, ok := it.Next()
		if !ok {
			break
		}
		got = append(got, v)
	}
	if !slices.Equal(got, want) {
		t.Fatalf("iterator order=%v want %v", got, want)
	}
}

func TestBTreeIterator_FromKey(t *testing.T) {
	t.Parallel()

	s := NewBTreeSetWithDegree[int](3)
	for _, v := range []int{10, 20, 30, 40, 50, 60, 70, 80, 90} {
		s.Add(v)
	}

	it := s.IteratorFrom(35)
	var got []int
	for {
		v, ok := it.Next()
		if !ok {
			break
		}
		got = append(got, v)
	}
	if want := []int{40, 50, 60, 70, 80, 90}; !slices.Equal(got, want) {
		t.Fatalf("IteratorFrom(35)=%v want %v", got, want)
	}

	// IteratorFrom past the max must yield no results.
	tail := s.IteratorFrom(1000)
	if _, ok := tail.Next(); ok {
		t.Fatalf("IteratorFrom past max should be empty")
	}

	// IteratorFrom on an empty tree should be empty.
	empty := NewBTreeSet[int]()
	if _, ok := empty.IteratorFrom(0).Next(); ok {
		t.Fatalf("IteratorFrom on empty should be empty")
	}
}

func TestBTreeRange_Boundaries(t *testing.T) {
	t.Parallel()

	s := NewBTreeSet[int]()
	for _, v := range []int{1, 3, 5, 7, 9} {
		s.Add(v)
	}
	if got, want := s.Range(3, 7), []int{3, 5, 7}; !slices.Equal(got, want) {
		t.Fatalf("Range(3,7)=%v want %v", got, want)
	}
	if got := s.Range(20, 30); len(got) != 0 {
		t.Fatalf("Range above max should be empty, got %v", got)
	}
	if got, want := s.Range(0, 100), []int{1, 3, 5, 7, 9}; !slices.Equal(got, want) {
		t.Fatalf("Range(0,100)=%v want %v", got, want)
	}
}
