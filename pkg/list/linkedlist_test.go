package list

import (
	"math/rand"
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

func TestLinkedList_NewFromSlice(t *testing.T) {
	t.Parallel()

	slice := []int{1, 2, 3, 4, 5}
	l := NewLinkedListFromSlice(slice)

	if got, want := l.Len(), 5; got != want {
		t.Fatalf("len=%d want %d", got, want)
	}
	if got, want := l.Elements(), slice; !slices.Equal(got, want) {
		t.Fatalf("elements=%v want %v", got, want)
	}
}

func TestLinkedList_Aliases(t *testing.T) {
	t.Parallel()

	l := NewLinkedList[int]()
	l.PushBack(1)
	l.PushFront(0)
	l.Append(2)

	if got, want := l.Elements(), []int{0, 1, 2}; !slices.Equal(got, want) {
		t.Fatalf("after pushes: %v want %v", got, want)
	}

	v, ok := l.PopFront()
	if !ok || v != 0 {
		t.Fatalf("PopFront()=%d,%v want 0,true", v, ok)
	}
	v, ok = l.PopBack()
	if !ok || v != 2 {
		t.Fatalf("PopBack()=%d,%v want 2,true", v, ok)
	}
}

func TestLinkedList_String(t *testing.T) {
	t.Parallel()

	l := NewLinkedList[int]()
	if got, want := l.String(), "[]"; got != want {
		t.Fatalf("empty String()=%q want %q", got, want)
	}

	l.AddLast(1)
	l.AddLast(2)
	l.AddLast(3)
	if got, want := l.String(), "[1, 2, 3]"; got != want {
		t.Fatalf("String()=%q want %q", got, want)
	}
}

func TestLinkedList_Stream(t *testing.T) {
	t.Parallel()

	l := NewLinkedList[int]()
	for i := 1; i <= 5; i++ {
		l.AddLast(i)
	}

	s := l.Stream()
	got := s.ToSlice()
	want := []int{1, 2, 3, 4, 5}
	if !slices.Equal(got, want) {
		t.Fatalf("Stream().ToSlice()=%v want %v", got, want)
	}
}

func TestLinkedList_IteratorExtras(t *testing.T) {
	t.Parallel()

	l := NewLinkedList[int]()
	for i := 1; i <= 3; i++ {
		l.AddLast(i)
	}

	t.Run("ReverseIter", func(t *testing.T) {
		it := l.ReverseIter()
		var got []int
		for {
			v, ok := it.Next()
			if !ok {
				break
			}
			got = append(got, v)
		}
		want := []int{3, 2, 1}
		if !slices.Equal(got, want) {
			t.Fatalf("ReverseIter sequence=%v want %v", got, want)
		}
	})

	t.Run("Prev forward", func(t *testing.T) {
		it := l.Iter()
		it.Next() // returns 1, moves to 2
		it.Next() // returns 2, moves to 3
		v, ok := it.Prev()
		if !ok || v != 2 {
			t.Fatalf("Prev() after 2 Next()=%d,%v want 2,true", v, ok)
		}
		// now curr is 2, Next should return 2 and move to 3
		v, ok = it.Next()
		if !ok || v != 2 {
			t.Fatalf("Next() after Prev()=%d,%v want 2,true", v, ok)
		}
	})

	t.Run("Reset", func(t *testing.T) {
		it := l.Iter()
		it.Next()
		it.Next()
		it.Reset()
		v, ok := it.Next()
		if !ok || v != 1 {
			t.Fatalf("Next() after Reset()=%d,%v want 1,true", v, ok)
		}
	})

	t.Run("Edge cases", func(t *testing.T) {
		it := l.Iter()
		if _, ok := it.Prev(); ok {
			t.Fatal("Prev() at start of Iter should be false")
		}
		it.Next() // 1
		it.Next() // 2
		it.Next() // 3
		it.Next() // nil
		v, ok := it.Prev()
		if !ok || v != 3 {
			t.Fatalf("Prev() at end of Iter should be 3, got %d, %v", v, ok)
		}

		it = l.ReverseIter()
		if _, ok := it.Prev(); ok {
			t.Fatal("Prev() at start of ReverseIter should be false")
		}
		it.Next() // 3
		it.Next() // 2
		it.Next() // 1
		it.Next() // nil
		v, ok = it.Prev()
		if !ok || v != 1 {
			t.Fatalf("Prev() at end of ReverseIter should be 1, got %d, %v", v, ok)
		}
	})
}

func TestSyncLinkedList_Aliases(t *testing.T) {
	t.Parallel()

	sl := NewSyncLinkedList[int]()
	sl.PushBack(1)
	sl.PushFront(0)
	sl.Append(2)

	if got, want := sl.Elements(), []int{0, 1, 2}; !slices.Equal(got, want) {
		t.Fatalf("after pushes: %v want %v", got, want)
	}

	v, ok := sl.PopFront()
	if !ok || v != 0 {
		t.Fatalf("PopFront()=%d,%v want 0,true", v, ok)
	}
	v, ok = sl.PopBack()
	if !ok || v != 2 {
		t.Fatalf("PopBack()=%d,%v want 2,true", v, ok)
	}
	if got, want := sl.Elements(), []int{1}; !slices.Equal(got, want) {
		t.Fatalf("after pops: %v want %v", got, want)
	}
}

func TestSyncLinkedList_RemoveLast(t *testing.T) {
	t.Parallel()

	sl := NewSyncLinkedList[int]()
	if _, ok := sl.RemoveLast(); ok {
		t.Fatal("RemoveLast on empty should return ok=false")
	}

	sl.AddLast(10)
	sl.AddLast(20)
	sl.AddLast(30)

	v, ok := sl.RemoveLast()
	if !ok || v != 30 {
		t.Fatalf("RemoveLast()=%d,%v want 30,true", v, ok)
	}
	if got, want := sl.Elements(), []int{10, 20}; !slices.Equal(got, want) {
		t.Fatalf("after RemoveLast: %v want %v", got, want)
	}
}

func TestSyncLinkedList_InsertAt(t *testing.T) {
	t.Parallel()

	sl := NewSyncLinkedList[int]()
	sl.AddLast(1)
	sl.AddLast(3)

	if !sl.InsertAt(1, 2) {
		t.Fatal("InsertAt(1,2) should return true")
	}
	if got, want := sl.Elements(), []int{1, 2, 3}; !slices.Equal(got, want) {
		t.Fatalf("after middle insert: %v want %v", got, want)
	}

	if !sl.InsertAt(0, 0) {
		t.Fatal("InsertAt(0,0) should return true")
	}
	if !sl.InsertAt(sl.Len(), 4) {
		t.Fatal("InsertAt(end,4) should return true")
	}
	if got, want := sl.Elements(), []int{0, 1, 2, 3, 4}; !slices.Equal(got, want) {
		t.Fatalf("after edge inserts: %v want %v", got, want)
	}

	if sl.InsertAt(-1, 99) {
		t.Fatal("InsertAt(-1) should return false")
	}
	if sl.InsertAt(sl.Len()+1, 99) {
		t.Fatal("InsertAt past end should return false")
	}
}

func TestSyncLinkedList_RemoveAt(t *testing.T) {
	t.Parallel()

	sl := NewSyncLinkedList[int]()
	for _, v := range []int{10, 20, 30, 40, 50} {
		sl.AddLast(v)
	}

	v, ok := sl.RemoveAt(2)
	if !ok || v != 30 {
		t.Fatalf("RemoveAt(2)=%d,%v want 30,true", v, ok)
	}
	if got, want := sl.Elements(), []int{10, 20, 40, 50}; !slices.Equal(got, want) {
		t.Fatalf("after middle remove: %v want %v", got, want)
	}

	if _, ok := sl.RemoveAt(-1); ok {
		t.Fatal("RemoveAt(-1) should return false")
	}
	if _, ok := sl.RemoveAt(sl.Len()); ok {
		t.Fatal("RemoveAt(len) should return false")
	}
}

func TestSyncLinkedList_Clear(t *testing.T) {
	t.Parallel()

	sl := NewSyncLinkedList[int]()
	sl.AddLast(1)
	sl.AddLast(2)
	sl.AddLast(3)

	sl.Clear()
	if !sl.IsEmpty() || sl.Len() != 0 {
		t.Fatalf("Clear() failed, len=%d", sl.Len())
	}
	if _, ok := sl.First(); ok {
		t.Fatal("First() should be false after Clear")
	}
	if _, ok := sl.Last(); ok {
		t.Fatal("Last() should be false after Clear")
	}
}

func TestSyncLinkedList_String(t *testing.T) {
	t.Parallel()
	sl := NewSyncLinkedList[int]()
	sl.AddLast(1)
	sl.AddLast(2)
	if got, want := sl.String(), "[1, 2]"; got != want {
		t.Fatalf("Sync String()=%q want %q", got, want)
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

// Random access pattern (realistic workload)
func BenchmarkLinkedList_GetRandom(b *testing.B) {
	const N = 10_000
	l := NewLinkedList[int]()
	for i := 0; i < N; i++ {
		l.AddLast(i)
	}
	rng := rand.New(rand.NewSource(42))
	b.ResetTimer()
	for b.Loop() {
		idx := rng.Intn(N)
		_, _ = l.Get(idx)
	}
}

// Sequential forward access (best-case cache behavior)
func BenchmarkLinkedList_GetSequential(b *testing.B) {
	const N = 10_000
	l := NewLinkedList[int]()
	for i := 0; i < N; i++ {
		l.AddLast(i)
	}
	b.ResetTimer()
	for b.Loop() {
		for i := 0; i < N; i++ {
			_, _ = l.Get(i)
		}
	}
}

// Alternating ends (worst-case for branch prediction?)
func BenchmarkLinkedList_GetAlternating(b *testing.B) {
	const N = 10_000
	l := NewLinkedList[int]()
	for i := 0; i < N; i++ {
		l.AddLast(i)
	}
	b.ResetTimer()
	for b.Loop() {
		for i := 0; i < N; i++ {
			idx := i % 2 * (N - 1) // alternates 0, 9999, 0, 9999...
			_, _ = l.Get(idx)
		}
	}
}

func BenchmarkLinkedList_IterSequential(b *testing.B) {
	const N = 10_000
	l := NewLinkedList[int]()
	for i := 0; i < N; i++ {
		l.AddLast(i)
	}
	b.ResetTimer()
	for b.Loop() {
		it := l.Iter()
		for {
			val, ok := it.Next()
			if !ok {
				break
			}
			// Prevent compiler from optimizing away the loop
			_ = val
		}
	}
}
