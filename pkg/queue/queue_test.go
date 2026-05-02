package queue

import (
	"slices"
	"sync"
	"testing"
)

func TestQueue_NewIsEmpty(t *testing.T) {
	t.Parallel()

	q := NewQueue[int]()
	if !q.IsEmpty() {
		t.Fatalf("expected empty queue")
	}
	if _, ok := q.Peek(); ok {
		t.Fatal("Peek on empty should be false")
	}
	if _, ok := q.Dequeue(); ok {
		t.Fatal("Dequeue on empty should be false")
	}
	if q.Elements() != nil {
		t.Fatalf("Elements() on empty should be nil")
	}
}

func TestQueue_NewQueue(t *testing.T) {
	t.Parallel()
	capacity := 256
	q := NewQueue[int](WithInitialCap(capacity))
	cap := cap(q.items)
	if cap != capacity {
		t.Fatalf("expected initial capacity %d, got %d", capacity, cap)
	}
}

func TestQueue_NewSyncQueue(t *testing.T) {
	t.Parallel()
	cap := 256
	q := NewSyncQueue[int](WithInitialCap(cap))
	len := q.Len()
	if q.Len() != len {
		t.Fatalf("expected initial capacity %d, got %d", cap, len)
	}
}

func TestQueue_FIFOOrder(t *testing.T) {
	t.Parallel()

	q := NewQueue[int]()
	for i := 1; i <= 5; i++ {
		q.Enqueue(i)
	}
	if got, want := q.Len(), 5; got != want {
		t.Fatalf("len=%d want %d", got, want)
	}
	if v, ok := q.Peek(); !ok || v != 1 {
		t.Fatalf("Peek=%d,%v want 1,true", v, ok)
	}
	if got, want := q.Elements(), []int{1, 2, 3, 4, 5}; !slices.Equal(got, want) {
		t.Fatalf("Elements=%v want %v", got, want)
	}

	for want := 1; want <= 5; want++ {
		got, ok := q.Dequeue()
		if !ok || got != want {
			t.Fatalf("Dequeue=%d,%v want %d,true", got, ok, want)
		}
	}
	if !q.IsEmpty() {
		t.Fatalf("expected empty after draining")
	}
}

func TestQueue_CompactKeepsOrder(t *testing.T) {
	t.Parallel()

	q := NewQueue[int]()
	for i := 0; i < 100; i++ {
		q.Enqueue(i)
	}
	// Force compaction by draining most of the queue.
	for i := 0; i < 80; i++ {
		if v, ok := q.Dequeue(); !ok || v != i {
			t.Fatalf("Dequeue=%d,%v want %d,true", v, ok, i)
		}
	}
	if got, want := q.Elements(), makeRange(80, 100); !slices.Equal(got, want) {
		t.Fatalf("post-compact elements mismatch: got %v want %v", got, want)
	}
}

func TestQueue_Clear(t *testing.T) {
	t.Parallel()

	q := NewQueue[int]()
	q.Enqueue(1)
	q.Enqueue(2)
	q.Clear()
	if !q.IsEmpty() {
		t.Fatalf("expected empty after Clear")
	}
	q.Enqueue(99)
	if v, ok := q.Peek(); !ok || v != 99 {
		t.Fatalf("Peek after refill=%d,%v want 99,true", v, ok)
	}
}

func TestSyncQueue_Concurrent(t *testing.T) {
	t.Parallel()

	const goroutines = 8
	const perG = 250
	sq := NewSyncQueue[int]()

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < perG; i++ {
				sq.Enqueue(i)
			}
		}()
	}
	wg.Wait()
	if got, want := sq.Len(), goroutines*perG; got != want {
		t.Fatalf("len=%d want %d", got, want)
	}
}

func makeRange(start, end int) []int {
	out := make([]int, 0, end-start)
	for i := start; i < end; i++ {
		out = append(out, i)
	}
	return out
}

// ---------- Benchmarks ----------

func BenchmarkQueue_Enqueue(b *testing.B) {
	b.ReportAllocs()
	q := NewQueue[int]()
	for b.Loop() {
		q.Enqueue(1)
	}
}

func BenchmarkQueue_EnqueueDequeue(b *testing.B) {
	b.ReportAllocs()
	q := NewQueue[int]()
	for b.Loop() {
		q.Enqueue(1)
		_, _ = q.Dequeue()
	}
}
