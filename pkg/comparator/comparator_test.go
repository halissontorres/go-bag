package comparator_test

import (
	"testing"

	"github.com/halissontorres/go-bag/pkg/comparator"
)

type person struct {
	Name string
	Age  int
}

func TestNaturalInt(t *testing.T) {
	cmp := comparator.Natural[int]()
	if !cmp(1, 2) {
		t.Error("Natural[int]: 1 should come before 2")
	}
	if cmp(2, 1) {
		t.Error("Natural[int]: 2 should not come before 1")
	}
	if cmp(1, 1) {
		t.Error("Natural[int]: 1 should not come before 1")
	}
}

func TestNaturalString(t *testing.T) {
	cmp := comparator.Natural[string]()
	if !cmp("a", "b") {
		t.Error(`Natural[string]: "a" should come before "b"`)
	}
	if cmp("b", "a") {
		t.Error(`Natural[string]: "b" should not come before "a"`)
	}
	if cmp("a", "a") {
		t.Error(`Natural[string]: "a" should not come before "a"`)
	}
}

func TestReverseInt(t *testing.T) {
	cmp := comparator.Reverse[int]()
	if !cmp(2, 1) {
		t.Error("Reverse[int]: 2 should come before 1 (descending)")
	}
	if cmp(1, 2) {
		t.Error("Reverse[int]: 1 should not come before 2 (descending)")
	}
	if cmp(1, 1) {
		t.Error("Reverse[int]: 1 should not come before 1")
	}
}

func TestByFieldName(t *testing.T) {
	cmp := comparator.ByField(func(p person) string { return p.Name })

	alice := person{Name: "Alice", Age: 30}
	bob := person{Name: "Bob", Age: 25}

	if !cmp(alice, bob) {
		t.Error("ByField[Name]: Alice should come before Bob")
	}
	if cmp(bob, alice) {
		t.Error("ByField[Name]: Bob should not come before Alice")
	}
}

func TestByFieldAge(t *testing.T) {
	cmp := comparator.ByField(func(p person) int { return p.Age })

	young := person{Name: "Alice", Age: 25}
	old := person{Name: "Bob", Age: 30}

	if !cmp(young, old) {
		t.Error("ByField[Age]: 25 should come before 30")
	}
	if cmp(old, young) {
		t.Error("ByField[Age]: 30 should not come before 25")
	}
}

func TestThenPrimaryDecides(t *testing.T) {
	byAge := comparator.ByField(func(p person) int { return p.Age })
	byName := comparator.ByField(func(p person) string { return p.Name })
	combined := byAge.Then(byName)

	young := person{Name: "Zoe", Age: 25}
	old := person{Name: "Alice", Age: 30}

	if !combined(young, old) {
		t.Error("Then: primary (age) should decide when ages differ; 25 < 30")
	}
	if combined(old, young) {
		t.Error("Then: 30 should not come before 25 when primary is age")
	}
}

func TestThenFallbackToSecondary(t *testing.T) {
	byAge := comparator.ByField(func(p person) int { return p.Age })
	byName := comparator.ByField(func(p person) string { return p.Name })
	combined := byAge.Then(byName)

	alice := person{Name: "Alice", Age: 30}
	bob := person{Name: "Bob", Age: 30}

	if !combined(alice, bob) {
		t.Error("Then: secondary (name) should decide when ages are equal; Alice < Bob")
	}
	if combined(bob, alice) {
		t.Error("Then: Bob should not come before Alice when ages are equal")
	}
}

func TestThenAllEqual(t *testing.T) {
	byAge := comparator.ByField(func(p person) int { return p.Age })
	byName := comparator.ByField(func(p person) string { return p.Name })
	combined := byAge.Then(byName)

	a1 := person{Name: "Alice", Age: 30}
	a2 := person{Name: "Alice", Age: 30}

	if combined(a1, a2) {
		t.Error("Then: equal elements should not be ordered either way")
	}
	if combined(a2, a1) {
		t.Error("Then: equal elements should not be ordered either way")
	}
}
