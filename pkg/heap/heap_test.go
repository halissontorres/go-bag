package heap

import (
	"math/rand"
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
