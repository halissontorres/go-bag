package opt

import (
	"fmt"
)

// Optional is a container that may or may not hold a non-nil value.
// It provides a type-safe alternative to using nil pointers.
type Optional[T any] struct {
	value   T
	present bool
}

// Of creates an Optional with a present value.
func Of[T any](value T) Optional[T] {
	return Optional[T]{value: value, present: true}
}

// Empty returns an empty Optional.
func Empty[T any]() Optional[T] {
	return Optional[T]{present: false}
}

// OfNullable creates an Optional from a value that may be a zero value.
// In Go every type has a zero value, so this behaves identically to Of for
// non-pointer types. For pointer types, prefer OfPtr to distinguish nil.
func OfNullable[T any](value T) Optional[T] {
	return Of(value)
}

// OfPtr creates an Optional from a pointer. Returns an empty Optional if the
// pointer is nil; otherwise returns a present Optional with the dereferenced value.
func OfPtr[T any](ptr *T) Optional[T] {
	if ptr == nil {
		return Empty[T]()
	}
	return Of(*ptr)
}

// IsPresent returns true if the Optional contains a value.
func (o Optional[T]) IsPresent() bool {
	return o.present
}

// Get returns the value and true if present; otherwise the zero value of T and false.
func (o Optional[T]) Get() (T, bool) {
	return o.value, o.present
}

// OrElse returns the value if present, otherwise returns other.
func (o Optional[T]) OrElse(other T) T {
	if o.present {
		return o.value
	}
	return other
}

// OrElseGet returns the value if present, otherwise calls supplier and returns its result.
func (o Optional[T]) OrElseGet(supplier func() T) T {
	if o.present {
		return o.value
	}
	return supplier()
}

// OrElseFunc calls consumer with the value if present; otherwise calls onEmpty.
func (o Optional[T]) OrElseFunc(consumer func(T), onEmpty func()) {
	if o.present {
		consumer(o.value)
	} else {
		onEmpty()
	}
}

// IfPresent executes the action if a value is present.
func (o Optional[T]) IfPresent(action func(T)) {
	if o.present {
		action(o.value)
	}
}

// Map applies f to the value if present and returns an Optional of the result.
func Map[T, U any](o Optional[T], f func(T) U) Optional[U] {
	if !o.present {
		return Empty[U]()
	}
	return Of(f(o.value))
}

// FlatMap applies f, which itself returns an Optional, avoiding nested Optionals.
func FlatMap[T, U any](o Optional[T], f func(T) Optional[U]) Optional[U] {
	if !o.present {
		return Empty[U]()
	}
	return f(o.value)
}

// Filter returns the Optional unchanged if present and the predicate holds; otherwise empty.
func (o Optional[T]) Filter(pred func(T) bool) Optional[T] {
	if o.present && pred(o.value) {
		return o
	}
	return Empty[T]()
}

// String returns a human-readable representation of the Optional.
func (o Optional[T]) String() string {
	if o.present {
		return fmt.Sprintf("Optional[%v]", o.value)
	}
	return "Optional[empty]"
}
