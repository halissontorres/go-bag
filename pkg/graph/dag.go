package graph

import (
	"fmt"
	"strings"

	"github.com/halissontorres/go-bag/pkg/set"
)

// DAG is a directed acyclic graph. Vertices must be comparable.
//
// Adjacency is stored as a slice per source vertex. This trades O(1)
// edge lookup for O(out-degree) lookup, but eliminates the per-vertex
// inner map allocation and is faster on sparse graphs (the common case).
type DAG[T comparable] struct {
	vertices       map[T]struct{}
	adj            map[T][]T
	inDegree       map[T]int
	skipCycleCheck bool
}

// NewDAG creates an empty DAG.
func NewDAG[T comparable](opts ...DAGOption[T]) *DAG[T] {
	c := &dagConfig{}
	for _, o := range opts {
		o(c)
	}
	return &DAG[T]{
		vertices:       make(map[T]struct{}, c.initialVertices),
		adj:            make(map[T][]T, c.initialVertices),
		inDegree:       make(map[T]int, c.initialVertices),
		skipCycleCheck: c.skipCycleCheck,
	}
}

// AddVertex inserts a vertex if it does not already exist.
// Returns true when the vertex was added.
func (g *DAG[T]) AddVertex(v T) bool {
	if _, ok := g.vertices[v]; ok {
		return false
	}
	g.vertices[v] = struct{}{}
	g.inDegree[v] = 0
	return true
}

// AddEdge inserts a directed edge from -> to.
// Returns false if either vertex is missing, the edge already exists,
// or — unless WithSkipCycleCheck was set — adding it would introduce a cycle.
func (g *DAG[T]) AddEdge(from, to T) bool {
	if _, ok := g.vertices[from]; !ok {
		return false
	}
	if _, ok := g.vertices[to]; !ok {
		return false
	}
	if from == to {
		return false
	}
	for _, n := range g.adj[from] {
		if n == to {
			return false
		}
	}

	if !g.skipCycleCheck && len(g.adj[to]) > 0 && g.hasPath(to, from) {
		return false
	}

	g.adj[from] = append(g.adj[from], to)
	g.inDegree[to]++
	return true
}

// RemoveVertex removes a vertex and every edge incident to it.
func (g *DAG[T]) RemoveVertex(v T) bool {
	if _, ok := g.vertices[v]; !ok {
		return false
	}

	for _, neighbor := range g.adj[v] {
		g.inDegree[neighbor]--
	}
	for src, dests := range g.adj {
		if src == v {
			continue
		}
		for i, n := range dests {
			if n == v {
				g.adj[src] = append(dests[:i], dests[i+1:]...)
				break
			}
		}
	}

	delete(g.vertices, v)
	delete(g.adj, v)
	delete(g.inDegree, v)
	return true
}

// RemoveEdge removes the edge from -> to.
func (g *DAG[T]) RemoveEdge(from, to T) bool {
	dests := g.adj[from]
	for i, n := range dests {
		if n == to {
			g.adj[from] = append(dests[:i], dests[i+1:]...)
			g.inDegree[to]--
			return true
		}
	}
	return false
}

// HasVertex reports whether the vertex exists.
func (g *DAG[T]) HasVertex(v T) bool {
	_, ok := g.vertices[v]
	return ok
}

// HasEdge reports whether the edge from -> to exists.
func (g *DAG[T]) HasEdge(from, to T) bool {
	for _, n := range g.adj[from] {
		if n == to {
			return true
		}
	}
	return false
}

// OutDegree returns the out-degree of v.
func (g *DAG[T]) OutDegree(v T) int {
	return len(g.adj[v])
}

// InDegree returns the in-degree of v.
func (g *DAG[T]) InDegree(v T) int {
	return g.inDegree[v]
}

// TopologicalSort returns a topological ordering using Kahn's algorithm.
// The boolean is true on success; false means the graph contains a cycle.
func (g *DAG[T]) TopologicalSort() ([]T, bool) {
	n := len(g.vertices)
	inDeg := make(map[T]int, n)
	queue := make([]T, 0, n)
	for v, deg := range g.inDegree {
		inDeg[v] = deg
		if deg == 0 {
			queue = append(queue, v)
		}
	}

	order := make([]T, 0, n)
	for i := 0; i < len(queue); i++ {
		v := queue[i]
		order = append(order, v)
		for _, neighbor := range g.adj[v] {
			inDeg[neighbor]--
			if inDeg[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if len(order) != n {
		return nil, false
	}
	return order, true
}

// Vertices returns every vertex in the graph.
func (g *DAG[T]) Vertices() []T {
	result := make([]T, 0, len(g.vertices))
	for v := range g.vertices {
		result = append(result, v)
	}
	return result
}

// Edges returns every edge as a [from, to] pair.
func (g *DAG[T]) Edges() [][2]T {
	edgeCount := 0
	for _, dests := range g.adj {
		edgeCount += len(dests)
	}
	edges := make([][2]T, 0, edgeCount)
	for from, dests := range g.adj {
		for _, to := range dests {
			edges = append(edges, [2]T{from, to})
		}
	}
	return edges
}

// HasPath reports whether a path from -> to exists.
func (g *DAG[T]) HasPath(from, to T) bool {
	return g.hasPath(from, to)
}

// hasPath performs an iterative DFS.
func (g *DAG[T]) hasPath(from, to T) bool {
	visited := make(map[T]struct{})
	stack := []T{from}
	for len(stack) > 0 {
		n := len(stack) - 1
		current := stack[n]
		stack = stack[:n]
		if current == to {
			return true
		}
		if _, ok := visited[current]; ok {
			continue
		}
		visited[current] = struct{}{}
		stack = append(stack, g.adj[current]...)
	}
	return false
}

// Ancestors returns every vertex that can reach v (excluding v itself).
func (g *DAG[T]) Ancestors(v T) *set.Set[T] {
	anc := set.NewSet[T]()
	reverseAdj := g.reverseAdj()
	n := len(g.vertices)
	visited := make(map[T]struct{}, n)
	visited[v] = struct{}{}
	queue := make([]T, 1, n)
	queue[0] = v
	for i := 0; i < len(queue); i++ {
		for _, neighbor := range reverseAdj[queue[i]] {
			if _, seen := visited[neighbor]; seen {
				continue
			}
			visited[neighbor] = struct{}{}
			anc.Add(neighbor)
			queue = append(queue, neighbor)
		}
	}
	return anc
}

// Descendants returns every vertex reachable from v (excluding v itself).
func (g *DAG[T]) Descendants(v T) *set.Set[T] {
	desc := set.NewSet[T]()
	n := len(g.vertices)
	visited := make(map[T]struct{}, n)
	visited[v] = struct{}{}
	queue := make([]T, 1, n)
	queue[0] = v
	for i := 0; i < len(queue); i++ {
		for _, neighbor := range g.adj[queue[i]] {
			if _, seen := visited[neighbor]; seen {
				continue
			}
			visited[neighbor] = struct{}{}
			desc.Add(neighbor)
			queue = append(queue, neighbor)
		}
	}
	return desc
}

func (g *DAG[T]) reverseAdj() map[T][]T {
	rev := make(map[T][]T, len(g.vertices))
	for from, dests := range g.adj {
		for _, to := range dests {
			rev[to] = append(rev[to], from)
		}
	}
	return rev
}

// String returns a textual representation of the graph.
func (g *DAG[T]) String() string {
	var sb strings.Builder
	sb.WriteString("Vertices:\n")
	for v := range g.vertices {
		sb.WriteString(fmt.Sprintf("  %v\n", v))
	}
	sb.WriteString("Edges:\n")
	for from, dests := range g.adj {
		for _, to := range dests {
			sb.WriteString(fmt.Sprintf("  %v -> %v\n", from, to))
		}
	}
	return sb.String()
}
