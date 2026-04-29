package tree

import (
	"math/rand/v2"
	"slices"
	"testing"
)

func TestBTreeSet_AddContainsRemove(t *testing.T) {
	t.Parallel()

	s := NewBTreeSet[int]()
	if !s.IsEmpty() {
		t.Fatalf("expected empty set")
	}
	values := []int{10, 20, 5, 6, 12, 30, 7, 17}
	for _, v := range values {
		if !s.Add(v) {
			t.Fatalf("Add(%d) should be true", v)
		}
	}
	// Adding a duplicate must be a no-op.
	if s.Add(10) {
		t.Fatalf("Add(10) duplicate should return false")
	}
	if got, want := s.Len(), len(values); got != want {
		t.Fatalf("len=%d want %d", got, want)
	}
	for _, v := range values {
		if !s.Contains(v) {
			t.Fatalf("missing %d", v)
		}
	}
	if s.Contains(999) {
		t.Fatalf("should not contain 999")
	}

	// Inorder traversal must be sorted.
	got := s.Elements()
	want := slices.Clone(values)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Fatalf("Elements=%v want %v", got, want)
	}
}

func TestBTreeSet_RemoveAcrossSplits(t *testing.T) {
	t.Parallel()

	s := NewBTreeSetWithDegree[int](3)
	const N = 200
	for i := 0; i < N; i++ {
		s.Add(i)
	}
	// Remove every other element and check ordering and length invariants.
	for i := 0; i < N; i += 2 {
		if !s.Remove(i) {
			t.Fatalf("Remove(%d) should return true", i)
		}
	}
	if s.Remove(0) {
		t.Fatalf("Remove(0) on missing should return false")
	}
	if got, want := s.Len(), N/2; got != want {
		t.Fatalf("len=%d want %d", got, want)
	}
	got := s.Elements()
	want := make([]int, 0, N/2)
	for i := 1; i < N; i += 2 {
		want = append(want, i)
	}
	if !slices.Equal(got, want) {
		t.Fatalf("post-remove elements mismatch")
	}
}

func TestBTreeSet_MinMaxRange(t *testing.T) {
	t.Parallel()

	s := NewBTreeSet[int]()
	for _, v := range []int{50, 10, 80, 30, 20, 60, 90, 40, 70} {
		s.Add(v)
	}
	if v, ok := s.Min(); !ok || v != 10 {
		t.Fatalf("Min=%d,%v want 10,true", v, ok)
	}
	if v, ok := s.Max(); !ok || v != 90 {
		t.Fatalf("Max=%d,%v want 90,true", v, ok)
	}
	if got, want := s.Range(25, 65), []int{30, 40, 50, 60}; !slices.Equal(got, want) {
		t.Fatalf("Range=%v want %v", got, want)
	}

	empty := NewBTreeSet[int]()
	if _, ok := empty.Min(); ok {
		t.Fatalf("Min on empty should be false")
	}
	if _, ok := empty.Max(); ok {
		t.Fatalf("Max on empty should be false")
	}
	if got := empty.Range(0, 10); len(got) != 0 {
		t.Fatalf("Range on empty=%v want empty", got)
	}
}

func TestBTreeSet_Clear(t *testing.T) {
	t.Parallel()

	s := NewBTreeSet[int]()
	for i := 0; i < 50; i++ {
		s.Add(i)
	}
	s.Clear()
	if !s.IsEmpty() {
		t.Fatalf("expected empty after Clear")
	}
	s.Add(7)
	if !s.Contains(7) || s.Len() != 1 {
		t.Fatalf("post-Clear add failed")
	}
}

func TestBTreeSet_RandomizedAgainstSlice(t *testing.T) {
	t.Parallel()

	r := rand.New(rand.NewPCG(42, 99))
	s := NewBTreeSetWithDegree[int](3)
	ref := make(map[int]struct{})
	const ops = 5_000

	for i := 0; i < ops; i++ {
		v := r.IntN(1000)
		switch r.IntN(3) {
		case 0, 1: // add
			_, exists := ref[v]
			added := s.Add(v)
			if added == exists {
				t.Fatalf("Add(%d) returned %v, ref exists=%v", v, added, exists)
			}
			ref[v] = struct{}{}
		case 2: // remove
			_, exists := ref[v]
			removed := s.Remove(v)
			if removed != exists {
				t.Fatalf("Remove(%d) returned %v, ref exists=%v", v, removed, exists)
			}
			delete(ref, v)
		}
	}

	if got, want := s.Len(), len(ref); got != want {
		t.Fatalf("len=%d want %d", got, want)
	}
	for k := range ref {
		if !s.Contains(k) {
			t.Fatalf("set should contain %d", k)
		}
	}
	// In-order traversal must be sorted.
	got := s.Elements()
	if !slices.IsSorted(got) {
		t.Fatalf("Elements not sorted")
	}
}

// ---------- Benchmarks ----------

func BenchmarkBTreeSet_Add(b *testing.B) {
	b.ReportAllocs()
	s := NewBTreeSet[int]()
	i := 0
	for b.Loop() {
		s.Add(i)
		i++
	}
}

func BenchmarkBTreeSet_Contains(b *testing.B) {
	const N = 10_000
	s := NewBTreeSet[int]()
	for i := 0; i < N; i++ {
		s.Add(i)
	}
	b.ReportAllocs()
	b.ResetTimer()
	i := 0
	for b.Loop() {
		_ = s.Contains(i % N)
		i++
	}
}

func BenchmarkBTreeSet_Range(b *testing.B) {
	const N = 10_000
	s := NewBTreeSet[int]()
	for i := 0; i < N; i++ {
		s.Add(i)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = s.Range(N/4, 3*N/4)
	}
}
