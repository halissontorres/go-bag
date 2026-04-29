package set

import (
	"slices"
	"testing"
)

// color is a tiny enum used to exercise EnumSet across multiple uint64 words.
type color int

func (c color) Index() int { return int(c) }

const (
	red color = iota
	green
	blue
	yellow
	purple
	farAway color = 130 // forces growth past the first 64-bit word
)

func TestEnumSet_AddContainsRemove(t *testing.T) {
	t.Parallel()

	es := NewEnumSet[color]()
	es.Add(red, green, blue)
	if got, want := es.Len(), 3; got != want {
		t.Fatalf("len=%d want %d", got, want)
	}
	if !es.Contains(red) || !es.Contains(green) || !es.Contains(blue) {
		t.Fatalf("missing element")
	}
	if es.Contains(yellow) {
		t.Fatalf("should not contain yellow")
	}
	es.Remove(green, yellow) // yellow is a no-op
	if es.Contains(green) {
		t.Fatalf("Remove(green) failed")
	}
	if got, want := es.Len(), 2; got != want {
		t.Fatalf("len after remove=%d want %d", got, want)
	}
}

func TestEnumSet_NegativeIndexSkipped(t *testing.T) {
	t.Parallel()

	es := NewEnumSet[color]()
	es.Add(color(-1))
	if es.Len() != 0 {
		t.Fatalf("negative index should be skipped, len=%d", es.Len())
	}
	if es.Contains(color(-1)) {
		t.Fatalf("Contains on negative index should be false")
	}
}

func TestEnumSet_Growth(t *testing.T) {
	t.Parallel()

	es := NewEnumSet[color]()
	es.Add(farAway) // index 130 -> requires growth to 3 words
	if !es.Contains(farAway) {
		t.Fatalf("Contains(farAway) should be true after growth")
	}
	if es.Len() != 1 {
		t.Fatalf("len=%d want 1", es.Len())
	}
}

func TestEnumSet_SetOperations(t *testing.T) {
	t.Parallel()

	a := NewEnumSet[color]()
	a.Add(red, green, blue)
	b := NewEnumSet[color]()
	b.Add(green, blue, yellow, farAway)

	if got := a.Union(b).Len(); got != 5 {
		t.Fatalf("Union len=%d want 5", got)
	}
	if got := a.Intersection(b).Len(); got != 2 {
		t.Fatalf("Intersection len=%d want 2", got)
	}
	if got := a.Difference(b).Len(); got != 1 {
		t.Fatalf("Difference len=%d want 1", got)
	}

	// Equal: same set must be equal.
	if !a.Equal(a) {
		t.Fatalf("set should equal itself")
	}
	clone := NewEnumSet[color]()
	clone.Add(red, green, blue)
	if !a.Equal(clone) {
		t.Fatalf("a should equal clone")
	}
	if a.Equal(b) {
		t.Fatalf("a and b should not be equal")
	}

	sub := NewEnumSet[color]()
	sub.Add(green, blue)
	if !sub.SubsetOf(a) {
		t.Fatalf("{green,blue} should be subset of a")
	}
	if a.SubsetOf(sub) {
		t.Fatalf("a should not be subset of sub")
	}
}

func TestEnumSet_Clear(t *testing.T) {
	t.Parallel()

	es := NewEnumSet[color]()
	es.Add(red, green, blue, farAway)
	es.Clear()
	if es.Len() != 0 {
		t.Fatalf("Len after Clear=%d want 0", es.Len())
	}
	if es.Contains(red) {
		t.Fatalf("Contains after Clear should be false")
	}
}

func TestEnumSet_ForEachVisitsAll(t *testing.T) {
	t.Parallel()

	es := NewEnumSet[color]()
	es.Add(red, blue, farAway)

	var got []int
	es.ForEach(func(idx int) color {
		got = append(got, idx)
		return color(idx)
	})
	slices.Sort(got)
	want := []int{int(red), int(blue), int(farAway)}
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Fatalf("ForEach indices=%v want %v", got, want)
	}
}

func TestEnumSet_String(t *testing.T) {
	t.Parallel()

	es := NewEnumSet[color]()
	if got, want := es.String(), "{}"; got != want {
		t.Fatalf("empty String=%q want %q", got, want)
	}
	es.Add(red)
	if got, want := es.String(), "{0}"; got != want {
		t.Fatalf("String=%q want %q", got, want)
	}
}

// ---------- Benchmarks ----------

func BenchmarkEnumSet_Add(b *testing.B) {
	b.ReportAllocs()
	es := NewEnumSet[color]()
	for b.Loop() {
		es.Add(red)
	}
}

func BenchmarkEnumSet_Contains(b *testing.B) {
	es := NewEnumSet[color]()
	es.Add(red, green, blue, farAway)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = es.Contains(farAway)
	}
}

func BenchmarkEnumSet_Union(b *testing.B) {
	a := NewEnumSet[color]()
	a.Add(red, green, blue)
	c := NewEnumSet[color]()
	c.Add(blue, yellow, purple, farAway)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = a.Union(c)
	}
}
