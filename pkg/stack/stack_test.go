package stack

import (
	"slices"
	"sync"
	"testing"
)

func TestStack_NewIsEmpty(t *testing.T) {
	t.Parallel()

	s := NewStack[int]()
	if !s.IsEmpty() || s.Len() != 0 {
		t.Fatalf("expected empty stack, len=%d", s.Len())
	}
	if _, ok := s.Peek(); ok {
		t.Fatal("Peek on empty should be false")
	}
	if _, ok := s.Pop(); ok {
		t.Fatal("Pop on empty should be false")
	}
	if s.Elements() != nil {
		t.Fatalf("Elements on empty should be nil")
	}
}

func TestStack_LIFOOrder(t *testing.T) {
	t.Parallel()

	s := NewStack[int]()
	for i := 1; i <= 5; i++ {
		s.Push(i)
	}
	if got, want := s.Len(), 5; got != want {
		t.Fatalf("len=%d want %d", got, want)
	}
	if v, ok := s.Peek(); !ok || v != 5 {
		t.Fatalf("Peek=%d,%v want 5,true", v, ok)
	}
	if got, want := s.Elements(), []int{1, 2, 3, 4, 5}; !slices.Equal(got, want) {
		t.Fatalf("Elements=%v want %v", got, want)
	}
	for want := 5; want >= 1; want-- {
		got, ok := s.Pop()
		if !ok || got != want {
			t.Fatalf("Pop=%d,%v want %d,true", got, ok, want)
		}
	}
	if !s.IsEmpty() {
		t.Fatalf("expected empty after draining")
	}
}

func TestStack_Clear(t *testing.T) {
	t.Parallel()

	s := NewStack[string]()
	s.Push("a")
	s.Push("b")
	s.Clear()
	if !s.IsEmpty() {
		t.Fatalf("Clear failed")
	}
	s.Push("c")
	if v, _ := s.Peek(); v != "c" {
		t.Fatalf("expected 'c' after refill, got %q", v)
	}
}

func TestSyncStack_Concurrent(t *testing.T) {
	t.Parallel()

	const goroutines = 16
	const perG = 200
	ss := NewSyncStack[int]()

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < perG; i++ {
				ss.Push(i)
			}
		}()
	}
	wg.Wait()
	if got, want := ss.Len(), goroutines*perG; got != want {
		t.Fatalf("len=%d want %d", got, want)
	}

	// Drain concurrently and confirm we recover exactly that many items.
	var popped sync.WaitGroup
	popped.Add(goroutines)
	count := make(chan int, goroutines)
	for g := 0; g < goroutines; g++ {
		go func() {
			defer popped.Done()
			n := 0
			for {
				_, ok := ss.Pop()
				if !ok {
					break
				}
				n++
			}
			count <- n
		}()
	}
	popped.Wait()
	close(count)
	total := 0
	for n := range count {
		total += n
	}
	if total != goroutines*perG {
		t.Fatalf("popped=%d want %d", total, goroutines*perG)
	}
	if !ss.IsEmpty() {
		t.Fatalf("expected drained stack to be empty")
	}
}

// ---------- Benchmarks ----------

func BenchmarkStack_Push(b *testing.B) {
	b.ReportAllocs()
	s := NewStack[int]()
	for b.Loop() {
		s.Push(1)
	}
}

func BenchmarkStack_PushPop(b *testing.B) {
	b.ReportAllocs()
	s := NewStack[int]()
	for b.Loop() {
		s.Push(1)
		_, _ = s.Pop()
	}
}
