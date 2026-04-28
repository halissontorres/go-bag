package opts

import (
	"testing"
)

func TestOf_IsPresent(t *testing.T) {
	o := Of(42)
	if !o.IsPresent() {
		t.Fatal("Of() should be present")
	}
	v, ok := o.Get()
	if !ok || v != 42 {
		t.Fatalf("Get()=%d,%v want 42,true", v, ok)
	}
}

func TestEmpty_IsNotPresent(t *testing.T) {
	o := Empty[int]()
	if o.IsPresent() {
		t.Fatal("Empty() should not be present")
	}
	_, ok := o.Get()
	if ok {
		t.Fatal("Get() on empty should return false")
	}
}

func TestOfPtr_Nil(t *testing.T) {
	o := OfPtr[int](nil)
	if o.IsPresent() {
		t.Fatal("OfPtr(nil) should not be present")
	}
}

func TestOfPtr_NonNil(t *testing.T) {
	n := 99
	o := OfPtr(&n)
	v, ok := o.Get()
	if !ok || v != 99 {
		t.Fatalf("OfPtr(&n)=%d,%v want 99,true", v, ok)
	}
}

func TestOfNullable(t *testing.T) {
	o := OfNullable("hello")
	if !o.IsPresent() {
		t.Fatal("OfNullable should be present for non-zero value")
	}
}

func TestOrElse(t *testing.T) {
	if got := Of(1).OrElse(99); got != 1 {
		t.Fatalf("OrElse on present=%d want 1", got)
	}
	if got := Empty[int]().OrElse(99); got != 99 {
		t.Fatalf("OrElse on empty=%d want 99", got)
	}
}

func TestOrElseGet(t *testing.T) {
	supplier := func() int { return 42 }
	if got := Of(7).OrElseGet(supplier); got != 7 {
		t.Fatalf("OrElseGet on present=%d want 7", got)
	}
	if got := Empty[int]().OrElseGet(supplier); got != 42 {
		t.Fatalf("OrElseGet on empty=%d want 42", got)
	}
}

func TestOrElseFunc(t *testing.T) {
	var consumed int
	var emptyCalled bool

	Of(5).OrElseFunc(func(v int) { consumed = v }, func() { emptyCalled = true })
	if consumed != 5 || emptyCalled {
		t.Fatalf("OrElseFunc(present): consumed=%d emptyCalled=%v", consumed, emptyCalled)
	}

	consumed = 0
	Empty[int]().OrElseFunc(func(v int) { consumed = v }, func() { emptyCalled = true })
	if consumed != 0 || !emptyCalled {
		t.Fatalf("OrElseFunc(empty): consumed=%d emptyCalled=%v", consumed, emptyCalled)
	}
}

func TestIfPresent(t *testing.T) {
	called := false
	Of(1).IfPresent(func(v int) { called = true })
	if !called {
		t.Fatal("IfPresent should call action when present")
	}

	called = false
	Empty[int]().IfPresent(func(v int) { called = true })
	if called {
		t.Fatal("IfPresent should not call action when empty")
	}
}

func TestFilter(t *testing.T) {
	even := func(x int) bool { return x%2 == 0 }

	if !Of(4).Filter(even).IsPresent() {
		t.Fatal("Filter(4, even) should be present")
	}
	if Of(3).Filter(even).IsPresent() {
		t.Fatal("Filter(3, even) should be empty")
	}
	if Empty[int]().Filter(even).IsPresent() {
		t.Fatal("Filter on empty should be empty")
	}
}

func TestMap(t *testing.T) {
	result := Map(Of(3), func(x int) string {
		return "val"
	})
	if !result.IsPresent() {
		t.Fatal("Map on present should be present")
	}
	v, _ := result.Get()
	if v != "val" {
		t.Fatalf("Map result=%q want \"val\"", v)
	}

	empty := Map(Empty[int](), func(x int) string { return "x" })
	if empty.IsPresent() {
		t.Fatal("Map on empty should be empty")
	}
}

func TestFlatMap(t *testing.T) {
	double := func(x int) Optional[int] { return Of(x * 2) }

	result := FlatMap(Of(5), double)
	v, ok := result.Get()
	if !ok || v != 10 {
		t.Fatalf("FlatMap(5, double)=%d,%v want 10,true", v, ok)
	}

	empty := FlatMap(Empty[int](), double)
	if empty.IsPresent() {
		t.Fatal("FlatMap on empty should be empty")
	}
}

func TestString(t *testing.T) {
	if got, want := Of(7).String(), "Optional[7]"; got != want {
		t.Fatalf("String()=%q want %q", got, want)
	}
	if got, want := Empty[int]().String(), "Optional[empty]"; got != want {
		t.Fatalf("String()=%q want %q", got, want)
	}
}
