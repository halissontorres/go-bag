package queue

import (
	"slices"
	"sync"
	"testing"
)

func TestDeque_NewIsEmpty(t *testing.T) {
	t.Parallel()

	d := NewDeque[int]()
	if !d.IsEmpty() {
		t.Fatalf("expected empty deque")
	}
	if _, ok := d.PeekFront(); ok {
		t.Fatal("PeekFront on empty should be false")
	}
	if _, ok := d.PeekBack(); ok {
		t.Fatal("PeekBack on empty should be false")
	}
	if _, ok := d.PopFront(); ok {
		t.Fatal("PopFront on empty should be false")
	}
	if _, ok := d.PopBack(); ok {
		t.Fatal("PopBack on empty should be false")
	}
}

func TestDeque_PushPopOrder(t *testing.T) {
	t.Parallel()

	d := NewDeque[int]()
	d.PushBack(2)
	d.PushBack(3)
	d.PushFront(1)
	d.PushFront(0)
	d.PushBack(4)

	if got, want := d.Elements(), []int{0, 1, 2, 3, 4}; !slices.Equal(got, want) {
		t.Fatalf("Elements=%v want %v", got, want)
	}
	if v, _ := d.PeekFront(); v != 0 {
		t.Fatalf("PeekFront=%d want 0", v)
	}
	if v, _ := d.PeekBack(); v != 4 {
		t.Fatalf("PeekBack=%d want 4", v)
	}

	if v, ok := d.PopFront(); !ok || v != 0 {
		t.Fatalf("PopFront=%d,%v want 0,true", v, ok)
	}
	if v, ok := d.PopBack(); !ok || v != 4 {
		t.Fatalf("PopBack=%d,%v want 4,true", v, ok)
	}
	if got, want := d.Elements(), []int{1, 2, 3}; !slices.Equal(got, want) {
		t.Fatalf("Elements after pops=%v want %v", got, want)
	}
}

func TestDeque_GrowResize(t *testing.T) {
	t.Parallel()

	// Force several grow cycles.
	d := NewDequeWithCap[int](2)
	const N = 1000
	for i := 0; i < N; i++ {
		d.PushBack(i)
	}
	if got, want := d.Len(), N; got != want {
		t.Fatalf("len=%d want %d", got, want)
	}
	for i := 0; i < N; i++ {
		v, ok := d.PopFront()
		if !ok || v != i {
			t.Fatalf("PopFront=%d,%v want %d,true", v, ok, i)
		}
	}
	if !d.IsEmpty() {
		t.Fatalf("expected empty after draining")
	}
}

func TestDeque_WrapAroundElements(t *testing.T) {
	t.Parallel()

	d := NewDequeWithCap[int](4)
	d.PushBack(1)
	d.PushBack(2)
	d.PushBack(3)
	if v, _ := d.PopFront(); v != 1 {
		t.Fatalf("expected 1 popped from front")
	}
	d.PushBack(4)
	d.PushBack(5) // forces wrap of tail in the small ring
	if got, want := d.Elements(), []int{2, 3, 4, 5}; !slices.Equal(got, want) {
		t.Fatalf("wrap elements=%v want %v", got, want)
	}
}

func TestDeque_NewDequeWithCapClampsLow(t *testing.T) {
	t.Parallel()

	d := NewDequeWithCap[int](-5)
	d.PushBack(42)
	if v, _ := d.PeekFront(); v != 42 {
		t.Fatalf("expected 42 after push on clamped-cap deque, got %d", v)
	}
}

func TestDeque_Clear(t *testing.T) {
	t.Parallel()

	d := NewDeque[int]()
	for i := 0; i < 5; i++ {
		d.PushBack(i)
	}
	d.Clear()
	if !d.IsEmpty() {
		t.Fatalf("expected empty after Clear")
	}
	if d.Elements() != nil {
		t.Fatalf("Elements after Clear should be nil")
	}
}

func TestDeque_String(t *testing.T) {
	t.Parallel()

	d := NewDeque[int]()
	d.PushBack(1)
	d.PushBack(2)
	if got, want := d.String(), "[1 2]"; got != want {
		t.Fatalf("String=%q want %q", got, want)
	}
}

func TestSyncDeque_Concurrent(t *testing.T) {
	t.Parallel()

	const goroutines = 8
	const perG = 250
	sd := NewSyncDeque[int]()

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < perG; i++ {
				if i%2 == 0 {
					sd.PushBack(i)
				} else {
					sd.PushFront(i)
				}
			}
		}()
	}
	wg.Wait()
	if got, want := sd.Len(), goroutines*perG; got != want {
		t.Fatalf("len=%d want %d", got, want)
	}
}

// ---------- Benchmarks ----------

func BenchmarkDeque_PushBack(b *testing.B) {
	b.ReportAllocs()
	d := NewDeque[int]()
	for b.Loop() {
		d.PushBack(1)
	}
}

func BenchmarkDeque_PushFront(b *testing.B) {
	b.ReportAllocs()
	d := NewDeque[int]()
	for b.Loop() {
		d.PushFront(1)
	}
}

func BenchmarkDeque_PingPong(b *testing.B) {
	b.ReportAllocs()
	d := NewDeque[int]()
	for b.Loop() {
		d.PushBack(1)
		_, _ = d.PopFront()
	}
}
