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

Each collection lives in its own subpackage, named after the directory. Import the ones you need:

```go
import (
    "github.com/halissontorres/go-bag/bag/lists"
)
```

### Linked List

```go
l := lists.NewLinkedList[int]()
l.AddLast(1)
l.AddLast(2)
l.AddFirst(0)

fmt.Println(l.Elements()) // [0 1 2]
v, _ := l.RemoveFirst()
fmt.Println(v)            // 0
```

## Contributing

Issues and pull requests are welcome. Please run `go test ./...` and `go vet ./...` before submitting.

## License

Distributed under the MIT License. See [LICENSE](LICENSE) for details.
