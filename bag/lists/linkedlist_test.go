package lists

import (
	"slices"
	"sync"
	"testing"
)

func TestLinkedList_NewIsEmpty(t *testing.T) {
	t.Parallel()

	l := NewLinkedList[int]()
	if !l.IsEmpty() {
		t.Fatalf("expected empty list, got len=%d", l.Len())
	}
	if l.Len() != 0 {
		t.Fatalf("expected len 0, got %d", l.Len())
	}
	if _, ok := l.First(); ok {
		t.Fatal("expected First() ok=false on empty list")
	}
	if _, ok := l.Last(); ok {
		t.Fatal("expected Last() ok=false on empty list")
	}
	if _, ok := l.RemoveFirst(); ok {
		t.Fatal("expected RemoveFirst() ok=false on empty list")
	}
	if _, ok := l.RemoveLast(); ok {
		t.Fatal("expected RemoveLast() ok=false on empty list")
	}
}

func TestLinkedList_AddFirstAddLast(t *testing.T) {
	t.Parallel()

	l := NewLinkedList[int]()
	l.AddLast(2)
	l.AddLast(3)
	l.AddFirst(1)
	l.AddLast(4)

	if got, want := l.Len(), 4; got != want {
		t.Fatalf("len=%d, want %d", got, want)
	}
	if got, want := l.Elements(), []int{1, 2, 3, 4}; !slices.Equal(got, want) {
		t.Fatalf("elements=%v, want %v", got, want)
	}
}

func TestLinkedList_RemoveFirstLast(t *testing.T) {
	t.Parallel()

	l := NewLinkedList[string]()
	for _, v := range []string{"a", "b", "c"} {
		l.AddLast(v)
	}

	v, ok := l.RemoveFirst()
	if !ok || v != "a" {
		t.Fatalf("RemoveFirst()=%q,%v want a,true", v, ok)
	}
	v, ok = l.RemoveLast()
	if !ok || v != "c" {
		t.Fatalf("RemoveLast()=%q,%v want c,true", v, ok)
	}
	if got, want := l.Elements(), []string{"b"}; !slices.Equal(got, want) {
		t.Fatalf("elements=%v, want %v", got, want)
	}

	// Drain the last element and confirm both pointers are nil-equivalent.
	if v, ok := l.RemoveLast(); !ok || v != "b" {
		t.Fatalf("RemoveLast()=%q,%v want b,true", v, ok)
	}
	if !l.IsEmpty() {
		t.Fatalf("expected empty after draining, len=%d", l.Len())
	}
}

func TestLinkedList_Get(t *testing.T) {
	t.Parallel()

	l := NewLinkedList[int]()
	for i := 0; i < 10; i++ {
		l.AddLast(i)
	}

	tests := []struct {
		name  string
		index int
		want  int
		ok    bool
	}{
		{"first", 0, 0, true},
		{"middle low", 3, 3, true},
		{"middle high", 7, 7, true},
		{"last", 9, 9, true},
		{"negative", -1, 0, false},
		{"oob", 10, 0, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := l.Get(tc.index)
			if ok != tc.ok || (ok && got != tc.want) {
				t.Fatalf("Get(%d)=%d,%v want %d,%v", tc.index, got, ok, tc.want, tc.ok)
			}
		})
	}
}

func TestLinkedList_InsertAt(t *testing.T) {
	t.Parallel()

	l := NewLinkedList[int]()
	l.AddLast(1)
	l.AddLast(3)

	if !l.InsertAt(1, 2) {
		t.Fatal("InsertAt(1,2)=false, want true")
	}
	if got, want := l.Elements(), []int{1, 2, 3}; !slices.Equal(got, want) {
		t.Fatalf("after middle insert: %v want %v", got, want)
	}

	if !l.InsertAt(0, 0) {
		t.Fatal("InsertAt(0,0)=false, want true")
	}
	if !l.InsertAt(l.Len(), 4) {
		t.Fatal("InsertAt(end,4)=false, want true")
	}
	if got, want := l.Elements(), []int{0, 1, 2, 3, 4}; !slices.Equal(got, want) {
		t.Fatalf("after edge inserts: %v want %v", got, want)
	}

	if l.InsertAt(-1, 99) {
		t.Fatal("InsertAt(-1) should return false")
	}
	if l.InsertAt(l.Len()+1, 99) {
		t.Fatal("InsertAt past end should return false")
	}
}

func TestLinkedList_RemoveAt(t *testing.T) {
	t.Parallel()

	l := NewLinkedList[int]()
	for _, v := range []int{10, 20, 30, 40, 50} {
		l.AddLast(v)
	}

	v, ok := l.RemoveAt(2)
	if !ok || v != 30 {
		t.Fatalf("RemoveAt(2)=%d,%v want 30,true", v, ok)
	}
	if got, want := l.Elements(), []int{10, 20, 40, 50}; !slices.Equal(got, want) {
		t.Fatalf("after middle remove: %v want %v", got, want)
	}

	if v, ok := l.RemoveAt(0); !ok || v != 10 {
		t.Fatalf("RemoveAt(0)=%d,%v want 10,true", v, ok)
	}
	if v, ok := l.RemoveAt(l.Len() - 1); !ok || v != 50 {
		t.Fatalf("RemoveAt(last)=%d,%v want 50,true", v, ok)
	}

	if _, ok := l.RemoveAt(-1); ok {
		t.Fatal("RemoveAt(-1) should return false")
	}
	if _, ok := l.RemoveAt(l.Len()); ok {
		t.Fatal("RemoveAt(len) should return false")
	}
}

func TestLinkedList_ForEachAndClear(t *testing.T) {
	t.Parallel()

	l := NewLinkedList[int]()
	for i := 1; i <= 5; i++ {
		l.AddLast(i)
	}

	sum := 0
	l.ForEach(func(v int) { sum += v })
	if sum != 15 {
		t.Fatalf("sum=%d want 15", sum)
	}

	l.Clear()
	if !l.IsEmpty() || l.Len() != 0 {
		t.Fatalf("Clear() failed, len=%d", l.Len())
	}
}

func TestLinkedList_RemoveLast(t *testing.T) {
	t.Parallel()

	l := NewLinkedList[int]()
	if _, ok := l.RemoveLast(); ok {
		t.Fatal("RemoveLast on empty should return ok=false")
	}

	for _, v := range []int{1, 2, 3} {
		l.AddLast(v)
	}
	if v, ok := l.RemoveLast(); !ok || v != 3 {
		t.Fatalf("RemoveLast()=%d,%v want 3,true", v, ok)
	}
	if got, want := l.Elements(), []int{1, 2}; !slices.Equal(got, want) {
		t.Fatalf("after RemoveLast: %v want %v", got, want)
	}

	// Drain to exercise the size==1 branch.
	if v, ok := l.RemoveLast(); !ok || v != 2 {
		t.Fatalf("RemoveLast()=%d,%v want 2,true", v, ok)
	}
	if v, ok := l.RemoveLast(); !ok || v != 1 {
		t.Fatalf("RemoveLast()=%d,%v want 1,true", v, ok)
	}
	if !l.IsEmpty() {
		t.Fatalf("expected empty after draining, len=%d", l.Len())
	}
}

func TestSyncLinkedList_ConcurrentAddLast(t *testing.T) {
	t.Parallel()

	const goroutines = 16
	const perG = 200
	sl := NewSyncLinkedList[int]()

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < perG; i++ {
				sl.AddLast(i)
			}
		}()
	}
	wg.Wait()

	if got, want := sl.Len(), goroutines*perG; got != want {
		t.Fatalf("len=%d want %d", got, want)
	}
}

func TestSyncLinkedList_ConcurrentReads(t *testing.T) {
	t.Parallel()

	sl := NewSyncLinkedList[int]()
	for i := 0; i < 1000; i++ {
		sl.AddLast(i)
	}

	const readers = 8
	var wg sync.WaitGroup
	wg.Add(readers)
	for r := 0; r < readers; r++ {
		go func() {
			defer wg.Done()
			for i := 0; i < 500; i++ {
				_ = sl.Len()
				_ = sl.IsEmpty()
				_, _ = sl.First()
				_, _ = sl.Last()
				_, _ = sl.Get(i % 1000)
				_ = sl.Elements()
			}
		}()
	}
	wg.Wait()
}

func TestSyncLinkedList_ConcurrentMixed(t *testing.T) {
	t.Parallel()

	sl := NewSyncLinkedList[int]()
	for i := 0; i < 1000; i++ {
		sl.AddLast(i)
	}

	const goroutines = 8
	var wg sync.WaitGroup
	wg.Add(goroutines * 4)
	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				sl.AddLast(i)
			}
		}()
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				sl.AddFirst(i)
			}
		}()
		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				_, _ = sl.RemoveFirst()
			}
		}()
		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				_ = sl.Len()
				_, _ = sl.First()
				_, _ = sl.Last()
			}
		}()
	}
	wg.Wait()

	// 1000 initial + 8*200 added - 8*50 removed = 1400.
	if got, want := sl.Len(), 1000+goroutines*200-goroutines*50; got != want {
		t.Fatalf("len=%d want %d", got, want)
	}
}

// ---------- Benchmarks ----------

func BenchmarkLinkedList_AddLast(b *testing.B) {
	b.ReportAllocs()
	l := NewLinkedList[int]()
	for b.Loop() {
		l.AddLast(1)
	}
}

func BenchmarkLinkedList_AddFirst(b *testing.B) {
	b.ReportAllocs()
	l := NewLinkedList[int]()
	for b.Loop() {
		l.AddFirst(1)
	}
}

func BenchmarkLinkedList_GetMiddle(b *testing.B) {
	const N = 10_000
	l := NewLinkedList[int]()
	for i := 0; i < N; i++ {
		l.AddLast(i)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = l.Get(N / 2)
	}
}

func BenchmarkLinkedList_RemoveFirst(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		b.StopTimer()
		l := NewLinkedList[int]()
		for i := 0; i < 1024; i++ {
			l.AddLast(i)
		}
		b.StartTimer()
		for !l.IsEmpty() {
			_, _ = l.RemoveFirst()
		}
	}
}

func BenchmarkLinkedList_RemoveLast(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		b.StopTimer()
		l := NewLinkedList[int]()
		for i := 0; i < 1024; i++ {
			l.AddLast(i)
		}
		b.StartTimer()
		for !l.IsEmpty() {
			_, _ = l.RemoveLast()
		}
	}
}

// Insert+Remove pair at the middle: walk distance is ~N/2 from either end.
// Each op is paired so the list size stays at N across iterations.
func BenchmarkLinkedList_InsertRemoveMiddle(b *testing.B) {
	const N = 1024
	l := NewLinkedList[int]()
	for i := 0; i < N; i++ {
		l.AddLast(i)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		l.InsertAt(N/2, 0)
		_, _ = l.RemoveAt(N / 2)
	}
}

// Insert+Remove pair near the end: shows the bidirectional walk benefit
// (~1 hop from tail instead of ~N from head).
func BenchmarkLinkedList_InsertRemoveNearEnd(b *testing.B) {
	const N = 1024
	l := NewLinkedList[int]()
	for i := 0; i < N; i++ {
		l.AddLast(i)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		l.InsertAt(N-1, 0)
		_, _ = l.RemoveAt(N - 1)
	}
}

func BenchmarkLinkedList_RemoveFirst_One(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		l := NewLinkedList[int]()
		l.AddLast(1)
		_, _ = l.RemoveFirst()
	}
}
