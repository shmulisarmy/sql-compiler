package unwrap

// Option represents a value that may or may not be present.
type Option[T any] struct {
	has_value bool
	Value     T
}

// Some creates an Option containing a value.
func Some[T any](v T) Option[T] {
	return Option[T]{has_value: true, Value: v}
}

// None creates an Option with no value.
func None[T any]() Option[T] {
	var zero T
	return Option[T]{has_value: false, Value: zero}
}

// IsSome returns true if the Option contains a value.
func (o Option[T]) IsSome() bool {
	return o.has_value
}

// IsNone returns true if the Option is empty.
func (o Option[T]) IsNone() bool {
	return !o.has_value
}

// Unwrap returns the value or panics if None.
func (o Option[T]) Expect(msg string) T {
	if !o.has_value {
		panic(msg)
	}
	return o.Value
}
func (o Option[T]) Unwrap() T {
	if !o.has_value {
		panic("called Unwrap on None")
	}
	return o.Value
}

// UnwrapOr returns the value or a default if None.
func (o Option[T]) UnwrapOr(defaultValue T) T {
	if o.has_value {
		return o.Value
	}
	return defaultValue
}
