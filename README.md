<p align="center">
  <img src="assets/go-bag.png" alt="go.bag logo" />
</p>

<p align="center">
  <strong>Gobag</strong> is a Go library that provides strongly-typed, generic data collections and a first-class enum toolkit, designed for safety, clarity, and performance.
</p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/halissontorres/go-bag"><img src="https://pkg.go.dev/badge/github.com/halissontorres/go-bag.svg" alt="Go Reference"></a>
  <a href="https://github.com/halissontorres/go-bag/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License"></a>
  <a href="https://go.dev/doc/devel/release"><img src="https://img.shields.io/github/go-mod/go-version/halissontorres/go-bag" alt="Go Version"></a>
</p>

---

## Overview

Gobag brings expressive, type-safe abstractions to everyday Go programming. It builds on Go 1.18+ generics to deliver a coherent set of collection types — lists, queues, deques, stacks, sets, ordered trees, and enum-aware bitmaps — together with a small code generator that promotes plain Go constants to first-class enums with parsing, JSON marshalling, and database integration.

Every collection ships in two flavors: a fast, single-threaded core type and a `Sync*` wrapper that adds a `sync.RWMutex` for safe concurrent use.

## Features

- **Generic collections.** `LinkedList`, `Queue`, `Deque`, `Stack`, `Set`, `BTreeSet`, `BTreeMap`, and `DAG`, all parameterized on `any` or `comparable`/`Ordered` as appropriate.
- **Concurrency-ready.** Drop-in `SyncLinkedList`, `SyncQueue`, `SyncDeque`, `SyncStack`, and `SyncSet` types for safe access from multiple goroutines.
- **Lazy streams.** `Stream[T]` with a pipeline-style API — `Filter`, `Map`, `FlatMap`, `Distinct`, `Sorted`, `Limit`, `Skip`, `Concat`, `Peek` — plus terminal operations `ToSlice`, `ForEach`, `Count`, `Any`, `All`, `Reduce`, and `FindFirst`. Streams are single-pass and not goroutine-safe.
- **Optional.** `Optional[T]` is a type-safe container that makes absent values explicit — no nil, no sentinel. Supports `Of`, `OfPtr`, `OrElse`, `OrElseGet`, `IfPresent`, `Filter`, `Map`, and `FlatMap`.
- **Bitmap-backed `EnumSet`.** O(1) membership, union, intersection, and difference for any type that exposes an `Index() int` method.
- **Ordered trees.** B-Tree-backed `BTreeSet` and `BTreeMap` with sorted iteration, in-order range queries, min/max lookup, and a forward iterator.
- **Directed acyclic graph.** `DAG[T]` with cycle-safe edge insertion, topological sort (Kahn's algorithm), reachability queries, and ancestor/descendant lookup.
- **Functional helpers.** `ForEach`, `Filter`, `Any`, `All`, `Clone`, `MapSet`, and `ReduceSet` over set types.
- **Enum code generator.** A `go generate`-friendly tool that emits `IsValid`, `Values`, `String`, `Parse`, `MarshalJSON`/`UnmarshalJSON`, `database/sql` `Valuer`/`Scanner`, `Index`, and an `Exhaustive` switch helper for any string- or int-backed constant family.
- **Zero external runtime dependencies.** Standard library only.

## Requirements

- Go 1.26 or newer (the module is built and tested on Go 1.26.2).

## Installation

```bash
go get github.com/halissontorres/go-bag
```

To use the enum generator as a CLI:

```bash
go install github.com/halissontorres/go-bag/cmd@latest
```

## Quick Start

Each collection lives in its own subpackage under `pkg/`, named after the data structure. Import the ones you need:

```go
import (
    "github.com/halissontorres/go-bag/pkg/list"
)
```

### Linked List

```go
import "github.com/halissontorres/go-bag/pkg/list"

// Create from scratch or from a slice.
l := list.NewLinkedList[int]()
l.AddLast(1)
l.AddLast(2)
l.AddFirst(0)

fmt.Println(l.Elements()) // [0 1 2]
fmt.Println(l.String())   // [0, 1, 2]

// Random-access (O(min(i, n-i)) bidirectional walk).
v, _ := l.Get(1)     // 1
l.InsertAt(1, 99)    // [0, 99, 1, 2]
l.RemoveAt(1)        // [0, 1, 2]

// Pop from either end.
first, _ := l.RemoveFirst() // 0
last,  _ := l.RemoveLast()  // 2

// Alias methods (Stack / Queue style).
l.PushFront(10); l.PushBack(20); l.Append(30)
l.PopFront(); l.PopBack()

// Forward and reverse iteration.
it := l.Iter()
for v, ok := it.Next(); ok; v, ok = it.Next() {
    fmt.Println(v)
}

rit := l.ReverseIter()
for v, ok := rit.Next(); ok; v, ok = rit.Next() {
    fmt.Println(v)
}

// Thread-safe variant — same API, protected by a sync.RWMutex.
sl := list.NewSyncLinkedList[int]()
sl.AddLast(42)
fmt.Println(sl.String()) // [42]
```

### Stream

```go
import "github.com/halissontorres/go-bag/pkg/stream"

// Build a pipeline: double every number, keep only those > 10, take the first 3.
result := stream.Limit(
    stream.Filter(
        stream.Map(
            stream.FromSlice([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}),
            func(x int) int { return x * 2 },
        ),
        func(x int) bool { return x > 10 },
    ),
    3,
).ToSlice()

fmt.Println(result) // [12 14 16]
```

Other sources and operations:

```go
// From a channel
ch := make(chan string, 3)
ch <- "a"; ch <- "b"; ch <- "c"
close(ch)
words := stream.FromChannel(ch).ToSlice() // ["a" "b" "c"]

// Deduplicate and sort
unique := stream.Sorted(
    stream.Distinct(stream.FromSlice([]int{3, 1, 2, 1, 3})),
).ToSlice() // [1 2 3]

// Aggregate
sum := stream.FromSlice([]int{1, 2, 3, 4}).
    Reduce(0, func(a, b int) int { return a + b }) // 10

// Short-circuit checks
hasEven := stream.FromSlice([]int{1, 2, 3}).
    Any(func(x int) bool { return x%2 == 0 }) // true

allPositive := stream.FromSlice([]int{1, 2, 3}).
    All(func(x int) bool { return x > 0 }) // true

// FlatMap
pairs := stream.FlatMap(
    stream.FromSlice([]int{1, 2}),
    func(x int) []int { return []int{x, x * 10} },
).ToSlice() // [1 10 2 20]

// FindFirst returns an Optional — present if a match is found, empty otherwise.
result := stream.FindFirst(
    stream.FromSlice([]int{1, 3, 4, 6}),
    func(x int) bool { return x%2 == 0 },
)
if v, ok := result.Get(); ok {
    fmt.Println(v) // 4
}
```

> **Note:** Streams are single-pass. Once consumed, re-calling any terminal operation returns an empty result.

### Optional

```go
import "github.com/halissontorres/go-bag/pkg/opt"

// Create
present := opt.Of(42)
empty   := opt.Empty[int]()

n := 99
fromPtr := opt.OfPtr(&n)      // present
fromNil := opt.OfPtr[int](nil) // empty

// Unwrap
v, ok := present.Get()          // 42, true
present.IfPresent(func(v int) { fmt.Println(v) }) // 42

// Fallback
empty.OrElse(0)                  // 0
empty.OrElseGet(func() int { return computeDefault() })

// Transform
doubled := opt.Map(present, func(x int) int { return x * 2 }) // Optional[84]
nested  := opt.FlatMap(present, func(x int) opt.Optional[string] {
    return opt.Of(fmt.Sprintf("val=%d", x))
}) // Optional["val=42"]

// Filter
even := present.Filter(func(x int) bool { return x%2 == 0 }) // Optional[42]
odd  := opt.Of(3).Filter(func(x int) bool { return x%2 == 0 }) // empty

fmt.Println(present) // Optional[42]
fmt.Println(empty)   // Optional[empty]
```

## Contributing

Issues and pull requests are welcome. Please run `go test ./...` and `go vet ./...` before submitting.

## License

Distributed under the MIT License. See [LICENSE](LICENSE) for details.
