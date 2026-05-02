package stack

// stackConfig holds construction-time options for a Stack.
type option struct {
	initialCap int
}

// StackOption configures a Stack or SyncStack at creation time.
type Option func(*option)

// WithInitialCap hints the expected number of elements, reducing
// internal slice reallocations for large stacks.
func WithInitialCap(cap int) Option {
	return func(c *option) {
		if cap >= 1 {
			c.initialCap = cap
		}
	}
}

func applyStackOptions(opts []Option) *option {
	c := &option{}
	for _, o := range opts {
		o(c)
	}
	return c
}
