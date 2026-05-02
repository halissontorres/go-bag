package graph

// dagConfig holds construction-time options for a DAG.
type dagConfig struct {
	initialVertices int
	skipCycleCheck  bool
}

// DAGOption configures a DAG at creation time.
type DAGOption[T comparable] func(*dagConfig)

// WithInitialVertices hints the expected number of vertices, reducing
// internal map reallocations for large graphs.
func WithInitialVertices[T comparable](n int) DAGOption[T] {
	return func(c *dagConfig) {
		if n > 0 {
			c.initialVertices = n
		}
	}
}

// WithSkipCycleCheck disables the DFS cycle check on AddEdge.
// Use only when the caller guarantees that no cycles will be introduced —
// e.g. when building a DAG from a pre-validated source.
// Adding a back-edge with this option active silently corrupts the DAG.
func WithSkipCycleCheck[T comparable]() DAGOption[T] {
	return func(c *dagConfig) {
		c.skipCycleCheck = true
	}
}
