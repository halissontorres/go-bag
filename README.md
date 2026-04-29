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

- **Generic collections.** `LinkedList`, `Queue`, `Deque`, `Stack`, `Heap`, `Set`, `BTreeSet`, `BTreeMap`, and `DAG`, all parameterized on `any` or `comparable`/`Ordered` as appropriate.
- **Concurrency-ready.** Drop-in `SyncLinkedList`, `SyncQueue`, `SyncDeque`, `SyncStack`, `SyncHeap`, and `SyncSet` types for safe access from multiple goroutines.
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

### Heap

```go
import "github.com/halissontorres/go-bag/pkg/heap"

// Works with any cmp.Ordered type: int, float64, string, …
h := heap.New[int]()

h.Push(50)
h.Push(10)
h.Push(30)
h.Push(5)
h.Push(100)

// Peek returns the minimum without removing it.
min, _ := h.Peek() // 5

// Pop removes and returns elements in ascending order (Min-Heap).
v, _ := h.Pop() // 5
v, _ = h.Pop()  // 10

fmt.Println(h.Len()) // 3

// Works with strings too — ordering is lexicographic.
words := heap.New[string]()
words.Push("Zebra")
words.Push("Abacaxi")
words.Push("Banana")
first, _ := words.Pop() // "Abacaxi"

// Pop and Peek return an error when the heap is empty.
empty := heap.New[float64]()
_, err := empty.Pop() // err: "empty heap"

// Thread-safe variant — same API, protected by a sync.RWMutex.
sh := heap.NewSync[int]()
sh.Push(3)
sh.Push(1)
sh.Push(2)
min, _ = sh.Pop() // 1
```

> **Note:** `Heap` is a Min-Heap — `Pop` always returns the smallest element. Composite operations (e.g. peek-then-pop) are not atomic even on `SyncHeap`.

### Queue

```go
import "github.com/halissontorres/go-bag/pkg/queue"

q := queue.NewQueue[string]()
q.Enqueue("a")
q.Enqueue("b")
q.Enqueue("c")

front, _ := q.Peek()    // "a" — does not remove
val, _   := q.Dequeue() // "a"

fmt.Println(q.Len())      // 2
fmt.Println(q.IsEmpty())  // false
fmt.Println(q.Elements()) // [b c]

q.Clear()
fmt.Println(q.IsEmpty()) // true

// Thread-safe variant — same API, protected by a sync.RWMutex.
sq := queue.NewSyncQueue[int]()
sq.Enqueue(1)
sq.Enqueue(2)
v, _ := sq.Dequeue() // 1
```

### Deque

```go
import "github.com/halissontorres/go-bag/pkg/queue"

// Backed by a circular slice that doubles/halves automatically.
d := queue.NewDeque[int]()

d.PushBack(1)
d.PushBack(2)
d.PushFront(0) // [0 1 2]

front, _ := d.PeekFront() // 0
back,  _ := d.PeekBack()  // 2

v, _ := d.PopFront() // 0 → [1 2]
v, _ = d.PopBack()   // 2 → [1]

fmt.Println(d.Len())      // 1
fmt.Println(d.Elements()) // [1]
fmt.Println(d.String())   // [1]

// Custom initial capacity to avoid early reallocations.
large := queue.NewDequeWithCap[float64](256)

// Thread-safe variant.
sd := queue.NewSyncDeque[string]()
sd.PushBack("x")
sd.PushFront("y")
sd.PopFront() // "y"
```

### Stack

```go
import "github.com/halissontorres/go-bag/pkg/stack"

s := stack.NewStack[int]()
s.Push(10)
s.Push(20)
s.Push(30)

top, _ := s.Peek() // 30 — does not remove
v, _   := s.Pop()  // 30

fmt.Println(s.Len())      // 2
fmt.Println(s.IsEmpty())  // false
fmt.Println(s.Elements()) // [10 20] — bottom to top

s.Clear()
fmt.Println(s.IsEmpty()) // true

// Thread-safe variant — same API, protected by a sync.RWMutex.
ss := stack.NewSyncStack[string]()
ss.Push("hello")
ss.Push("world")
v2, _ := ss.Pop() // "world"
```

### Set

```go
import "github.com/halissontorres/go-bag/pkg/set"

a := set.NewSet[int]()
a.Add(1, 2, 3)

b := set.NewSet[int]()
b.Add(2, 3, 4)

fmt.Println(a.Contains(2)) // true
a.Remove(1)
fmt.Println(a.Len()) // 2

// Set algebra — each returns a new Set.
u    := set.Union(a, b)        // {2 3 4}
i    := set.Intersection(a, b) // {2 3}
diff := set.Difference(a, b)   // {2 3} \ {2 3 4} = {}

fmt.Println(set.SubsetOf(a, b)) // true
fmt.Println(set.Equal(a, b))    // false

// Functional helpers.
evens := a.Filter(func(x int) bool { return x%2 == 0 })
a.ForEach(func(x int) { fmt.Println(x) })
fmt.Println(a.Any(func(x int) bool { return x > 5 })) // false
fmt.Println(a.All(func(x int) bool { return x > 0 })) // true
clone := a.Clone()

// MapSet and ReduceSet accept different source and target types.
strs   := set.MapSet(a, func(x int) string { return fmt.Sprintf("%d", x) })
total  := set.ReduceSet(a, 0, func(acc, x int) int { return acc + x })

// Thread-safe variant — same API, protected by a sync.RWMutex.
ss := set.NewSyncSet[string]()
ss.Add("go", "bag")
fmt.Println(ss.Contains("go")) // true
```

### EnumSet

`EnumSet` is a bitmap-backed set for types that implement `Index() int`. All core operations run in O(words) — effectively O(1) for small enum families.

```go
import "github.com/halissontorres/go-bag/pkg/set"

// Any type that implements Index() int satisfies set.Enum.
type Role string

const (
    Admin  Role = "admin"
    Editor Role = "editor"
    Viewer Role = "viewer"
)

func (r Role) Index() int {
    switch r {
    case Admin:  return 0
    case Editor: return 1
    case Viewer: return 2
    default:     return -1
    }
}

es := set.NewEnumSet[Role]()
es.Add(Admin, Viewer)

fmt.Println(es.Contains(Admin))  // true
fmt.Println(es.Contains(Editor)) // false
fmt.Println(es.Len())            // 2

es.Remove(Viewer)
fmt.Println(es.Len()) // 1

// Bitmap operations — each returns a new EnumSet.
a := set.NewEnumSet[Role]()
a.Add(Admin, Editor)

b := set.NewEnumSet[Role]()
b.Add(Editor, Viewer)

u    := a.Union(b)        // {Admin Editor Viewer}
i    := a.Intersection(b) // {Editor}
diff := a.Difference(b)   // {Admin}

fmt.Println(a.SubsetOf(u)) // true
fmt.Println(a.Equal(b))    // false
fmt.Println(es.String())   // {0}
```

### BTreeSet

```go
import "github.com/halissontorres/go-bag/pkg/tree"

// Works with any cmp.Ordered type: int, float64, string, …
s := tree.NewBTreeSet[int]()
s.Add(5)
s.Add(3)
s.Add(8)
s.Add(1)

fmt.Println(s.Contains(3)) // true
fmt.Println(s.Len())       // 4

min, _ := s.Min() // 1
max, _ := s.Max() // 8

// In-order range query — closed interval [low, high].
slice := s.Range(3, 7) // [3 5]

// In-order traversal.
s.ForEach(func(v int) { fmt.Print(v, " ") }) // 1 3 5 8

// Forward iterator — advances one element at a time.
it := s.Iterator()
for v, ok := it.Next(); ok; v, ok = it.Next() {
    fmt.Println(v)
}

// Start iteration from a specific key.
it2 := s.IteratorFrom(3)
v, _ := it2.Next() // 3

s.Remove(3)
fmt.Println(s.Elements()) // [1 5 8]

// Custom minimum degree (controls tree fanout).
large := tree.NewBTreeSetWithDegree[string](16)
_ = large
```

### BTreeMap

```go
import "github.com/halissontorres/go-bag/pkg/tree"

m := tree.NewBTreeMap[string, int]()
m.Put("banana", 2)
m.Put("apple", 5)
m.Put("cherry", 1)

v, ok := m.Get("apple") // 5, true
fmt.Println(v, ok)

fmt.Println(m.Contains("banana")) // true
fmt.Println(m.Len())              // 3

// Keys and values are always returned in sorted key order.
fmt.Println(m.Keys())   // [apple banana cherry]
fmt.Println(m.Values()) // [5 2 1]

minK, minV, _ := m.Min() // "apple", 5
maxK, maxV, _ := m.Max() // "cherry", 1
_, _, _ = minK, minV, maxK

// Range returns KeyValuePair[K, V] — closed interval [low, high].
pairs := m.Range("apple", "banana")
for _, p := range pairs {
    fmt.Printf("%s=%d\n", p.Key, p.Value)
}

// ForEach iterates in sorted key order.
m.ForEach(func(k string, v int) {
    fmt.Printf("%s: %d\n", k, v)
})

m.Remove("banana")
fmt.Println(m.Len()) // 2
```

### DAG

```go
import "github.com/halissontorres/go-bag/pkg/graph"

g := graph.NewDAG[string]()

// AddVertex returns true when the vertex is new.
g.AddVertex("A")
g.AddVertex("B")
g.AddVertex("C")
g.AddVertex("D")

// AddEdge returns false if either vertex is missing, the edge already exists,
// or if adding it would form a cycle.
g.AddEdge("A", "B")
g.AddEdge("A", "C")
g.AddEdge("B", "D")
g.AddEdge("C", "D")

fmt.Println(g.HasVertex("A"))   // true
fmt.Println(g.HasEdge("A", "B")) // true

fmt.Println(g.OutDegree("A")) // 2
fmt.Println(g.InDegree("D"))  // 2

// Kahn's algorithm — always succeeds on a well-formed DAG.
order, ok := g.TopologicalSort() // ok=true, order e.g. [A B C D]
fmt.Println(ok, order)

// Reachability.
fmt.Println(g.HasPath("A", "D")) // true
fmt.Println(g.HasPath("D", "A")) // false

// Ancestors returns every vertex that can reach v.
anc := g.Ancestors("D") // {A B C}

// Descendants returns every vertex reachable from v.
desc := g.Descendants("A") // {B C D}
_ = anc
_ = desc

// Inspection.
fmt.Println(g.Vertices()) // [A B C D] (order not guaranteed)
fmt.Println(g.Edges())    // [[A B] [A C] [B D] [C D]] (order not guaranteed)

// Removal.
g.RemoveEdge("C", "D")
g.RemoveVertex("C")

fmt.Println(g.String()) // textual adjacency list
```

> **Note:** `AddEdge` performs a cycle check before insertion, so the graph is always acyclic. Attempting to add an edge that would close a cycle returns `false`.

### Enum Generator

The enum generator reads a `.go` file that declares string or int constants and emits a companion `*_enum.gen.go` file with the following helpers: `IsValid`, `Values`, `String`, `Parse<Type>`, `MarshalJSON`/`UnmarshalJSON`, `Value`/`Scan` (for `database/sql`), `Index`, and `Exhaustive`.

**1. Define your enum constants:**

```go
// status.go
package order

type Status string

const (
    StatusPending   Status = "pending"
    StatusApproved  Status = "approved"
    StatusRejected  Status = "rejected"
)
```

**2. Run the generator:**

```bash
# From within the package directory:
go run github.com/halissontorres/go-bag/cmd -dir . -type Status

# Or wire it to go generate:
//go:generate go run github.com/halissontorres/go-bag/cmd -dir . -type Status
go generate ./...
```

**3. Use the generated code:**

```go
s := StatusPending
fmt.Println(s.IsValid())  // true
fmt.Println(s.String())   // "pending"
fmt.Println(s.Index())    // 0

all := s.Values() // [StatusPending StatusApproved StatusRejected]

parsed, err := ParseStatus("approved") // StatusApproved, nil
_, err = ParseStatus("unknown")        // error: invalid Status: unknown

// Exhaustive panics if a new constant is added but the switch is not updated.
label := s.Exhaustive() // "pending"

// JSON and database/sql are handled automatically.
data, _ := json.Marshal(StatusApproved)     // "\"approved\""
db.Exec("INSERT INTO orders (status) ...", StatusApproved.Value)
```

The generated file is safe to commit and will not be overwritten unless you rerun the generator.

## Contributing

Issues and pull requests are welcome. Please run `go test ./...` and `go vet ./...` before submitting.

## License

Distributed under the MIT License. See [LICENSE](LICENSE) for details.
