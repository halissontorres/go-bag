package sketch

import (
	"hash/fnv"
	"math"
	"sync"
	"testing"
)

// hashUint64 is a splitmix64-style mixer that distributes sequential integers
// uniformly across the 64-bit space — needed for HyperLogLog correctness.
func hashUint64(x uint64) uint64 {
	x = (x ^ (x >> 30)) * 0xbf58476d1ce4e5b9
	x = (x ^ (x >> 27)) * 0x94d049bb133111eb
	x ^= x >> 31
	return x
}

// TestHyperLogLog_BasicCount verifies cardinality estimates for small sets.
func TestHyperLogLog_BasicCount(t *testing.T) {
	hll := NewHyperLogLog(12, hashString)

	items := []string{"a", "b", "c", "d", "e", "a", "b", "c"}
	for _, item := range items {
		hll.Add(item)
	}

	// 5 distinct items (a, b, c, d, e).
	count := hll.Count()
	expected := uint64(5)

	// With p=12 error ~1.6%, so tolerance is generous.
	errPct := math.Abs(float64(count)-float64(expected)) / float64(expected) * 100
	if errPct > 10 {
		t.Errorf("expected ~%d, got %d (%.1f%% error)", expected, count, errPct)
	}
}

// TestHyperLogLog_LargeScale verifies accuracy with many items.
func TestHyperLogLog_LargeScale(t *testing.T) {
	hll := NewHyperLogLog(14, hashUint64) // p=14 → ~0.81% error

	const n = 100_000
	for i := range uint64(n) {
		hll.Add(i)
	}

	count := hll.Count()
	errPct := math.Abs(float64(count)-float64(n)) / float64(n) * 100

	// p=14 gives ~1.04/sqrt(16384) ≈ 0.81% theoretical error.
	// Allow 2x that in practice for a single run.
	if errPct > 3.0 {
		t.Errorf("expected ~%d, got %d (%.2f%% error, want < 3%%)", n, count, errPct)
	}
}

// TestHyperLogLog_EmptySketch returns 0 for no items.
func TestHyperLogLog_EmptySketch(t *testing.T) {
	hll := NewHyperLogLog(10, hashString)

	if count := hll.Count(); count != 0 {
		t.Errorf("expected 0 for empty sketch, got %d", count)
	}
}

// TestHyperLogLog_AllSameItem returns 1.
func TestHyperLogLog_AllSameItem(t *testing.T) {
	hll := NewHyperLogLog(10, hashInt)

	for range 10000 {
		hll.Add(42)
	}

	if count := hll.Count(); count != 1 {
		t.Errorf("expected 1 for single distinct item, got %d", count)
	}
}

// TestHyperLogLog_Merge combines two sketches.
func TestHyperLogLog_Merge(t *testing.T) {
	a := NewHyperLogLog(12, hashUint64)
	b := NewHyperLogLog(12, hashUint64)

	for i := range uint64(1000) {
		a.Add(i)
	}
	for i := range uint64(1000) {
		b.Add(i + 500) // 500..1499
	}

	ok := a.Merge(b)
	if !ok {
		t.Fatal("merge should succeed with same precision")
	}

	count := a.Count()
	expected := uint64(1500)

	errPct := math.Abs(float64(count)-float64(expected)) / float64(expected) * 100
	if errPct > 10 {
		t.Errorf("expected ~%d after merge, got %d (%.1f%% error)", expected, count, errPct)
	}
}

// TestHyperLogLog_MergeDifferentPrecision fails.
func TestHyperLogLog_MergeDifferentPrecision(t *testing.T) {
	a := NewHyperLogLog(10, hashInt)
	b := NewHyperLogLog(12, hashInt)

	if a.Merge(b) {
		t.Error("merge with different precision should return false")
	}
}

// TestHyperLogLog_Precision validates precision clamping.
func TestHyperLogLog_Precision(t *testing.T) {
	tests := []struct {
		input, expected uint8
	}{
		{0, 4},  // clamped to min
		{3, 4},  // clamped to min
		{4, 4},
		{14, 14},
		{18, 18},
		{19, 18}, // clamped to max
		{20, 18}, // clamped to max
	}

	for _, tt := range tests {
		hll := NewHyperLogLog(tt.input, hashInt)
		if hll.Precision() != tt.expected {
			t.Errorf("precision(%d): expected %d, got %d", tt.input, tt.expected, hll.Precision())
		}
	}
}

// TestHyperLogLog_CustomType works with structs.
func TestHyperLogLog_CustomType(t *testing.T) {
	type user struct {
		id    int64
		email string
	}

	hashUser := func(u user) uint64 {
		h := fnv.New64a()
		var buf [8]byte
		buf[0] = byte(u.id >> 56)
		buf[1] = byte(u.id >> 48)
		buf[2] = byte(u.id >> 40)
		buf[3] = byte(u.id >> 32)
		buf[4] = byte(u.id >> 24)
		buf[5] = byte(u.id >> 16)
		buf[6] = byte(u.id >> 8)
		buf[7] = byte(u.id)
		h.Write(buf[:])
		h.Write([]byte(u.email))
		return h.Sum64()
	}

	hll := NewHyperLogLog(12, hashUser)

	for i := range 500 {
		hll.Add(user{id: int64(i), email: "user@example.com"})
	}

	count := hll.Count()
	errPct := math.Abs(float64(count)-500) / 500 * 100
	if errPct > 10 {
		t.Errorf("expected ~500, got %d (%.1f%% error)", count, errPct)
	}
}

// TestSyncHyperLogLog_ConcurrentAdd verifies concurrent inserts.
func TestSyncHyperLogLog_ConcurrentAdd(t *testing.T) {
	hll := NewSyncHyperLogLog(14, hashUint64)

	const goroutines = 20
	const itemsPerGoroutine = 5000

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func(base uint64) {
			defer wg.Done()
			for i := uint64(0); i < itemsPerGoroutine; i++ {
				hll.Add(base*uint64(1<<20) + i)
			}
		}(uint64(g))
	}
	wg.Wait()

	expected := uint64(goroutines * itemsPerGoroutine)
	count := hll.Count()
	errPct := math.Abs(float64(count)-float64(expected)) / float64(expected) * 100
	if errPct > 3.0 {
		t.Errorf("expected ~%d, got %d (%.2f%% error)", expected, count, errPct)
	}
}

// TestSyncHyperLogLog_ConcurrentCount verifies read/write safety.
func TestSyncHyperLogLog_ConcurrentCount(t *testing.T) {
	hll := NewSyncHyperLogLog(12, hashUint64)

	for i := range uint64(1000) {
		hll.Add(i)
	}

	var wg sync.WaitGroup
	wg.Add(50)
	for range 25 {
		go func() {
			defer wg.Done()
			for range 200 {
				hll.Add(uint64(1000))
			}
		}()
	}
	for range 25 {
		go func() {
			defer wg.Done()
			for range 200 {
				_ = hll.Count()
			}
		}()
	}
	wg.Wait()

	// Only 1000 distinct values were added (the rest are duplicates of 1000).
	count := hll.Count()
	errPct := math.Abs(float64(count)-1000) / 1000 * 100
	if errPct > 20 {
		t.Errorf("expected ~1000, got %d (%.1f%% error)", count, errPct)
	}
}

func BenchmarkHyperLogLog_Add(b *testing.B) {
	hll := NewHyperLogLog(14, hashUint64)
	b.ResetTimer()
	b.ReportAllocs()

	for i := range uint64(b.N) {
		hll.Add(i)
	}
}

func BenchmarkHyperLogLog_Count(b *testing.B) {
	hll := NewHyperLogLog(14, hashUint64)
	for i := range uint64(100000) {
		hll.Add(i)
	}
	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		_ = hll.Count()
	}
}

func BenchmarkHyperLogLog_AddAndCount(b *testing.B) {
	hll := NewHyperLogLog(14, hashUint64)
	b.ResetTimer()
	b.ReportAllocs()

	for i := range uint64(b.N) {
		hll.Add(i)
		if i%10000 == 0 {
			_ = hll.Count()
		}
	}
}

func BenchmarkSyncHyperLogLog_Parallel(b *testing.B) {
	hll := NewSyncHyperLogLog(14, hashUint64)
	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := uint64(0)
		for pb.Next() {
			hll.Add(i)
			i++
		}
	})
}
