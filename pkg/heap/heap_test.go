package heap

import (
	"math/rand"
	"sync"
	"testing"

	"github.com/halissontorres/go-bag/pkg/comparator"
)

func TestHeap_Int_MinHeapBehavior(t *testing.T) {
	h := New(comparator.Natural[int]())

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
	h := New(comparator.Natural[string]())

	h.Push("Zebra")
	h.Push("Abacaxi")
	h.Push("Maçã")
	h.Push("Banana")

	val, err := h.Pop()
	if err != nil {
		t.Fatalf("unexpected error in Pop: %v", err)
	}
	if val != "Abacaxi" {
		t.Errorf("expected 'Abacaxi', got '%s'", val)
	}

	val, _ = h.Pop()
	if val != "Banana" {
		t.Errorf("expected 'Banana', got '%s'", val)
	}
}

func TestHeap_EmptyErrors(t *testing.T) {
	h := New(comparator.Natural[float64]())

	_, err := h.Peek()
	if err == nil {
		t.Error("expected error when Peek on empty heap, but none occurred")
	}

	_, err = h.Pop()
	if err == nil {
		t.Error("expected error when Pop on empty heap, but none occurred")
	}
}

func TestSyncHeap_EmptyErrors(t *testing.T) {
	h := NewSync(comparator.Natural[float64]())

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

	h := NewSync(comparator.Natural[int]())
	var wg sync.WaitGroup

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
	h := NewSync(comparator.Natural[int]())
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

func TestHeap_MaxHeapBehavior(t *testing.T) {
	h := New(comparator.Reverse[int]())

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

func TestHeap_CustomComparator(t *testing.T) {
	h := New(comparator.Comparator[int](func(a, b int) bool {
		return a%10 < b%10
	}))

	h.Push(21) // last digit: 1
	h.Push(35) // last digit: 5
	h.Push(43) // last digit: 3
	h.Push(59) // last digit: 9

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

type task struct {
	name     string
	priority int
}

func TestHeap_StructByField(t *testing.T) {
	h := New(comparator.ByField(func(t task) int { return t.priority }))

	h.Push(task{name: "low", priority: 50})
	h.Push(task{name: "critical", priority: 1})
	h.Push(task{name: "medium", priority: 30})

	first, err := h.Pop()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if first.priority != 1 || first.name != "critical" {
		t.Errorf("expected {critical 1}, got {%s %d}", first.name, first.priority)
	}

	second, _ := h.Pop()
	if second.priority != 30 {
		t.Errorf("expected priority 30, got %d", second.priority)
	}

	third, _ := h.Pop()
	if third.priority != 50 {
		t.Errorf("expected priority 50, got %d", third.priority)
	}
}

func TestHeap_StructMultiCriterion(t *testing.T) {
	byPriority := comparator.ByField(func(t task) int { return t.priority })
	byName := comparator.ByField(func(t task) string { return t.name })
	cmp := byPriority.Then(byName)

	h := New(cmp)

	h.Push(task{name: "B", priority: 1})
	h.Push(task{name: "A", priority: 1})
	h.Push(task{name: "C", priority: 0})

	first, _ := h.Pop()
	if first.priority != 0 || first.name != "C" {
		t.Errorf("expected {C 0}, got {%s %d}", first.name, first.priority)
	}

	second, _ := h.Pop()
	if second.priority != 1 || second.name != "A" {
		t.Errorf("expected {A 1}, got {%s %d}", second.name, second.priority)
	}

	third, _ := h.Pop()
	if third.priority != 1 || third.name != "B" {
		t.Errorf("expected {B 1}, got {%s %d}", third.name, third.priority)
	}
}

func TestSyncHeap_Reverse(t *testing.T) {
	h := NewSync(comparator.Reverse[int]())

	h.Push(10)
	h.Push(50)
	h.Push(30)

	first, err := h.Pop()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if first != 50 {
		t.Errorf("expected 50, got %d", first)
	}

	second, _ := h.Pop()
	if second != 30 {
		t.Errorf("expected 30, got %d", second)
	}

	third, _ := h.Pop()
	if third != 10 {
		t.Errorf("expected 10, got %d", third)
	}
}

func TestSyncHeap_MinHeapBehavior(t *testing.T) {
	h := NewSync(comparator.Natural[int]())

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
	h := NewSync(comparator.Reverse[int]())

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

func BenchmarkHeap_PushSequential(b *testing.B) {
	h := New(comparator.Natural[int]())
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		h.Push(i)
	}
}

func BenchmarkHeap_PushRandom(b *testing.B) {
	nums := make([]int, b.N)
	for i := 0; i < b.N; i++ {
		nums[i] = rand.Intn(1000000)
	}

	h := New(comparator.Natural[int]())
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		h.Push(nums[i])
	}
}

func BenchmarkHeap_Pop(b *testing.B) {
	h := New(comparator.Natural[int]())

	for i := 0; i < b.N; i++ {
		h.Push(rand.Intn(1000000))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = h.Pop()
	}
}

func BenchmarkHeap_MinVsMax(b *testing.B) {
	b.Run("MinHeap", func(b *testing.B) {
		h := New(comparator.Natural[int]())
		for i := 0; i < b.N; i++ {
			h.Push(rand.Intn(1000000))
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = h.Pop()
		}
	})

	b.Run("MaxHeap", func(b *testing.B) {
		h := New(comparator.Reverse[int]())
		for i := 0; i < b.N; i++ {
			h.Push(rand.Intn(1000000))
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = h.Pop()
		}
	})
}
