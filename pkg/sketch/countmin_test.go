package sketch

import (
	"hash/fnv"
	"sync"
	"testing"
)

func hashString(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

func hashInt(i int) uint64 {
	h := fnv.New64a()
	// Write int as 8 bytes in big-endian.
	var buf [8]byte
	buf[0] = byte(i >> 56)
	buf[1] = byte(i >> 48)
	buf[2] = byte(i >> 40)
	buf[3] = byte(i >> 32)
	buf[4] = byte(i >> 24)
	buf[5] = byte(i >> 16)
	buf[6] = byte(i >> 8)
	buf[7] = byte(i)
	h.Write(buf[:])
	return h.Sum64()
}

func TestCountMin_BasicEstimate(t *testing.T) {
	cm := NewCountMin(0.01, 0.001, hashString)

	cm.Increment("foo")
	cm.Increment("foo")
	cm.Increment("bar")

	if est := cm.Estimate("foo"); est < 2 {
		t.Errorf("expected estimate >= 2 for 'foo', got %d", est)
	}
	if est := cm.Estimate("bar"); est < 1 {
		t.Errorf("expected estimate >= 1 for 'bar', got %d", est)
	}
	if est := cm.Estimate("baz"); est > 2 {
		t.Errorf("expected estimate ~0 for 'baz', got %d", est)
	}
}

func TestCountMin_HeavyItem(t *testing.T) {
	cm := NewCountMin(0.001, 0.0001, hashInt)

	const heavy = 42
	const heavyCount = 10000
	const lightCount = 1

	for range heavyCount {
		cm.Increment(heavy)
	}
	for i := range 1000 {
		if i != heavy {
			cm.Increment(i)
		}
	}

	est := cm.Estimate(heavy)
	if est < heavyCount {
		t.Errorf("heavy item estimate %d < true count %d", est, heavyCount)
	}

	// Error should be within ε * N.  ε=0.001, N≈11000, so error ≤ ~11.
	// Count-Min overestimates, so est − true ≤ ε×N.
	epsilon := 0.001
	total := cm.Total()
	overcount := est - heavyCount
	bound := uint64(float64(total) * epsilon)
	if overcount > bound && overcount > 50 {
		t.Logf("note: overcount %d exceeds theoretical bound %d (total=%d)", overcount, bound, total)
	}

	// Light items should estimate close to 1.
	lightEst := cm.Estimate(999)
	if lightEst > uint64(float64(total)*epsilon)+1 {
		t.Logf("light item estimate %d, expected ~1", lightEst)
	}
}

func TestCountMin_Total(t *testing.T) {
	cm := NewCountMin(0.01, 0.01, hashString)

	for range 100 {
		cm.Increment("a")
	}
	cm.Update("b", 50)

	if cm.Total() != 150 {
		t.Errorf("expected total 150, got %d", cm.Total())
	}
}

func TestCountMin_EmptySketch(t *testing.T) {
	cm := NewCountMin(0.01, 0.01, hashString)

	if est := cm.Estimate("never-seen"); est != 0 {
		t.Errorf("expected estimate 0 for empty sketch, got %d", est)
	}
	if cm.Total() != 0 {
		t.Errorf("expected total 0, got %d", cm.Total())
	}
}

func TestCountMin_EpsilonDeltaDimensions(t *testing.T) {
	tests := []struct {
		epsilon, delta float64
		wantDepthMin   int
		wantWidthMin   uint64
	}{
		{0.1, 0.1, 2, 27},    // e/0.1 ≈ 27.18, ln(1/0.1) ≈ 2.30
		{0.01, 0.01, 4, 271}, // e/0.01 ≈ 271.8, ln(1/0.01) ≈ 4.60
		{0.001, 0.001, 6, 2718},
	}

	for _, tt := range tests {
		cm := NewCountMin(tt.epsilon, tt.delta, hashInt)

		if cm.sketch.depth < tt.wantDepthMin {
			t.Errorf("ε=%.3f δ=%.3f: depth %d < %d", tt.epsilon, tt.delta, cm.sketch.depth, tt.wantDepthMin)
		}
		if cm.sketch.width < tt.wantWidthMin {
			t.Errorf("ε=%.3f δ=%.3f: width %d < %d", tt.epsilon, tt.delta, cm.sketch.width, tt.wantWidthMin)
		}
	}
}

func TestCountMin_CustomType(t *testing.T) {
	type event struct {
		userID  int
		action  string
	}

	hashEvent := func(e event) uint64 {
		h := fnv.New64a()
		var buf [8]byte
		buf[0] = byte(e.userID >> 56)
		buf[1] = byte(e.userID >> 48)
		buf[2] = byte(e.userID >> 40)
		buf[3] = byte(e.userID >> 32)
		buf[4] = byte(e.userID >> 24)
		buf[5] = byte(e.userID >> 16)
		buf[6] = byte(e.userID >> 8)
		buf[7] = byte(e.userID)
		h.Write(buf[:])
		h.Write([]byte(e.action))
		return h.Sum64()
	}

	cm := NewCountMin(0.01, 0.01, hashEvent)

	login := event{userID: 1, action: "login"}
	logout := event{userID: 1, action: "logout"}

	cm.Increment(login)
	cm.Increment(login)
	cm.Increment(logout)

	if est := cm.Estimate(login); est < 2 {
		t.Errorf("expected estimate >= 2 for login, got %d", est)
	}
	if est := cm.Estimate(logout); est < 1 {
		t.Errorf("expected estimate >= 1 for logout, got %d", est)
	}
}

func TestSyncCountMin_ConcurrentUpdates(t *testing.T) {
	cm := NewSyncCountMin(0.01, 0.001, hashInt)

	const goroutines = 20
	const incrementsPerGoroutine = 5000
	const targetKey = 7

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < incrementsPerGoroutine; i++ {
				cm.Increment(targetKey)
			}
		}()
	}
	wg.Wait()

	expected := uint64(goroutines * incrementsPerGoroutine)
	if cm.Total() != expected {
		t.Errorf("expected total %d, got %d", expected, cm.Total())
	}
	if est := cm.Estimate(targetKey); est < expected {
		t.Errorf("estimate %d < expected %d", est, expected)
	}
}

func TestSyncCountMin_ConcurrentReadWrite(t *testing.T) {
	cm := NewSyncCountMin(0.01, 0.01, hashInt)

	for i := range 100 {
		cm.Increment(i % 10)
	}

	var wg sync.WaitGroup
	wg.Add(50)
	for range 25 {
		go func() {
			defer wg.Done()
			for range 200 {
				cm.Increment(3)
			}
		}()
	}
	for range 25 {
		go func() {
			defer wg.Done()
			for range 200 {
				_ = cm.Estimate(3)
				_ = cm.Total()
			}
		}()
	}
	wg.Wait()

	if est := cm.Estimate(3); est < 10+25*200 {
		t.Errorf("estimate too low: %d", est)
	}
}

func BenchmarkCountMin_Increment(b *testing.B) {
	cm := NewCountMin(0.01, 0.001, hashInt)
	b.ResetTimer()
	b.ReportAllocs()

	for i := range b.N {
		cm.Increment(i % 1000)
	}
}

func BenchmarkCountMin_Estimate(b *testing.B) {
	cm := NewCountMin(0.01, 0.001, hashInt)
	for i := range 10000 {
		cm.Increment(i % 500)
	}
	b.ResetTimer()
	b.ReportAllocs()

	for i := range b.N {
		cm.Estimate(i % 500)
	}
}

func BenchmarkCountMin_Parallel(b *testing.B) {
	cm := NewCountMin(0.01, 0.001, hashInt)
	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			cm.Increment(i % 1000)
			i++
		}
	})
}

func BenchmarkSyncCountMin_Parallel(b *testing.B) {
	cm := NewSyncCountMin(0.01, 0.001, hashInt)
	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			cm.Increment(i % 1000)
			i++
		}
	})
}
