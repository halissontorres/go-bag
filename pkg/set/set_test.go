package set

import (
	"slices"
	"sync"
	"testing"
)

func TestSet_AddContainsRemove(t *testing.T) {
	t.Parallel()

	s := NewSet[int]()
	s.Add(1, 2, 3, 2)
	if got, want := s.Len(), 3; got != want {
		t.Fatalf("len=%d want %d", got, want)
	}
	if !s.Contains(1) || !s.Contains(2) || !s.Contains(3) {
		t.Fatalf("missing element")
	}
	if s.Contains(99) {
		t.Fatalf("should not contain 99")
	}

	s.Remove(2, 99)
	if s.Contains(2) {
		t.Fatalf("Remove(2) failed")
	}
	if got, want := s.Len(), 2; got != want {
		t.Fatalf("len after remove=%d want %d", got, want)
	}
}

func TestSet_ZeroValueIsUsableViaConstructor(t *testing.T) {
	t.Parallel()

	// Cover the lazy-init paths on a constructed Set whose internal map was zeroed.
	s := &Set[int]{}
	if s.Len() != 0 {
		t.Fatalf("zero set len=%d want 0", s.Len())
	}
	if s.Contains(1) {
		t.Fatalf("zero set should not contain anything")
	}
	if got := s.Elements(); got != nil {
		t.Fatalf("zero set Elements=%v want nil", got)
	}
	s.Remove(1) // must not panic
	s.Add(1)    // triggers lazy init
	if !s.Contains(1) {
		t.Fatalf("Add after zero-value failed")
	}
	s.Clear()
	if s.Len() != 0 {
		t.Fatalf("Clear failed")
	}
}

func TestSet_Elements(t *testing.T) {
	t.Parallel()

	s := NewSet[int]()
	s.Add(3, 1, 2)
	got := s.Elements()
	slices.Sort(got)
	if want := []int{1, 2, 3}; !slices.Equal(got, want) {
		t.Fatalf("Elements=%v want %v", got, want)
	}
}

func TestSet_Operations(t *testing.T) {
	t.Parallel()

	a := NewSet[int]()
	a.Add(1, 2, 3)
	b := NewSet[int]()
	b.Add(2, 3, 4)

	un := Union(a, b).Elements()
	slices.Sort(un)
	if want := []int{1, 2, 3, 4}; !slices.Equal(un, want) {
		t.Fatalf("Union=%v want %v", un, want)
	}

	in := Intersection(a, b).Elements()
	slices.Sort(in)
	if want := []int{2, 3}; !slices.Equal(in, want) {
		t.Fatalf("Intersection=%v want %v", in, want)
	}

	diff := Difference(a, b).Elements()
	slices.Sort(diff)
	if want := []int{1}; !slices.Equal(diff, want) {
		t.Fatalf("Difference=%v want %v", diff, want)
	}

	if !SubsetOf(NewSet[int](), a) {
		t.Fatalf("empty set should be subset of any")
	}
	sub := NewSet[int]()
	sub.Add(2, 3)
	if !SubsetOf(sub, a) {
		t.Fatalf("{2,3} should be subset of {1,2,3}")
	}
	if SubsetOf(a, sub) {
		t.Fatalf("{1,2,3} should not be subset of {2,3}")
	}

	if !Equal(a, a.Clone()) {
		t.Fatalf("Clone should equal source")
	}
	if Equal(a, b) {
		t.Fatalf("a and b are not equal")
	}
}

func TestSet_OperationsNilSafe(t *testing.T) {
	t.Parallel()

	a := NewSet[int]()
	a.Add(1)
	if Union[int](nil, nil).Len() != 0 {
		t.Fatalf("Union(nil,nil) should be empty")
	}
	if Intersection(nil, a).Len() != 0 {
		t.Fatalf("Intersection(nil,a) should be empty")
	}
	if got := Difference(a, nil).Len(); got != 1 {
		t.Fatalf("Difference(a,nil) len=%d want 1", got)
	}
	if !SubsetOf(nil, a) {
		t.Fatalf("nil should be a subset of any set")
	}
	if !Equal[int](nil, nil) {
		t.Fatalf("Equal(nil,nil) should be true")
	}
	if Equal(nil, a) || Equal(a, nil) {
		t.Fatalf("nil and non-nil should not be equal")
	}
}

func TestSet_FunctionalHelpers(t *testing.T) {
	t.Parallel()

	s := NewSet[int]()
	s.Add(1, 2, 3, 4, 5)

	count := 0
	s.ForEach(func(int) { count++ })
	if count != 5 {
		t.Fatalf("ForEach count=%d want 5", count)
	}

	even := s.Filter(func(v int) bool { return v%2 == 0 }).Elements()
	slices.Sort(even)
	if want := []int{2, 4}; !slices.Equal(even, want) {
		t.Fatalf("Filter=%v want %v", even, want)
	}

	if !s.Any(func(v int) bool { return v == 5 }) {
		t.Fatalf("Any(==5) should be true")
	}
	if s.Any(func(v int) bool { return v > 100 }) {
		t.Fatalf("Any(>100) should be false")
	}
	if !s.All(func(v int) bool { return v > 0 }) {
		t.Fatalf("All(>0) should be true")
	}
	if s.All(func(v int) bool { return v%2 == 0 }) {
		t.Fatalf("All(even) should be false")
	}

	mapped := MapSet(s, func(v int) int { return v * 10 }).Elements()
	slices.Sort(mapped)
	if want := []int{10, 20, 30, 40, 50}; !slices.Equal(mapped, want) {
		t.Fatalf("MapSet=%v want %v", mapped, want)
	}

	sum := ReduceSet(s, 0, func(acc, v int) int { return acc + v })
	if sum != 15 {
		t.Fatalf("ReduceSet sum=%d want 15", sum)
	}
}

func TestSyncSet_Concurrent(t *testing.T) {
	t.Parallel()

	const goroutines = 16
	const perG = 200
	ss := NewSyncSet[int]()

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func(base int) {
			defer wg.Done()
			for i := 0; i < perG; i++ {
				ss.Add(base*perG + i)
			}
		}(g)
	}
	wg.Wait()
	if got, want := ss.Len(), goroutines*perG; got != want {
		t.Fatalf("len=%d want %d", got, want)
	}
	if !ss.Contains(0) {
		t.Fatalf("expected 0 to be present")
	}
	if got := len(ss.Elements()); got != goroutines*perG {
		t.Fatalf("Elements len=%d want %d", got, goroutines*perG)
	}
}

// ---------- Benchmarks ----------

func BenchmarkSet_Add(b *testing.B) {
	b.ReportAllocs()
	s := NewSet[int]()
	i := 0
	for b.Loop() {
		s.Add(i)
		i++
	}
}

func BenchmarkSet_Contains(b *testing.B) {
	const N = 10_000
	s := NewSet[int]()
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

func BenchmarkSet_Union(b *testing.B) {
	const N = 1024
	a := NewSet[int]()
	bb := NewSet[int]()
	for i := 0; i < N; i++ {
		a.Add(i)
		bb.Add(i + N/2)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = Union(a, bb)
	}
}
