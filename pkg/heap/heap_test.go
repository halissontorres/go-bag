package heap

import (
	"math/rand"
	"sync"
	"testing"
)

func TestHeap_Int_MinHeapBehavior(t *testing.T) {
	h := New[int]() // padrão: Min-Heap

	if h.Len() != 0 {
		t.Errorf("expected Len 0, got %d", h.Len())
	}

	h.Push(50)
	h.Push(10)
	h.Push(30)
	h.Push(5)
	h.Push(100)

	minVal, err := h.Peek()
	if err != nil {
		t.Fatalf("unexpected error in Peek: %v", err)
	}
	if minVal != 5 {
		t.Errorf("expected Peek 5, got %d", minVal)
	}

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

func TestHeap_Int_MaxHeapBehavior(t *testing.T) {
	h := New[int](WithMaxHeap[int]())

	h.Push(50)
	h.Push(10)
	h.Push(30)
	h.Push(5)
	h.Push(100)

	maxVal, err := h.Peek()
	if err != nil {
		t.Fatalf("unexpected error in Peek: %v", err)
	}
	if maxVal != 100 {
		t.Errorf("expected Peek 100, got %d", maxVal)
	}

	expectedOrder := []int{100, 50, 30, 10, 5}
	for i, expected := range expectedOrder {
		val, err := h.Pop()
		if err != nil {
			t.Fatalf("unexpected error in Pop at iteration %d: %v", i, err)
		}
		if val != expected {
			t.Errorf("expected Pop %d, got %d", expected, val)
		}
	}
}

func TestHeap_WithLessFunc(t *testing.T) {
	// ordena pelo último dígito
	h := New[int](WithLessFunc(func(a, b int) bool {
		return a%10 < b%10
	}))

	h.Push(21) // último dígito: 1
	h.Push(35) // último dígito: 5
	h.Push(43) // último dígito: 3
	h.Push(59) // último dígito: 9

	first, err := h.Peek()
	if err != nil {
		t.Fatalf("unexpected error in Peek: %v", err)
	}
	if first%10 != 1 {
		t.Errorf("expected element with last digit 1, got %d", first)
	}

	prev, _ := h.Pop()
	for h.Len() > 0 {
		curr, _ := h.Pop()
		if prev%10 > curr%10 {
			t.Errorf("order violation: %d came before %d", prev, curr)
		}
		prev = curr
	}
}

func TestHeap_WithMinHeap_Explicit(t *testing.T) {
	// WithMinHeap explícito deve se comportar igual ao padrão
	h := New[int](WithMinHeap[int]())

	h.Push(20)
	h.Push(3)
	h.Push(15)

	val, err := h.Pop()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 3 {
		t.Errorf("expected 3, got %d", val)
	}
}

// --- Testes existentes sem alteração ---

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

	minm, err := h.Peek()
	if err != nil {
		t.Fatalf("unexpected error in Peek: %v", err)
	}
	if minm != 5 {
		t.Errorf("expected Peek 5, got %d", minm)
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

func TestSyncHeap_MaxHeapBehavior(t *testing.T) {
	h := NewSync[int](WithMaxHeap[int]())

	h.Push(50)
	h.Push(10)
	h.Push(5)
	h.Push(30)

	maxx, err := h.Peek()
	if err != nil {
		t.Fatalf("unexpected error in Peek: %v", err)
	}
	if maxx != 50 {
		t.Errorf("expected Peek 50, got %d", maxx)
	}

	expected := []int{50, 30, 10, 5}
	for i, want := range expected {
		got, err := h.Pop()
		if err != nil {
			t.Fatalf("unexpected error in Pop at %d: %v", i, err)
		}
		if got != want {
			t.Errorf("Pop[%d]: expected %d, got %d", i, want, got)
		}
	}
}

// Benchmark comparativo Min vs Max para evidenciar que o custo é idêntico
func BenchmarkHeap_MinVsMax(b *testing.B) {
	b.Run("MinHeap", func(b *testing.B) {
		h := New[int](WithMinHeap[int]())
		for i := 0; i < b.N; i++ {
			h.Push(rand.Intn(1000000))
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = h.Pop()
		}
	})

	b.Run("MaxHeap", func(b *testing.B) {
		h := New[int](WithMaxHeap[int]())
		for i := 0; i < b.N; i++ {
			h.Push(rand.Intn(1000000))
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = h.Pop()
		}
	})
}
