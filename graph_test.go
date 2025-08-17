package fecanalysis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// SimpleVisitedMarker is a basic implementation of VisitedMarker using a boolean slice
type SimpleVisitedMarker struct {
	visited []bool
}

// NewSimpleVisitedMarker creates a new SimpleVisitedMarker for the given number of vertices
func NewSimpleVisitedMarker(numVertices int) *SimpleVisitedMarker {
	return &SimpleVisitedMarker{
		visited: make([]bool, numVertices),
	}
}

// MarkReachable marks the given vertex as reachable
func (m *SimpleVisitedMarker) MarkReachable(vertex int) {
	if vertex >= 0 && vertex < len(m.visited) {
		m.visited[vertex] = true
	}
}

// IsReachable returns true if the vertex has been marked as reachable
func (m *SimpleVisitedMarker) IsReachable(vertex int) bool {
	if vertex >= 0 && vertex < len(m.visited) {
		return m.visited[vertex]
	}
	return false
}

// Reset clears all reachability marks
func (m *SimpleVisitedMarker) Reset() {
	for i := range m.visited {
		m.visited[i] = false
	}
}

// GetReachableVertices returns a slice of all vertices marked as reachable
func (m *SimpleVisitedMarker) GetReachableVertices() []int {
	var reachable []int
	for i, isReachable := range m.visited {
		if isReachable {
			reachable = append(reachable, i)
		}
	}
	return reachable
}

// SimpleGraph is a basic implementation of the Graph interface using an adjacency list
type SimpleGraph struct {
	numVertices int
	adjList     [][]int
}

// NewSimpleGraph creates a new SimpleGraph with the specified number of vertices
func NewSimpleGraph(numVertices int) *SimpleGraph {
	return &SimpleGraph{
		numVertices: numVertices,
		adjList:     make([][]int, numVertices),
	}
}

// NumVertices returns the total number of vertices in the graph
func (g *SimpleGraph) NumVertices() int {
	return g.numVertices
}

// GetEdges returns a list of edges from the given vertex
func (g *SimpleGraph) GetEdges(vertex int) []int {
	if vertex >= 0 && vertex < g.numVertices {
		return g.adjList[vertex]
	}
	return nil
}

// AddEdge adds a directed edge from source to destination vertex
func (g *SimpleGraph) AddEdge(source, destination int) {
	if source >= 0 && source < g.numVertices && destination >= 0 && destination < g.numVertices {
		g.adjList[source] = append(g.adjList[source], destination)
	}
}

// AddUndirectedEdge adds an undirected edge between two vertices
func (g *SimpleGraph) AddUndirectedEdge(vertex1, vertex2 int) {
	g.AddEdge(vertex1, vertex2)
	g.AddEdge(vertex2, vertex1)
}

func TestSimpleGraph(t *testing.T) {
	// Create a graph with 5 vertices
	graph := NewSimpleGraph(5)

	// Test basic properties
	assert.Equal(t, 5, graph.NumVertices())

	// Initially, no edges should exist
	for i := 0; i < 5; i++ {
		edges := graph.GetEdges(i)
		assert.Empty(t, edges, "Vertex %d should have no edges initially", i)
	}

	// Add some directed edges
	graph.AddEdge(0, 1)
	graph.AddEdge(0, 2)
	graph.AddEdge(1, 3)
	graph.AddEdge(2, 4)

	// Test edges from vertex 0
	edges0 := graph.GetEdges(0)
	assert.Len(t, edges0, 2)
	assert.Contains(t, edges0, 1)
	assert.Contains(t, edges0, 2)

	// Test edges from vertex 1
	edges1 := graph.GetEdges(1)
	assert.Len(t, edges1, 1)
	assert.Contains(t, edges1, 3)

	// Test edges from vertex 2
	edges2 := graph.GetEdges(2)
	assert.Len(t, edges2, 1)
	assert.Contains(t, edges2, 4)

	// Test vertices with no outgoing edges
	edges3 := graph.GetEdges(3)
	assert.Empty(t, edges3)
	edges4 := graph.GetEdges(4)
	assert.Empty(t, edges4)
}

func TestSimpleGraphUndirectedEdges(t *testing.T) {
	graph := NewSimpleGraph(3)

	// Add undirected edge between 0 and 1
	graph.AddUndirectedEdge(0, 1)

	// Both vertices should have edges to each other
	edges0 := graph.GetEdges(0)
	assert.Len(t, edges0, 1)
	assert.Contains(t, edges0, 1)

	edges1 := graph.GetEdges(1)
	assert.Len(t, edges1, 1)
	assert.Contains(t, edges1, 0)
}

func TestSimpleGraphBoundaryConditions(t *testing.T) {
	graph := NewSimpleGraph(3)

	// Test invalid vertex indices
	assert.Nil(t, graph.GetEdges(-1))
	assert.Nil(t, graph.GetEdges(3))

	// Test adding invalid edges (should not panic)
	graph.AddEdge(-1, 0)
	graph.AddEdge(0, -1)
	graph.AddEdge(3, 0)
	graph.AddEdge(0, 3)

	// Graph should remain unchanged
	for i := 0; i < 3; i++ {
		assert.Empty(t, graph.GetEdges(i))
	}
}

func TestSimpleVisitedMarker(t *testing.T) {
	marker := NewSimpleVisitedMarker(5)

	// Initially, no vertices should be marked as reachable
	for i := 0; i < 5; i++ {
		assert.False(t, marker.IsReachable(i), "Vertex %d should not be reachable initially", i)
	}

	// Mark some vertices as reachable
	marker.MarkReachable(0)
	marker.MarkReachable(2)
	marker.MarkReachable(4)

	// Test reachability
	assert.True(t, marker.IsReachable(0))
	assert.False(t, marker.IsReachable(1))
	assert.True(t, marker.IsReachable(2))
	assert.False(t, marker.IsReachable(3))
	assert.True(t, marker.IsReachable(4))

	// Test GetReachableVertices
	reachable := marker.GetReachableVertices()
	assert.Len(t, reachable, 3)
	assert.Contains(t, reachable, 0)
	assert.Contains(t, reachable, 2)
	assert.Contains(t, reachable, 4)

	// Test reset
	marker.Reset()
	for i := 0; i < 5; i++ {
		assert.False(t, marker.IsReachable(i), "Vertex %d should not be reachable after reset", i)
	}
	assert.Empty(t, marker.GetReachableVertices())
}

func TestSimpleVisitedMarkerBoundaryConditions(t *testing.T) {
	marker := NewSimpleVisitedMarker(3)

	// Test invalid indices
	assert.False(t, marker.IsReachable(-1))
	assert.False(t, marker.IsReachable(3))

	// Marking invalid indices should not panic
	marker.MarkReachable(-1)
	marker.MarkReachable(3)

	// Should not affect valid vertices
	for i := 0; i < 3; i++ {
		assert.False(t, marker.IsReachable(i))
	}
}

func TestBFSLinearGraph(t *testing.T) {
	// Create a linear graph: 0 -> 1 -> 2 -> 3 -> 4
	graph := NewSimpleGraph(5)
	graph.AddEdge(0, 1)
	graph.AddEdge(1, 2)
	graph.AddEdge(2, 3)
	graph.AddEdge(3, 4)

	// Run BFS from vertex 0
	reachable := BFS(graph, []int{0})

	// All vertices should be reachable
	assert.Len(t, reachable, 5)
	for i := 0; i < 5; i++ {
		assert.Contains(t, reachable, i, "Vertex %d should be reachable from vertex 0", i)
	}
}

func TestBFSDisconnectedGraph(t *testing.T) {
	// Create a disconnected graph: 0 -> 1, 2 -> 3, 4 (isolated)
	graph := NewSimpleGraph(5)
	graph.AddEdge(0, 1)
	graph.AddEdge(2, 3)

	// Run BFS from vertex 0
	reachable := BFS(graph, []int{0})

	// Only vertices 0 and 1 should be reachable
	assert.Len(t, reachable, 2)
	assert.Contains(t, reachable, 0)
	assert.Contains(t, reachable, 1)
	assert.NotContains(t, reachable, 2)
	assert.NotContains(t, reachable, 3)
	assert.NotContains(t, reachable, 4)
}

func TestBFSStarGraph(t *testing.T) {
	// Create a star graph: 0 connected to all other vertices
	graph := NewSimpleGraph(5)
	for i := 1; i < 5; i++ {
		graph.AddUndirectedEdge(0, i)
	}

	// Run BFS from vertex 0
	reachable := BFS(graph, []int{0})

	// All vertices should be reachable
	assert.Len(t, reachable, 5)
	for i := 0; i < 5; i++ {
		assert.Contains(t, reachable, i, "Vertex %d should be reachable from center vertex 0", i)
	}

	// Run BFS from vertex 1 (leaf node)
	reachable = BFS(graph, []int{1})

	// All vertices should still be reachable (undirected graph)
	assert.Len(t, reachable, 5)
	for i := 0; i < 5; i++ {
		assert.Contains(t, reachable, i, "Vertex %d should be reachable from leaf vertex 1", i)
	}
}

func TestBFSCyclicGraph(t *testing.T) {
	// Create a cyclic graph: 0 -> 1 -> 2 -> 0
	graph := NewSimpleGraph(3)
	graph.AddEdge(0, 1)
	graph.AddEdge(1, 2)
	graph.AddEdge(2, 0)

	// Run BFS from vertex 0
	reachable := BFS(graph, []int{0})

	// All vertices should be reachable
	assert.Len(t, reachable, 3)
	for i := 0; i < 3; i++ {
		assert.Contains(t, reachable, i, "Vertex %d should be reachable in cyclic graph", i)
	}
}

func TestBFSSingleVertex(t *testing.T) {
	// Graph with single vertex and no edges
	graph := NewSimpleGraph(1)

	// Run BFS from the only vertex
	reachable := BFS(graph, []int{0})

	// Only the source vertex should be reachable
	assert.Len(t, reachable, 1)
	assert.Contains(t, reachable, 0)
}

func TestBFSMultiSource(t *testing.T) {
	// Create a disconnected graph: 0 -> 1, 2 -> 3, 4 (isolated)
	graph := NewSimpleGraph(5)
	graph.AddEdge(0, 1)
	graph.AddEdge(2, 3)

	// Run BFS from multiple sources: 0 and 2
	reachable := BFS(graph, []int{0, 2})

	// Should reach vertices 0, 1, 2, 3 but not 4
	assert.Len(t, reachable, 4)
	assert.Contains(t, reachable, 0)
	assert.Contains(t, reachable, 1)
	assert.Contains(t, reachable, 2)
	assert.Contains(t, reachable, 3)
	assert.NotContains(t, reachable, 4)
}

func TestBFSComplexGraph(t *testing.T) {
	// Create a more complex graph
	//     0
	//   / | \
	//  1  2  3
	//  |  |  |\
	//  4  5  6 7
	graph := NewSimpleGraph(8)
	graph.AddEdge(0, 1)
	graph.AddEdge(0, 2)
	graph.AddEdge(0, 3)
	graph.AddEdge(1, 4)
	graph.AddEdge(2, 5)
	graph.AddEdge(3, 6)
	graph.AddEdge(3, 7)

	// Run BFS from root vertex 0
	reachable := BFS(graph, []int{0})

	// All vertices should be reachable
	assert.Len(t, reachable, 8)
	for i := 0; i < 8; i++ {
		assert.Contains(t, reachable, i, "Vertex %d should be reachable from root vertex 0", i)
	}

	// Run BFS from a leaf vertex
	reachable = BFS(graph, []int{4})

	// Only vertex 4 should be reachable (no outgoing edges)
	assert.Len(t, reachable, 1)
	assert.Contains(t, reachable, 4)
}

func TestBFSWithSelfLoop(t *testing.T) {
	// Create a graph with self-loop: 0 -> 0, 0 -> 1
	graph := NewSimpleGraph(2)
	graph.AddEdge(0, 0) // Self-loop
	graph.AddEdge(0, 1)

	// Run BFS from vertex 0
	reachable := BFS(graph, []int{0})

	// Both vertices should be reachable
	assert.Len(t, reachable, 2)
	assert.Contains(t, reachable, 0)
	assert.Contains(t, reachable, 1)
}

func TestBFSMultiSourceOverlapping(t *testing.T) {
	// Create a graph where sources have overlapping reachability
	//   0 -> 1 -> 2
	//   3 -> 1 -> 4
	graph := NewSimpleGraph(5)
	graph.AddEdge(0, 1)
	graph.AddEdge(1, 2)
	graph.AddEdge(3, 1)
	graph.AddEdge(1, 4)

	// Run BFS from sources 0 and 3 (both can reach 1, 2, 4)
	reachable := BFS(graph, []int{0, 3})

	// Should reach all vertices
	assert.Len(t, reachable, 5)
	for i := 0; i < 5; i++ {
		assert.Contains(t, reachable, i, "Vertex %d should be reachable", i)
	}
}

func TestBFSMultiSourceDuplicates(t *testing.T) {
	// Create a simple linear graph: 0 -> 1 -> 2
	graph := NewSimpleGraph(3)
	graph.AddEdge(0, 1)
	graph.AddEdge(1, 2)

	// Run BFS with duplicate sources
	reachable := BFS(graph, []int{0, 0, 0})

	// Should still work correctly and reach all vertices
	assert.Len(t, reachable, 3)
	assert.Contains(t, reachable, 0)
	assert.Contains(t, reachable, 1)
	assert.Contains(t, reachable, 2)
}
