package queue

// initialCap defines the initial capacity for a deque when no specific capacity is provided.
const defaultInitialCap = 16

// Option is a functional option for configuring Deque and Queue.
type Option func(*options)

// options holds the options for Deque and Queue.
type options struct {
	initialCap int
}

func WithInitialCap(cap int) Option {
	return func(c *options) {
		if cap >= 1 {
			c.initialCap = cap
		} else {
			c.initialCap = defaultInitialCap
		}
	}
}

func applyDequeOptions(opts []Option) *options {
	c := &options{initialCap: defaultInitialCap}
	for _, o := range opts {
		o(c)
	}
	return c
}
