package cache

import (
	"slices"
	"sync"
	"testing"
)

func TestLRUCache_PutAndGet(t *testing.T) {
	t.Parallel()
	c := New[string, int](WithCapacity[string, int](3))
	c.Put("a", 1)
	c.Put("b", 2)
	c.Put("c", 3)

	v, ok := c.Get("a")
	if !ok || v != 1 {
		t.Errorf("Get(a) = (%v, %v), want (1, true)", v, ok)
	}
	v, ok = c.Get("b")
	if !ok || v != 2 {
		t.Errorf("Get(b) = (%v, %v), want (2, true)", v, ok)
	}
	_, ok = c.Get("nonexistent")
	if ok {
		t.Error("Get(nonexistent) should be false")
	}
}

func TestLRUCache_PutUpdate(t *testing.T) {
	t.Parallel()
	c := New[string, string](WithCapacity[string, string](3))
	c.Put("k", "v1")
	c.Put("k", "v2")

	v, ok := c.Get("k")
	if !ok || v != "v2" {
		t.Errorf("Get(k) = (%v, %v), want (v2, true)", v, ok)
	}
	if c.Len() != 1 {
		t.Errorf("Len = %d, want 1", c.Len())
	}
}

func TestLRUCache_GetUpdatesAccessOrder(t *testing.T) {
	t.Parallel()
	c := New[int, string](WithCapacity[int, string](3))
	c.Put(1, "a")
	c.Put(2, "b")
	c.Put(3, "c")

	c.Get(1) // 1 becomes MRU, 2 is now LRU

	c.Put(4, "d") // should evict 2

	if c.Contains(2) {
		t.Error("key 2 should have been evicted")
	}
	if !c.Contains(1) {
		t.Error("key 1 should still be present")
	}
	if !c.Contains(3) {
		t.Error("key 3 should still be present")
	}
	if !c.Contains(4) {
		t.Error("key 4 should be present")
	}
}

func TestLRUCache_PeekDoesNotUpdateOrder(t *testing.T) {
	t.Parallel()
	c := New[int, string](WithCapacity[int, string](3))
	c.Put(1, "a")
	c.Put(2, "b")
	c.Put(3, "c")

	c.Peek(1) // should NOT change order

	c.Put(4, "d") // should evict 1 (LRU)

	if c.Contains(1) {
		t.Error("key 1 should have been evicted")
	}
	if !c.Contains(2) {
		t.Error("key 2 should still be present")
	}
}

func TestLRUCache_EvictsLRUOnOverflow(t *testing.T) {
	t.Parallel()
	c := New[int, string](WithCapacity[int, string](3))
	c.Put(1, "a")
	c.Put(2, "b")
	c.Put(3, "c")
	c.Put(4, "d")

	if c.Contains(1) {
		t.Error("key 1 (LRU) should have been evicted")
	}
	if c.Len() != 3 {
		t.Errorf("Len = %d, want 3", c.Len())
	}
}

func TestLRUCache_PutUpdateDoesNotEvict(t *testing.T) {
	t.Parallel()
	c := New[int, string](WithCapacity[int, string](3))
	c.Put(1, "a")
	c.Put(2, "b")
	c.Put(3, "c")
	c.Put(1, "updated") // update, not new insert

	if c.Len() != 3 {
		t.Errorf("Len = %d, want 3", c.Len())
	}
	v, _ := c.Get(1)
	if v != "updated" {
		t.Errorf("Get(1) = %v, want updated", v)
	}
}

func TestLRUCache_Remove(t *testing.T) {
	t.Parallel()
	c := New[int, string](WithCapacity[int, string](3))
	c.Put(1, "a")
	c.Put(2, "b")

	v, ok := c.Remove(1)
	if !ok || v != "a" {
		t.Errorf("Remove(1) = (%v, %v), want (a, true)", v, ok)
	}
	if c.Contains(1) {
		t.Error("key 1 should be removed")
	}
	if c.Len() != 1 {
		t.Errorf("Len = %d, want 1", c.Len())
	}

	_, ok = c.Remove(99)
	if ok {
		t.Error("Remove(99) should be false")
	}
}

func TestLRUCache_Contains(t *testing.T) {
	t.Parallel()
	c := New[string, int](WithCapacity[string, int](3))
	c.Put("x", 1)

	if !c.Contains("x") {
		t.Error("Contains(x) should be true")
	}
	if c.Contains("y") {
		t.Error("Contains(y) should be false")
	}
}

func TestLRUCache_LenAndCap(t *testing.T) {
	t.Parallel()
	c := New[int, int](WithCapacity[int, int](5))
	c.Put(1, 10)
	c.Put(2, 20)

	if c.Len() != 2 {
		t.Errorf("Len = %d, want 2", c.Len())
	}
	if c.Cap() != 5 {
		t.Errorf("Cap = %d, want 5", c.Cap())
	}
}

func TestLRUCache_Keys_Order(t *testing.T) {
	t.Parallel()
	c := New[int, string](WithCapacity[int, string](4))
	c.Put(1, "a")
	c.Put(2, "b")
	c.Put(3, "c")
	c.Get(1) // 1 becomes MRU → expected order: 1, 3, 2
	c.Put(4, "d")

	want := []int{4, 1, 3, 2}
	got := c.Keys()
	if !slices.Equal(got, want) {
		t.Errorf("Keys() = %v, want %v", got, want)
	}
}

func TestLRUCache_Values_Order(t *testing.T) {
	t.Parallel()
	c := New[string, string](WithCapacity[string, string](3))
	c.Put("k1", "v1")
	c.Put("k2", "v2")
	c.Put("k3", "v3")
	c.Get("k1")

	want := []string{"v1", "v3", "v2"}
	got := c.Values()
	if !slices.Equal(got, want) {
		t.Errorf("Values() = %v, want %v", got, want)
	}
}

func TestLRUCache_Clear(t *testing.T) {
	t.Parallel()
	c := New[int, string](WithCapacity[int, string](3))
	c.Put(1, "a")
	c.Put(2, "b")
	c.Clear()

	if c.Len() != 0 {
		t.Errorf("Len = %d, want 0", c.Len())
	}
	if c.Contains(1) {
		t.Error("Contains(1) should be false after Clear")
	}
	if len(c.Keys()) != 0 {
		t.Error("Keys() should be empty after Clear")
	}
}

func TestLRUCache_Resize_Shrink(t *testing.T) {
	t.Parallel()
	c := New[int, string](WithCapacity[int, string](5))
	c.Put(1, "a")
	c.Put(2, "b")
	c.Put(3, "c")
	c.Put(4, "d")
	c.Put(5, "e")

	c.Resize(2)

	if c.Cap() != 2 {
		t.Errorf("Cap = %d, want 2", c.Cap())
	}
	if c.Len() != 2 {
		t.Errorf("Len = %d, want 2", c.Len())
	}
	// 1 and 2 should be evicted (they're LRU)
	if c.Contains(1) || c.Contains(2) || c.Contains(3) {
		t.Error("keys 1, 2, 3 should have been evicted")
	}
	if !c.Contains(4) && !c.Contains(5) {
		t.Error("keys 4 and 5 should remain (most recent)")
	}
}

func TestLRUCache_Resize_Grow(t *testing.T) {
	t.Parallel()
	c := New[int, string](WithCapacity[int, string](2))
	c.Put(1, "a")
	c.Put(2, "b")
	c.Resize(5)
	c.Put(3, "c")
	c.Put(4, "d")

	if c.Len() != 4 {
		t.Errorf("Len = %d, want 4", c.Len())
	}
}

func TestLRUCache_OnEvict(t *testing.T) {
	t.Parallel()
	var evictedKey int
	var evictedVal string
	c := New[int, string](
		WithCapacity[int, string](2),
		WithOnEvict[int, string](func(k int, v string) {
			evictedKey = k
			evictedVal = v
		}),
	)
	c.Put(1, "a")
	c.Put(2, "b")
	c.Put(3, "c") // evicts 1

	if evictedKey != 1 || evictedVal != "a" {
		t.Errorf("evicted = (%d, %s), want (1, a)", evictedKey, evictedVal)
	}

	c.Remove(2) // also fires onEvict
	if evictedKey != 2 || evictedVal != "b" {
		t.Errorf("after Remove, evicted = (%d, %s), want (2, b)", evictedKey, evictedVal)
	}
}

func TestLRUCache_OnEvict_Clear(t *testing.T) {
	t.Parallel()
	var evicted []int
	c := New[int, string](
		WithCapacity[int, string](3),
		WithOnEvict[int, string](func(k int, v string) {
			evicted = append(evicted, k)
		}),
	)
	c.Put(1, "a")
	c.Put(2, "b")
	c.Put(3, "c")
	c.Clear()

	if len(evicted) != 3 {
		t.Errorf("onEvict called %d times, want 3", len(evicted))
	}
}

func TestLRUCache_ZeroCapacityPanics(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("New with WithCapacity(0) should panic")
		}
	}()
	New[int, string](WithCapacity[int, string](0))
}

func TestLRUCache_NoCapacityPanics(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("New without WithCapacity should panic")
		}
	}()
	New[int, string]()
}

func TestLRUCache_ResizeZeroPanics(t *testing.T) {
	t.Parallel()
	c := New[int, string](WithCapacity[int, string](3))
	defer func() {
		if r := recover(); r == nil {
			t.Error("Resize(0) should panic")
		}
	}()
	c.Resize(0)
}

func TestLRUCache_StructKeys(t *testing.T) {
	t.Parallel()
	type key struct{ ID int }
	c := New[key, string](WithCapacity[key, string](2))
	c.Put(key{1}, "a")
	c.Put(key{2}, "b")

	v, ok := c.Get(key{1})
	if !ok || v != "a" {
		t.Errorf("Get({1}) = (%v, %v), want (a, true)", v, ok)
	}
}

// ---------- SyncLRUCache tests ----------

func TestSyncLRUCache_Basic(t *testing.T) {
	t.Parallel()
	sc := NewSync[int, string](WithCapacity[int, string](3))
	sc.Put(1, "a")
	sc.Put(2, "b")

	v, ok := sc.Get(1)
	if !ok || v != "a" {
		t.Errorf("Get(1) = (%v, %v), want (a, true)", v, ok)
	}
	if sc.Len() != 2 {
		t.Errorf("Len = %d, want 2", sc.Len())
	}
	if sc.Cap() != 3 {
		t.Errorf("Cap = %d, want 3", sc.Cap())
	}
}

func TestSyncLRUCache_ConcurrentPut(t *testing.T) {
	t.Parallel()
	const goroutines = 20
	const itemsEach = 500
	sc := NewSync[int, int](WithCapacity[int, int](goroutines * itemsEach))

	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(base int) {
			defer wg.Done()
			for i := 0; i < itemsEach; i++ {
				sc.Put(base*itemsEach+i, i)
			}
		}(g)
	}
	wg.Wait()

	if sc.Len() != goroutines*itemsEach {
		t.Errorf("Len = %d, want %d", sc.Len(), goroutines*itemsEach)
	}
}

func TestSyncLRUCache_ConcurrentGet(t *testing.T) {
	t.Parallel()
	sc := NewSync[int, string](WithCapacity[int, string](100))
	for i := 0; i < 100; i++ {
		sc.Put(i, "val")
	}

	const goroutines = 20
	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				v, ok := sc.Get(i)
				if !ok || v != "val" {
					t.Errorf("Get(%d) = (%v, %v)", i, v, ok)
				}
			}
		}()
	}
	wg.Wait()
}

func TestSyncLRUCache_ConcurrentMixed(t *testing.T) {
	t.Parallel()
	sc := NewSync[int, int](WithCapacity[int, int](1000))
	for i := 0; i < 500; i++ {
		sc.Put(i, i*10)
	}

	var wg sync.WaitGroup
	// Writers
	for g := 0; g < 10; g++ {
		wg.Add(1)
		go func(base int) {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				sc.Put(500+base*50+i, i)
			}
		}(g)
	}
	// Readers
	for g := 0; g < 10; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				sc.Get(i)
				sc.Contains(i)
				sc.Len()
				sc.Keys()
			}
		}()
	}
	wg.Wait()

	if sc.Len() != 1000 {
		t.Errorf("Len = %d, want 1000", sc.Len())
	}
}

// ---------- Benchmarks ----------

func BenchmarkCache_Put(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		b.StopTimer()
		c := New[int, int](WithCapacity[int, int](b.N+1))
		b.StartTimer()
		for i := 0; i < b.N; i++ {
			c.Put(i, i)
		}
	}
}

func BenchmarkCache_Get(b *testing.B) {
	c := New[int, int](WithCapacity[int, int](10_000))
	for i := 0; i < 10_000; i++ {
		c.Put(i, i)
	}
	b.ReportAllocs()
	b.ResetTimer()
	i := 0
	for b.Loop() {
		c.Get(i % 10000)
		i++
	}
}

func BenchmarkCache_PutWithEviction(b *testing.B) {
	b.ReportAllocs()
	c := New[int, int](WithCapacity[int, int](1))
	for b.Loop() {
		c.Put(0, 0)
	}
}
