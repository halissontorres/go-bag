package comparator

import gocmp "cmp"

// Comparator defines the ordering between two elements of the same type.
// Returns true if a should come before b.
type Comparator[T any] func(a, b T) bool

// Natural returns a Comparator that uses the natural ascending ordering
// for types that satisfy cmp.Ordered.
func Natural[T gocmp.Ordered]() Comparator[T] {
	return func(a, b T) bool { return a < b }
}

// Reverse returns a Comparator that uses the natural descending ordering
// for types that satisfy cmp.Ordered.
func Reverse[T gocmp.Ordered]() Comparator[T] {
	return func(a, b T) bool { return a > b }
}

// ByField returns a Comparator derived from a key extraction function.
// Useful for ordering structs by a specific field.
func ByField[T any, K gocmp.Ordered](key func(T) K) Comparator[T] {
	return func(a, b T) bool { return key(a) < key(b) }
}

// Then chains two Comparators: uses secondary when primary considers
// a and b equal (neither a < b nor b < a).
func (c Comparator[T]) Then(secondary Comparator[T]) Comparator[T] {
	return func(a, b T) bool {
		if c(a, b) {
			return true
		}
		if c(b, a) {
			return false
		}
		return secondary(a, b)
	}
}
