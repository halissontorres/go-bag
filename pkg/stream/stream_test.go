package stream

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestFromChannel(t *testing.T) {
	ch := make(chan int, 3)
	ch <- 10
	ch <- 20
	ch <- 30
	close(ch)

	result := FromChannel(ch).ToSlice()
	expected := []int{10, 20, 30}
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestFromChannelEmpty(t *testing.T) {
	ch := make(chan int)
	close(ch)

	result := FromChannel(ch).ToSlice()
	if len(result) != 0 {
		t.Errorf("Expected empty slice, got %v", result)
	}
}

func TestFromFunc(t *testing.T) {
	n := 0
	result := FromFunc(func() (int, bool) {
		if n < 4 {
			v := n
			n++
			return v, true
		}
		return 0, false
	}).ToSlice()

	expected := []int{0, 1, 2, 3}
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestFromFuncEmpty(t *testing.T) {
	result := FromFunc(func() (int, bool) { return 0, false }).ToSlice()
	if len(result) != 0 {
		t.Errorf("Expected empty slice, got %v", result)
	}
}

func TestStreamConsumedOnce(t *testing.T) {
	s := FromSlice([]int{1, 2, 3})
	_ = s.ToSlice()
	second := s.ToSlice()
	if len(second) != 0 {
		t.Errorf("Expected empty stream after consumption, got %v", second)
	}
}

func TestFromSlice(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}
	s := FromSlice(input)
	result := s.ToSlice()

	if !reflect.DeepEqual(input, result) {
		t.Errorf("Expected %v, got %v", input, result)
	}
}

func TestFilter(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}
	s := FromSlice(input)
	result := Filter(s, func(x int) bool { return x%2 == 0 }).ToSlice()

	expected := []int{2, 4}
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestMap(t *testing.T) {
	input := []int{1, 2, 3}
	s := FromSlice(input)
	result := Map(s, func(x int) int { return x * 2 }).ToSlice()

	expected := []int{2, 4, 6}
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestFlatMap(t *testing.T) {
	input := []int{1, 2}
	s := FromSlice(input)
	result := FlatMap(s, func(x int) []int { return []int{x, x * 10} }).ToSlice()

	expected := []int{1, 10, 2, 20}
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestDistinct(t *testing.T) {
	input := []int{1, 2, 2, 3, 1, 4}
	s := FromSlice(input)
	result := Distinct(s).ToSlice()

	expected := []int{1, 2, 3, 4}
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestSorted(t *testing.T) {
	input := []int{3, 1, 4, 1, 5, 9, 2}
	s := FromSlice(input)
	result := Sorted(s).ToSlice()

	expected := []int{1, 1, 2, 3, 4, 5, 9}
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestLimit(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}
	s := FromSlice(input)
	result := Limit(s, 3).ToSlice()

	expected := []int{1, 2, 3}
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestSkip(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}
	s := FromSlice(input)
	result := Skip(s, 2).ToSlice()

	expected := []int{3, 4, 5}
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestConcat(t *testing.T) {
	s1 := FromSlice([]int{1, 2})
	s2 := FromSlice([]int{3, 4})
	s3 := FromSlice([]int{5})
	result := Concat(s1, s2, s3).ToSlice()

	expected := []int{1, 2, 3, 4, 5}
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestPeek(t *testing.T) {
	input := []int{1, 2, 3}
	s := FromSlice(input)
	var peeked []int
	result := Peek(s, func(x int) { peeked = append(peeked, x) }).ToSlice()

	if !reflect.DeepEqual(input, result) {
		t.Errorf("Expected result %v, got %v", input, result)
	}
	if !reflect.DeepEqual(input, peeked) {
		t.Errorf("Expected peeked %v, got %v", input, peeked)
	}
}

func TestPipeline(t *testing.T) {
	input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	s := FromSlice(input)

	// (x*2) -> (x > 10) -> Limit(3)
	result := Limit(
		Filter(
			Map(s, func(x int) int { return x * 2 }),
			func(x int) bool { return x > 10 },
		),
		3,
	).ToSlice()

	expected := []int{12, 14, 16}
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestEmptyStream(t *testing.T) {
	s := FromSlice([]int{})
	if len(s.ToSlice()) != 0 {
		t.Error("Expected empty slice")
	}
}

func TestSkipMoreThanAvailable(t *testing.T) {
	s := FromSlice([]int{1, 2})
	result := Skip(s, 10).ToSlice()
	if len(result) != 0 {
		t.Errorf("Expected empty, got %v", result)
	}
}

func TestLimitZero(t *testing.T) {
	s := FromSlice([]int{1, 2, 3})
	result := Limit(s, 0).ToSlice()
	if len(result) != 0 {
		t.Errorf("Expected empty, got %v", result)
	}
}

func TestDistinctStrings(t *testing.T) {
	input := []string{"a", "b", "a", "c"}
	result := Distinct(FromSlice(input)).ToSlice()
	expected := []string{"a", "b", "c"}
	if !reflect.DeepEqual(expected, result) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

func TestAnyAll(t *testing.T) {
	s := FromSlice([]int{2, 4, 6})
	if !s.Any(func(x int) bool { return x%2 == 0 }) {
		t.Error("Any should return true")
	}
	if !s.All(func(x int) bool { return x%2 == 0 }) {
		t.Error("All should return true")
	}

	s2 := FromSlice([]int{1, 3, 5})
	if s2.Any(func(x int) bool { return x%2 == 0 }) {
		t.Error("Any should return false for even numbers")
	}
}

func TestReduce(t *testing.T) {
	s := FromSlice([]int{1, 2, 3, 4})
	sum := s.Reduce(0, func(a, b int) int { return a + b })
	if sum != 10 {
		t.Errorf("Expected 10, got %d", sum)
	}
}

func TestCount(t *testing.T) {
	s := FromSlice([]int{1, 2, 3})
	if count := s.Count(); count != 3 {
		t.Errorf("Expected 3, got %d", count)
	}
}

func TestFindFirst(t *testing.T) {
	s := FromSlice([]int{1, 3, 4, 6, 7})
	result := FindFirst(s, func(x int) bool { return x%2 == 0 })
	v, ok := result.Get()
	if !ok || v != 4 {
		t.Fatalf("FindFirst(even)=%d,%v want 4,true", v, ok)
	}
}

func TestFindFirstNotFound(t *testing.T) {
	s := FromSlice([]int{1, 3, 5})
	result := FindFirst(s, func(x int) bool { return x%2 == 0 })
	if result.IsPresent() {
		t.Fatal("FindFirst should return empty when no element matches")
	}
}

func TestFindFirstEmpty(t *testing.T) {
	s := FromSlice([]int{})
	result := FindFirst(s, func(x int) bool { return true })
	if result.IsPresent() {
		t.Fatal("FindFirst on empty stream should return empty")
	}
}

// Benchmarks

func BenchmarkFromSlice(b *testing.B) {
	input := make([]int, 1000)
	for i := range input {
		input[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := FromSlice(input)
		_ = s.ToSlice()
	}
}

func BenchmarkPipeline(b *testing.B) {
	input := make([]int, 1000)
	for i := range input {
		input[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := FromSlice(input)
		result := Limit(
			Filter(
				Map(s, func(x int) int { return x * 2 }),
				func(x int) bool { return x%3 == 0 },
			),
			100,
		).ToSlice()
		_ = result
	}
}

func BenchmarkSorted(b *testing.B) {
	input := make([]int, 1000)
	for i := range input {
		input[i] = 1000 - i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := FromSlice(input)
		result := Sorted(s).ToSlice()
		_ = result
	}
}

// stream_test.go

// Test operations without intermediate slice allocation
func BenchmarkCount(b *testing.B) {
	input := make([]int, 1000)
	for i := range input {
		input[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := FromSlice(input)
		_ = s.Count()
	}
}

// Test short-circuit in Any/All
func BenchmarkAnyFoundEarly(b *testing.B) {
	input := make([]int, 10000)
	for i := range input {
		input[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := FromSlice(input)
		_ = s.Any(func(x int) bool { return x > 10 }) // stops at index 11
	}
}

func BenchmarkAllShortCircuit(b *testing.B) {
	input := make([]int, 10000)
	for i := range input {
		input[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := FromSlice(input)
		_ = s.All(func(x int) bool { return x < 100 }) // fails early
	}
}

// Test Distinct with many duplicates vs. few duplicates
func BenchmarkDistinctManyDuplicates(b *testing.B) {
	input := make([]int, 1000)
	for i := range input {
		input[i] = i % 10
	} // only 10 unique values
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := FromSlice(input)
		_ = Distinct(s).ToSlice()
	}
}

// Test a pipeline with Skip+Limit (sliding window)
func BenchmarkWindowedPipeline(b *testing.B) {
	input := make([]int, 10000)
	for i := range input {
		input[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := FromSlice(input)
		_ = Limit(Skip(s, 100), 50).ToSlice()
	}
}

// Test Sorted with already sorted data (the worst case for some algorithms)
func BenchmarkSortedAlreadySorted(b *testing.B) {
	input := make([]int, 1000)
	for i := range input {
		input[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := FromSlice(input)
		_ = Sorted(s).ToSlice()
	}
}

func BenchmarkSortedLarge(b *testing.B) {
	input := make([]int, 100000) // 100k elementos
	for i := range input {
		input[i] = rand.Intn(100000)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := FromSlice(input)
		_ = Sorted(s).ToSlice()
	}
}

func BenchmarkDistinctHighCardinality(b *testing.B) {
	input := make([]int, 10000)
	for i := range input {
		input[i] = i
	} // todos únicos
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := FromSlice(input)
		_ = Distinct(s).ToSlice()
	}
}

func BenchmarkNestedFlatMap(b *testing.B) {
	input := make([]int, 100)
	for i := range input {
		input[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := FromSlice(input)
		result := FlatMap(s, func(x int) []int {
			return FlatMap(FromSlice([]int{1, 2, 3}), func(y int) []int {
				return []int{x * y, x + y}
			}).ToSlice()
		}).ToSlice()
		_ = result
	}
}
