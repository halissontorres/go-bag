package graph

import (
	"slices"
	"strings"
	"testing"
)

func TestDAG_AddVertex(t *testing.T) {
	t.Parallel()

	g := NewDAG[string]()
	if !g.AddVertex("a") {
		t.Fatalf("AddVertex(a) should return true on first insert")
	}
	if g.AddVertex("a") {
		t.Fatalf("AddVertex(a) should return false on duplicate")
	}
	if !g.HasVertex("a") {
		t.Fatalf("HasVertex(a) should be true")
	}
	if g.HasVertex("missing") {
		t.Fatalf("HasVertex(missing) should be false")
	}
}

func TestDAG_AddEdge(t *testing.T) {
	t.Parallel()

	g := NewDAG[int]()
	g.AddVertex(1)
	g.AddVertex(2)
	g.AddVertex(3)

	if !g.AddEdge(1, 2) {
		t.Fatalf("AddEdge(1,2) should succeed")
	}
	if g.AddEdge(1, 2) {
		t.Fatalf("AddEdge(1,2) should fail on duplicate")
	}
	if g.AddEdge(1, 1) {
		t.Fatalf("AddEdge(1,1) should fail (self-loop)")
	}
	if g.AddEdge(1, 99) {
		t.Fatalf("AddEdge with missing destination should fail")
	}
	if g.AddEdge(99, 1) {
		t.Fatalf("AddEdge with missing source should fail")
	}

	// 1 -> 2 -> 3, then 3 -> 1 would close a cycle.
	if !g.AddEdge(2, 3) {
		t.Fatalf("AddEdge(2,3) should succeed")
	}
	if g.AddEdge(3, 1) {
		t.Fatalf("AddEdge(3,1) should fail (would create cycle)")
	}

	if !g.HasEdge(1, 2) || !g.HasEdge(2, 3) {
		t.Fatalf("HasEdge missing expected edges")
	}
	if g.HasEdge(3, 1) {
		t.Fatalf("HasEdge(3,1) should be false")
	}
}

func TestDAG_RemoveVertex(t *testing.T) {
	t.Parallel()

	g := NewDAG[int]()
	for _, v := range []int{1, 2, 3} {
		g.AddVertex(v)
	}
	g.AddEdge(1, 2)
	g.AddEdge(2, 3)
	g.AddEdge(1, 3)

	if g.RemoveVertex(99) {
		t.Fatalf("RemoveVertex(missing) should return false")
	}
	if !g.RemoveVertex(2) {
		t.Fatalf("RemoveVertex(2) should return true")
	}
	if g.HasVertex(2) {
		t.Fatalf("vertex 2 should be gone")
	}
	if g.HasEdge(1, 2) || g.HasEdge(2, 3) {
		t.Fatalf("edges incident to 2 should be gone")
	}
	if !g.HasEdge(1, 3) {
		t.Fatalf("edge 1->3 should remain")
	}
	if got := g.OutDegree(1); got != 1 {
		t.Fatalf("OutDegree(1)=%d want 1", got)
	}
	if got := g.InDegree(3); got != 1 {
		t.Fatalf("InDegree(3)=%d want 1", got)
	}
}

func TestDAG_RemoveEdge(t *testing.T) {
	t.Parallel()

	g := NewDAG[int]()
	g.AddVertex(1)
	g.AddVertex(2)
	g.AddEdge(1, 2)

	if g.RemoveEdge(1, 99) {
		t.Fatalf("RemoveEdge of missing edge should return false")
	}
	if !g.RemoveEdge(1, 2) {
		t.Fatalf("RemoveEdge(1,2) should return true")
	}
	if g.HasEdge(1, 2) {
		t.Fatalf("edge 1->2 should be gone")
	}
	if g.OutDegree(1) != 0 || g.InDegree(2) != 0 {
		t.Fatalf("degrees should be zero after removing the only edge")
	}
}

func TestDAG_TopologicalSort(t *testing.T) {
	t.Parallel()

	g := NewDAG[string]()
	for _, v := range []string{"a", "b", "c", "d", "e"} {
		g.AddVertex(v)
	}
	// a -> b -> d, a -> c -> d -> e
	g.AddEdge("a", "b")
	g.AddEdge("a", "c")
	g.AddEdge("b", "d")
	g.AddEdge("c", "d")
	g.AddEdge("d", "e")

	order, ok := g.TopologicalSort()
	if !ok {
		t.Fatalf("TopologicalSort should succeed on a DAG")
	}
	if len(order) != 5 {
		t.Fatalf("order len=%d want 5", len(order))
	}

	pos := make(map[string]int, len(order))
	for i, v := range order {
		pos[v] = i
	}
	for _, edge := range [][2]string{{"a", "b"}, {"a", "c"}, {"b", "d"}, {"c", "d"}, {"d", "e"}} {
		if pos[edge[0]] >= pos[edge[1]] {
			t.Fatalf("topological order violates %s -> %s (got %v)", edge[0], edge[1], order)
		}
	}
}

func TestDAG_TopologicalSort_Empty(t *testing.T) {
	t.Parallel()

	g := NewDAG[int]()
	order, ok := g.TopologicalSort()
	if !ok {
		t.Fatalf("empty graph should be sortable")
	}
	if len(order) != 0 {
		t.Fatalf("empty graph should yield empty order, got %v", order)
	}
}

func TestDAG_VerticesAndEdges(t *testing.T) {
	t.Parallel()

	g := NewDAG[int]()
	for _, v := range []int{1, 2, 3} {
		g.AddVertex(v)
	}
	g.AddEdge(1, 2)
	g.AddEdge(2, 3)

	verts := g.Vertices()
	slices.Sort(verts)
	if want := []int{1, 2, 3}; !slices.Equal(verts, want) {
		t.Fatalf("Vertices=%v want %v", verts, want)
	}

	edges := g.Edges()
	if len(edges) != 2 {
		t.Fatalf("Edges len=%d want 2", len(edges))
	}
	seen := make(map[[2]int]bool, len(edges))
	for _, e := range edges {
		seen[e] = true
	}
	if !seen[[2]int{1, 2}] || !seen[[2]int{2, 3}] {
		t.Fatalf("Edges missing expected pairs: %v", edges)
	}
}

func TestDAG_HasPath(t *testing.T) {
	t.Parallel()

	g := NewDAG[int]()
	for _, v := range []int{1, 2, 3, 4} {
		g.AddVertex(v)
	}
	g.AddEdge(1, 2)
	g.AddEdge(2, 3)
	// 4 is disconnected.

	if !g.HasPath(1, 3) {
		t.Fatalf("HasPath(1,3) should be true")
	}
	if !g.HasPath(1, 1) {
		t.Fatalf("HasPath(v,v) should be true (v reaches itself)")
	}
	if g.HasPath(3, 1) {
		t.Fatalf("HasPath(3,1) should be false")
	}
	if g.HasPath(1, 4) {
		t.Fatalf("HasPath(1,4) should be false (disconnected)")
	}
}

func TestDAG_AncestorsAndDescendants(t *testing.T) {
	t.Parallel()

	g := NewDAG[int]()
	for _, v := range []int{1, 2, 3, 4, 5} {
		g.AddVertex(v)
	}
	// 1 -> 2 -> 4, 1 -> 3 -> 4, 4 -> 5
	g.AddEdge(1, 2)
	g.AddEdge(1, 3)
	g.AddEdge(2, 4)
	g.AddEdge(3, 4)
	g.AddEdge(4, 5)

	desc := g.Descendants(1).Elements()
	slices.Sort(desc)
	if want := []int{2, 3, 4, 5}; !slices.Equal(desc, want) {
		t.Fatalf("Descendants(1)=%v want %v", desc, want)
	}

	anc := g.Ancestors(5).Elements()
	slices.Sort(anc)
	if want := []int{1, 2, 3, 4}; !slices.Equal(anc, want) {
		t.Fatalf("Ancestors(5)=%v want %v", anc, want)
	}

	if got := g.Descendants(5).Len(); got != 0 {
		t.Fatalf("Descendants(leaf) len=%d want 0", got)
	}
	if got := g.Ancestors(1).Len(); got != 0 {
		t.Fatalf("Ancestors(root) len=%d want 0", got)
	}
}

func TestDAG_String(t *testing.T) {
	t.Parallel()

	g := NewDAG[string]()
	g.AddVertex("x")
	g.AddVertex("y")
	g.AddEdge("x", "y")

	out := g.String()
	if !strings.Contains(out, "Vertices:") || !strings.Contains(out, "Edges:") {
		t.Fatalf("String missing section headers: %q", out)
	}
	if !strings.Contains(out, "x -> y") {
		t.Fatalf("String missing edge representation: %q", out)
	}
}

// ---------- Benchmarks ----------

func buildLinearDAG(n int) *DAG[int] {
	g := NewDAG[int]()
	for i := 0; i < n; i++ {
		g.AddVertex(i)
	}
	for i := 0; i < n-1; i++ {
		g.AddEdge(i, i+1)
	}
	return g
}

func BenchmarkDAG_AddEdge(b *testing.B) {
	const N = 1024
	b.ReportAllocs()
	for b.Loop() {
		g := NewDAG[int]()
		for i := 0; i < N; i++ {
			g.AddVertex(i)
		}
		for i := 0; i < N-1; i++ {
			g.AddEdge(i, i+1)
		}
	}
}

func BenchmarkDAG_TopologicalSort(b *testing.B) {
	g := buildLinearDAG(1024)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = g.TopologicalSort()
	}
}

func BenchmarkDAG_HasPath(b *testing.B) {
	g := buildLinearDAG(1024)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = g.HasPath(0, 1023)
	}
}

func BenchmarkDAG_Descendants(b *testing.B) {
	g := buildLinearDAG(512)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = g.Descendants(0)
	}
}
