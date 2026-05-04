package tree

import (
	"slices"
	"testing"

	"github.com/halissontorres/go-bag/pkg/comparator"
)

func TestBTreeMap_PutGetRemove(t *testing.T) {
	t.Parallel()

	m := NewBTreeMap[string, int](comparator.Natural[string]())
	if got, _ := m.Get("missing"); got != 0 {
		t.Fatalf("Get on empty should return zero value, got %d", got)
	}
	if _, ok := m.Get("missing"); ok {
		t.Fatalf("Get on empty should return false")
	}
	m.Put("a", 1)
	m.Put("b", 2)
	m.Put("c", 3)
	m.Put("a", 10) // update
	if got, want := m.Len(), 3; got != want {
		t.Fatalf("len=%d want %d", got, want)
	}
	if v, ok := m.Get("a"); !ok || v != 10 {
		t.Fatalf("Get(a)=%d,%v want 10,true", v, ok)
	}
	if !m.Contains("b") {
		t.Fatalf("Contains(b) should be true")
	}
	if !m.Remove("b") {
		t.Fatalf("Remove(b) should be true")
	}
	if m.Remove("b") {
		t.Fatalf("Remove(b) twice should be false")
	}
	if got, want := m.Len(), 2; got != want {
		t.Fatalf("len after remove=%d want %d", got, want)
	}
}

func TestBTreeMap_KeysValuesSorted(t *testing.T) {
	t.Parallel()

	m := NewBTreeMap[int, string](comparator.Natural[int]())
	pairs := []KeyValuePair[int, string]{
		{3, "three"}, {1, "one"}, {4, "four"}, {1, "one-dup"}, {5, "five"}, {9, "nine"}, {2, "two"}, {6, "six"},
	}
	for _, p := range pairs {
		m.Put(p.Key, p.Value)
	}

	keys := m.Keys()
	if !slices.IsSorted(keys) {
		t.Fatalf("Keys not sorted: %v", keys)
	}
	wantKeys := []int{1, 2, 3, 4, 5, 6, 9}
	if !slices.Equal(keys, wantKeys) {
		t.Fatalf("Keys=%v want %v", keys, wantKeys)
	}

	values := m.Values()
	wantValues := []string{"one-dup", "two", "three", "four", "five", "six", "nine"}
	if !slices.Equal(values, wantValues) {
		t.Fatalf("Values=%v want %v", values, wantValues)
	}

	var seenKeys []int
	m.ForEach(func(k int, _ string) { seenKeys = append(seenKeys, k) })
	if !slices.Equal(seenKeys, wantKeys) {
		t.Fatalf("ForEach keys=%v want %v", seenKeys, wantKeys)
	}
}

func TestBTreeMap_MinMaxRange(t *testing.T) {
	t.Parallel()

	m := NewBTreeMap[int, string](comparator.Natural[int]())
	for _, k := range []int{10, 20, 5, 15, 25, 30, 8} {
		m.Put(k, "v")
	}
	if k, _, ok := m.Min(); !ok || k != 5 {
		t.Fatalf("Min=%d,%v want 5,true", k, ok)
	}
	if k, _, ok := m.Max(); !ok || k != 30 {
		t.Fatalf("Max=%d,%v want 30,true", k, ok)
	}

	pairs := m.Range(8, 25)
	gotKeys := make([]int, len(pairs))
	for i, p := range pairs {
		gotKeys[i] = p.Key
	}
	if want := []int{8, 10, 15, 20, 25}; !slices.Equal(gotKeys, want) {
		t.Fatalf("Range keys=%v want %v", gotKeys, want)
	}

	empty := NewBTreeMap[int, string](comparator.Natural[int]())
	if _, _, ok := empty.Min(); ok {
		t.Fatalf("Min on empty should be false")
	}
	if _, _, ok := empty.Max(); ok {
		t.Fatalf("Max on empty should be false")
	}
}

func TestBTreeMap_WithMinDegree(t *testing.T) {
	t.Parallel()

	m1 := NewBTreeMap[int, string](comparator.Natural[int]())
	m2 := NewBTreeMap[int, string](comparator.Natural[int](), WithMinDegree(2))
	for i := 0; i < 50; i++ {
		m1.Put(i, "v")
		m2.Put(i, "v")
	}
	if !slices.Equal(m1.Keys(), m2.Keys()) {
		t.Fatal("default degree and WithMinDegree(2) should produce identical key order")
	}

	m3 := NewBTreeMap[int, string](comparator.Natural[int](), WithMinDegree(1))
	for i := 0; i < 30; i++ {
		m3.Put(i, "x")
	}
	if m3.Len() != 30 {
		t.Fatalf("len=%d want 30", m3.Len())
	}

	m4 := NewBTreeMap[int, int](comparator.Natural[int](), WithMinDegree(4))
	const N = 200
	for i := N - 1; i >= 0; i-- {
		m4.Put(i, i*2)
	}
	if !slices.IsSorted(m4.Keys()) {
		t.Fatal("Keys not sorted for WithMinDegree(4)")
	}
	if m4.Len() != N {
		t.Fatalf("len=%d want %d", m4.Len(), N)
	}
}

type item struct {
	ID   int
	Name string
}

func TestBTreeMap_StructKeyByField(t *testing.T) {
	byID := comparator.ByField(func(i item) int { return i.ID })
	m := NewBTreeMap[item, string](byID)

	m.Put(item{ID: 3, Name: "C"}, "third")
	m.Put(item{ID: 1, Name: "A"}, "first")
	m.Put(item{ID: 2, Name: "B"}, "second")

	keys := m.Keys()
	if keys[0].ID != 1 || keys[1].ID != 2 || keys[2].ID != 3 {
		t.Fatalf("Keys not sorted by ID: %v", keys)
	}

	v, ok := m.Get(item{ID: 1})
	if !ok || v != "first" {
		t.Fatalf("Get({1})=%q,%v want first,true", v, ok)
	}
}

func TestBTreeMap_StructKeyMultiCriterion(t *testing.T) {
	byName := comparator.ByField(func(i item) string { return i.Name })
	byID := comparator.ByField(func(i item) int { return i.ID })
	cmp := byName.Then(byID)

	m := NewBTreeMap[item, string](cmp)

	m.Put(item{ID: 2, Name: "Alice"}, "a2")
	m.Put(item{ID: 1, Name: "Alice"}, "a1")
	m.Put(item{ID: 3, Name: "Bob"}, "b3")

	keys := m.Keys()
	if keys[0].Name != "Alice" || keys[0].ID != 1 {
		t.Fatalf("first should be Alice/1, got %v", keys[0])
	}
	if keys[1].Name != "Alice" || keys[1].ID != 2 {
		t.Fatalf("second should be Alice/2, got %v", keys[1])
	}
	if keys[2].Name != "Bob" || keys[2].ID != 3 {
		t.Fatalf("third should be Bob/3, got %v", keys[2])
	}
}

func BenchmarkBTreeMap_Put(b *testing.B) {
	b.ReportAllocs()
	m := NewBTreeMap[int, int](comparator.Natural[int]())
	i := 0
	for b.Loop() {
		m.Put(i, i)
		i++
	}
}

func BenchmarkBTreeMap_Get(b *testing.B) {
	const N = 10_000
	m := NewBTreeMap[int, int](comparator.Natural[int]())
	for i := 0; i < N; i++ {
		m.Put(i, i)
	}
	b.ReportAllocs()
	b.ResetTimer()
	i := 0
	for b.Loop() {
		_, _ = m.Get(i % N)
		i++
	}
}
