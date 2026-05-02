package tree

// options holds construction-time options for a BTreeSet or BTreeMap.
type options struct {
	minDegree int
}

// Option configures a BTreeSet or BTreeMap at creation time.
type Option func(*options)

// WithMinDegree sets the minimum degree t of the underlying B-Tree (t ≥ 2).
// Each internal node holds between t-1 and 2t-1 keys and between t and 2t children.
// Higher values reduce tree height at the cost of larger node allocations.
// Values below 2 are ignored and the default (2) is used instead.
func WithMinDegree(d int) Option {
	return func(c *options) {
		if d >= 2 {
			c.minDegree = d
		}
	}
}

func applyTreeOptions(opts []Option) options {
	c := options{minDegree: defaultMinDegree}
	for _, o := range opts {
		o(&c)
	}
	return c
}
