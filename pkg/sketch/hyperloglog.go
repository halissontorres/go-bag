package sketch

import (
	"math"
	"math/bits"
	"sync"
)

const (
	hllMinPrecision = 4
	hllMaxPrecision = 18
)

// alphaInf is the asymptotic value of alpha_m for large m.
const alphaInf = 0.7213

// alpha returns the bias correction constant for m registers.
func alpha(m uint64) float64 {
	switch m {
	case 16:
		return 0.673
	case 32:
		return 0.697
	case 64:
		return 0.709
	default:
		return alphaInf / (1.0 + 1.079/float64(m))
	}
}

// hyperLogLog holds the register array and parameters.
type hyperLogLog[T any] struct {
	registers []uint8
	precision uint8
	m         uint64 // number of registers (2^precision)
	hash      func(T) uint64
}

// HyperLogLog estimates the cardinality of a multiset (the number of distinct
// elements). It achieves a relative error of approximately 1.04/√m, where
// m = 2^precision is the number of registers.
//
// Memory usage: m bytes (e.g., precision=14 → 16 KB, error ≈ 0.81%).
//
// HyperLogLog is not safe for concurrent use. Use SyncHyperLogLog for shared
// access.
type HyperLogLog[T any] struct {
	hll *hyperLogLog[T]
}

// NewHyperLogLog creates a HyperLogLog with the given precision.
//
// precision controls the accuracy/memory trade-off: the sketch uses 2^precision
// one-byte registers. Valid range is [4, 18].
//
//	precision | registers | error
//	    4     |    16     | ~26%
//	    8     |   256     | ~6.5%
//	   12     |  4096     | ~1.6%
//	   14     | 16384     | ~0.81%
//	   16     | 65536     | ~0.40%
//	   18     | 262144    | ~0.20%
//
// The hash function must distribute items uniformly across uint64.
func NewHyperLogLog[T any](precision uint8, hash func(T) uint64, opts ...HyperLogLogOption[T]) *HyperLogLog[T] {
	if precision < hllMinPrecision {
		precision = hllMinPrecision
	}
	if precision > hllMaxPrecision {
		precision = hllMaxPrecision
	}
	m := uint64(1) << precision
	h := &hyperLogLog[T]{
		registers: make([]uint8, m),
		precision: precision,
		m:         m,
		hash:      hash,
	}
	for _, o := range opts {
		o(h)
	}
	return &HyperLogLog[T]{hll: h}
}

// Add inserts an item into the sketch.
func (h *HyperLogLog[T]) Add(item T) {
	x := h.hll.hash(item)
	// Avalanche mixer ensures good distribution regardless of input hash quality.
	x = mix64(x)
	// Lower p bits select the register.
	bucket := x & (h.hll.m - 1)
	// Remaining bits shifted right; top p bits are zero in the uint64.
	// The rank ρ is the position of the leftmost 1-bit (1-indexed) in the
	// (64−p)-bit window. Compensate for the p leading zeros of the uint64.
	w := x >> h.hll.precision
	rank := uint8(int(bits.LeadingZeros64(w)) - int(h.hll.precision) + 1)

	if rank > h.hll.registers[bucket] {
		h.hll.registers[bucket] = rank
	}
}

// Count returns the estimated number of distinct items added.
func (h *HyperLogLog[T]) Count() uint64 {
	m := h.hll.m
	regs := h.hll.registers

	// Harmonic mean via Σ 2^(-reg[j]).
	var sum float64
	var zeros uint64
	for _, r := range regs {
		if r == 0 {
			zeros++
		}
		sum += math.Exp2(-float64(r))
	}

	raw := alpha(m) * float64(m*m) / sum

	// Linear counting correction for small cardinalities.
	if raw <= 2.5*float64(m) && zeros > 0 {
		raw = float64(m) * math.Log(float64(m)/float64(zeros))
	}

	return uint64(raw + 0.5)
}

// Merge combines another HyperLogLog into this one. Both must have the same
// precision, otherwise Merge returns false without modifying h.
//
// Merge is not goroutine-safe.
func (h *HyperLogLog[T]) Merge(other *HyperLogLog[T]) bool {
	if h.hll.precision != other.hll.precision {
		return false
	}
	for i := range h.hll.registers {
		if other.hll.registers[i] > h.hll.registers[i] {
			h.hll.registers[i] = other.hll.registers[i]
		}
	}
	return true
}

// Precision returns the precision p (number of registers = 2^p).
func (h *HyperLogLog[T]) Precision() uint8 {
	return h.hll.precision
}

// mix64 is a splitmix64-style finalizer that uniformly distributes hash bits.
func mix64(x uint64) uint64 {
	x = (x ^ (x >> 30)) * 0xbf58476d1ce4e5b9
	x = (x ^ (x >> 27)) * 0x94d049bb133111eb
	x ^= x >> 31
	return x
}

// SyncHyperLogLog is a thread-safe HyperLogLog.
type SyncHyperLogLog[T any] struct {
	mu  sync.RWMutex
	hll *HyperLogLog[T]
}

// NewSyncHyperLogLog creates a thread-safe HyperLogLog.
func NewSyncHyperLogLog[T any](precision uint8, hash func(T) uint64, opts ...HyperLogLogOption[T]) *SyncHyperLogLog[T] {
	return &SyncHyperLogLog[T]{hll: NewHyperLogLog[T](precision, hash, opts...)}
}

// Add inserts an item atomically.
func (s *SyncHyperLogLog[T]) Add(item T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hll.Add(item)
}

// Count returns the estimated distinct count.
func (s *SyncHyperLogLog[T]) Count() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.hll.Count()
}

// Merge combines another HyperLogLog into this one atomically.
// Returns false if precisions differ.
func (s *SyncHyperLogLog[T]) Merge(other *HyperLogLog[T]) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.hll.Merge(other)
}

// Precision returns the precision p.
func (s *SyncHyperLogLog[T]) Precision() uint8 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.hll.Precision()
}
