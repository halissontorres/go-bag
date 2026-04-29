package heap

import (
	"math/rand"
	"sync"
	"testing"
)

func TestHeap_Int_MinHeapBehavior(t *testing.T) {
	h := New[int]()

	// 1. Test initial size
	if h.Len() != 0 {
		t.Errorf("expected Len 0, got %d", h.Len())
	}

	// 2. Insert elements out of order
	h.Push(50)
	h.Push(10)
	h.Push(30)
	h.Push(5)
	h.Push(100)

	// 3. Test if Peek returns the smallest without removing it
	minVal, err := h.Peek()
	if err != nil {
		t.Fatalf("unexpected error in Peek: %v", err)
	}
	if minVal != 5 {
		t.Errorf("expected Peek 5, got %d", minVal)
	}

	// 4. Test removal (Pop) verifying if ascending order is respected
	expectedOrder := []int{5, 10, 30, 50, 100}

	for i, expected := range expectedOrder {
		val, err := h.Pop()
		if err != nil {
			t.Fatalf("unexpected error in Pop at iteration %d: %v", i, err)
		}
		if val != expected {
			t.Errorf("expected Pop %d, got %d", expected, val)
		}
	}

	// 5. Verify if it is empty at the end
	if h.Len() != 0 {
		t.Errorf("expected Len 0 after removing everything, got %d", h.Len())
	}
}

func TestHeap_String_Ordering(t *testing.T) {
	h := New[string]()

	h.Push("Zebra")
	h.Push("Abacaxi")
	h.Push("Maçã")
	h.Push("Banana")

	// The smallest element (alphabetical order) should be "Abacaxi"
	val, err := h.Pop()
	if err != nil {
		t.Fatalf("unexpected error in Pop: %v", err)
	}
	if val != "Abacaxi" {
		t.Errorf("expected 'Abacaxi', got '%s'", val)
	}

	// The next one should be "Banana"
	val, _ = h.Pop()
	if val != "Banana" {
		t.Errorf("expected 'Banana', got '%s'", val)
	}
}

func TestHeap_EmptyErrors(t *testing.T) {
	h := New[float64]()

	// Test Peek on empty Heap
	_, err := h.Peek()
	if err == nil {
		t.Error("expected error when Peek on empty heap, but none occurred")
	}

	// Test Pop on empty Heap
	_, err = h.Pop()
	if err == nil {
		t.Error("expected error when Pop on empty heap, but none occurred")
	}
}

func TestSyncHeap_MinHeapBehavior(t *testing.T) {
	h := NewSync[int]()

	if !h.IsEmpty() {
		t.Fatal("expected empty heap")
	}

	h.Push(50)
	h.Push(10)
	h.Push(5)
	h.Push(30)

	if h.Len() != 4 {
		t.Errorf("expected Len 4, got %d", h.Len())
	}

	min, err := h.Peek()
	if err != nil {
		t.Fatalf("unexpected error in Peek: %v", err)
	}
	if min != 5 {
		t.Errorf("expected Peek 5, got %d", min)
	}
	if h.Len() != 4 {
		t.Errorf("Peek must not remove elements; expected Len 4, got %d", h.Len())
	}

	expected := []int{5, 10, 30, 50}
	for i, want := range expected {
		got, err := h.Pop()
		if err != nil {
			t.Fatalf("unexpected error in Pop at %d: %v", i, err)
		}
		if got != want {
			t.Errorf("Pop[%d]: expected %d, got %d", i, want, got)
		}
	}

	if !h.IsEmpty() {
		t.Error("expected empty heap after draining")
	}
}

func TestSyncHeap_EmptyErrors(t *testing.T) {
	h := NewSync[float64]()

	_, err := h.Peek()
	if err == nil {
		t.Error("expected error on Peek of empty SyncHeap")
	}

	_, err = h.Pop()
	if err == nil {
		t.Error("expected error on Pop of empty SyncHeap")
	}
}

func TestSyncHeap_ConcurrentPushPop(t *testing.T) {
	const goroutines = 10
	const itemsEach = 100

	h := NewSync[int]()
	var wg sync.WaitGroup

	// Concurrent pushes
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func(base int) {
			defer wg.Done()
			for i := 0; i < itemsEach; i++ {
				h.Push(base*itemsEach + i)
			}
		}(g)
	}
	wg.Wait()

	if h.Len() != goroutines*itemsEach {
		t.Errorf("expected %d elements, got %d", goroutines*itemsEach, h.Len())
	}

	// Concurrent pops
	results := make(chan int, goroutines*itemsEach)
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < itemsEach; i++ {
				v, err := h.Pop()
				if err != nil {
					t.Errorf("unexpected error in concurrent Pop: %v", err)
					return
				}
				results <- v
			}
		}()
	}
	wg.Wait()
	close(results)

	if h.Len() != 0 {
		t.Errorf("expected empty heap after concurrent pops, got Len %d", h.Len())
	}

	total := 0
	for range results {
		total++
	}
	if total != goroutines*itemsEach {
		t.Errorf("expected %d popped values, got %d", goroutines*itemsEach, total)
	}
}

func TestSyncHeap_ConcurrentPeek(t *testing.T) {
	h := NewSync[int]()
	for i := 100; i >= 1; i-- {
		h.Push(i)
	}

	var wg sync.WaitGroup
	wg.Add(20)
	for i := 0; i < 20; i++ {
		go func() {
			defer wg.Done()
			v, err := h.Peek()
			if err != nil {
				t.Errorf("unexpected error in concurrent Peek: %v", err)
				return
			}
			if v != 1 {
				t.Errorf("expected Peek 1, got %d", v)
			}
		}()
	}
	wg.Wait()
}

// BenchmarkPush_Sequential tests the worst case/best case depending on Min/Max heap,
// since numbers are already inserted in order.
func BenchmarkHeap_PushSequential(b *testing.B) {
	h := New[int]()
	b.ResetTimer() // Reset timer to ignore the initialization above

	for i := 0; i < b.N; i++ {
		h.Push(i)
	}
}

// BenchmarkPush_Random tests the most realistic scenario, inserting completely
// random values, forcing up-heap at various tree levels.
func BenchmarkHeap_PushRandom(b *testing.B) {
	// Pre-generate numbers to ignore rand time in the benchmark
	nums := make([]int, b.N)
	for i := 0; i < b.N; i++ {
		nums[i] = rand.Intn(1000000)
	}

	h := New[int]()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		h.Push(nums[i])
	}
}

// BenchmarkPop measures removal time. As Pop requires the Heap
// to have elements, we fill it before starting the timer.
func BenchmarkHeap_Pop(b *testing.B) {
	h := New[int]()

	// Setup: fill the Heap with b.N elements
	for i := 0; i < b.N; i++ {
		h.Push(rand.Intn(1000000))
	}

	b.ResetTimer() // Start measuring real time only for Pop

	for i := 0; i < b.N; i++ {
		_, _ = h.Pop()
	}
}
