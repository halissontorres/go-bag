package stream

// options holds options for stream operations.
type options struct {
	initialCap int
}

// Option configures stream terminal or intermediate operations.
type Option func(*options)

// WithInitialCap hints the expected number of elements for operations
// that pre-allocate internal buffers (ToSlice, Distinct, Sorted).
// Values less than 1 are ignored.
func WithInitialCap(cap int) Option {
	return func(c *options) {
		if cap >= 1 {
			c.initialCap = cap
		}
	}
}

const defaultInitialCap = 256

func applyStreamOptions(opts []Option) options {
	c := options{initialCap: defaultInitialCap}
	for _, o := range opts {
		o(&c)
	}
	return c
}
