package set

import (
	"fmt"
	"math/bits"
	"strings"
)

// Enum is the interface an enum type must implement to be used with EnumSet.
type Enum interface {
	Index() int
}

// EnumSet is a highly efficient set for enums.
// It uses an internal bitmap, making operations such as Union and Intersection O(words).
type EnumSet[T Enum] struct {
	bits []uint64
}

// NewEnumSet creates a new, empty EnumSet.
func NewEnumSet[T Enum]() *EnumSet[T] {
	return &EnumSet[T]{
		bits: make([]uint64, 0),
	}
}

// Add inserts one or more values into the set.
func (es *EnumSet[T]) Add(values ...T) {
	for _, v := range values {
		idx := v.Index()
		if idx < 0 {
			continue
		}
		es.ensureCapacity(idx)
		es.bits[idx/64] |= 1 << (idx % 64)
	}
}

// Remove deletes one or more values from the set.
func (es *EnumSet[T]) Remove(values ...T) {
	for _, v := range values {
		idx := v.Index()
		if idx < 0 || idx/64 >= len(es.bits) {
			continue
		}
		es.bits[idx/64] &^= 1 << (idx % 64)
	}
}

// Contains reports whether the value is present in the set.
func (es *EnumSet[T]) Contains(value T) bool {
	idx := value.Index()
	if idx < 0 || idx/64 >= len(es.bits) {
		return false
	}
	return es.bits[idx/64]&(1<<(idx%64)) != 0
}

// Clear removes all elements.
func (es *EnumSet[T]) Clear() {
	es.bits = make([]uint64, len(es.bits))
}

// Len returns the number of elements.
func (es *EnumSet[T]) Len() int {
	count := 0
	for _, w := range es.bits {
		count += bits.OnesCount64(w)
	}
	return count
}

// ForEach calls mapper for each present element.
// mapper receives the index and must return the corresponding enum value.
func (es *EnumSet[T]) ForEach(mapper func(int) T) {
	for i := 0; i < len(es.bits)*64; i++ {
		if es.bits[i/64]&(1<<(i%64)) != 0 {
			mapper(i)
		}
	}
}

// Elements returns every element as a slice of T.
func (es *EnumSet[T]) Elements(mapper func(int) T) []T {
	var result []T
	es.ForEach(mapper)
	return result
}

// Union returns a new EnumSet containing the union of es and other.
func (es *EnumSet[T]) Union(other *EnumSet[T]) *EnumSet[T] {
	result := NewEnumSet[T]()
	maxLen := max(len(es.bits), len(other.bits))
	result.bits = make([]uint64, maxLen)
	for i := 0; i < len(es.bits); i++ {
		result.bits[i] |= es.bits[i]
	}
	for i := 0; i < len(other.bits); i++ {
		result.bits[i] |= other.bits[i]
	}
	return result
}

// Intersection returns a new EnumSet containing the intersection of es and other.
func (es *EnumSet[T]) Intersection(other *EnumSet[T]) *EnumSet[T] {
	result := NewEnumSet[T]()
	minLen := min(len(es.bits), len(other.bits))
	result.bits = make([]uint64, minLen)
	for i := 0; i < minLen; i++ {
		result.bits[i] = es.bits[i] & other.bits[i]
	}
	return result
}

// Difference returns a new EnumSet containing the elements of es that are not in other.
func (es *EnumSet[T]) Difference(other *EnumSet[T]) *EnumSet[T] {
	result := NewEnumSet[T]()
	result.bits = make([]uint64, len(es.bits))
	copy(result.bits, es.bits)
	for i := 0; i < len(es.bits) && i < len(other.bits); i++ {
		result.bits[i] &^= other.bits[i]
	}
	return result
}

// Equal reports whether two EnumSets contain the same elements.
func (es *EnumSet[T]) Equal(other *EnumSet[T]) bool {
	if es == other {
		return true
	}
	if es.Len() != other.Len() {
		return false
	}
	maxLen := max(len(es.bits), len(other.bits))
	for i := 0; i < maxLen; i++ {
		var a, b uint64
		if i < len(es.bits) {
			a = es.bits[i]
		}
		if i < len(other.bits) {
			b = other.bits[i]
		}
		if a != b {
			return false
		}
	}
	return true
}

// String returns a textual representation of the set as a list of indices.
func (es *EnumSet[T]) String() string {
	var indices []string
	es.ForEach(func(idx int) T {
		indices = append(indices, fmt.Sprintf("%d", idx))
		var zero T
		return zero // The mapper signature requires a return value; the result is unused here.
	})
	return "{" + strings.Join(indices, " ") + "}"
}

// SubsetOf reports whether es is a subset of other.
func (es *EnumSet[T]) SubsetOf(other *EnumSet[T]) bool {
	for i := 0; i < len(es.bits); i++ {
		var otherWord uint64
		if i < len(other.bits) {
			otherWord = other.bits[i]
		}
		if es.bits[i]&^otherWord != 0 {
			return false
		}
	}
	return true
}

// ensureCapacity grows the bitmap to make room for the given index.
func (es *EnumSet[T]) ensureCapacity(idx int) {
	wordIdx := idx / 64
	for wordIdx >= len(es.bits) {
		es.bits = append(es.bits, 0)
	}
}
