package sketch

// CountMinOption configures a CountMin at creation time.
type CountMinOption[T any] func(*countMin[T])

// HyperLogLogOption configures a HyperLogLog at creation time.
type HyperLogLogOption[T any] func(*hyperLogLog[T])
