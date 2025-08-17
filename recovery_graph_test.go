package fecanalysis

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SimpleMask is a test implementation of the Mask interface
type SimpleMask struct {
	protectionMatrix [][]bool // [fecIndex][packetIndex]
	n                int      // number of media packets
	k                int      // number of FEC packets
}

// NewSimpleMask creates a new SimpleMask with the given protection matrix
func NewSimpleMask(protectionMatrix [][]bool, n, k int) *SimpleMask {
	return &SimpleMask{
		protectionMatrix: protectionMatrix,
		n:                n,
		k:                k,
	}
}

// IsProtected returns true if the packet at packetIndex is protected by FEC at fecIndex
func (m *SimpleMask) IsProtected(packetIndex, fecIndex int) bool {
	if fecIndex < 0 || fecIndex >= len(m.protectionMatrix) {
		return false
	}
	if packetIndex < 0 || packetIndex >= len(m.protectionMatrix[fecIndex]) {
		return false
	}
	return m.protectionMatrix[fecIndex][packetIndex]
}

// N returns the number of media packets
func (m *SimpleMask) N() int {
	return m.n
}

// K returns the number of FEC packets
func (m *SimpleMask) K() int {
	return m.k
}

func TestRecoveryGraphBasic(t *testing.T) {
	// Create a simple mask where FEC 0 protects packets 0 and 1
	protectionMatrix := [][]bool{
		{true, true, false}, // FEC 0 protects packets 0 and 1
	}
	mask := NewSimpleMask(protectionMatrix, 3, 1)

	// Create recovery graph with N=3 media packets and K=1 FEC packets
	graph := NewRecoveryGraph(mask)

	// Test basic properties
	assert.Equal(t, 16, graph.NumVertices()) // 2^(3+1) = 16 vertices
	assert.Equal(t, 3, graph.N)
	assert.Equal(t, 1, graph.K)

	// Test that graph implements Graph interface
	var graphInterface Graph = graph
	require.NotNil(t, graphInterface)
}

func TestRecoveryGraphEdges(t *testing.T) {
	// Create a mask where FEC 0 protects packets 0 and 1
	protectionMatrix := [][]bool{
		{true, true, false}, // FEC 0 protects packets 0 and 1
	}
	mask := NewSimpleMask(protectionMatrix, 3, 1)
	graph := NewRecoveryGraph(mask)

	// Now with N=3, K=1, vertices are 4 bits: [media2][media1][media0][fec0]
	// Vertex 11 (binary 1011) has packets 0,1 and FEC 0, so FEC 0 can be used
	edges := graph.GetEdges(11)
	assert.Len(t, edges, 2)
	assert.Contains(t, edges, 9)  // Remove packet 1: 1011 -> 1001
	assert.Contains(t, edges, 10) // Remove packet 0: 1011 -> 1010

	// Vertex 15 (binary 1111) has all packets and FEC 0, so FEC 0 can be used
	edges = graph.GetEdges(15)
	assert.Len(t, edges, 2)
	assert.Contains(t, edges, 13) // Remove packet 1: 1111 -> 1101
	assert.Contains(t, edges, 14) // Remove packet 0: 1111 -> 1110

	// Vertex 3 (binary 0011) has packets 0,1 but no FEC 0, so FEC 0 cannot be used
	edges = graph.GetEdges(3)
	assert.Empty(t, edges)

	// Vertex 0 (binary 0000) has no packets and no FEC, so FEC 0 cannot be used
	edges = graph.GetEdges(0)
	assert.Empty(t, edges)
}

func TestRecoveryGraphMultipleFEC(t *testing.T) {
	// Create a mask with multiple FEC packets
	protectionMatrix := [][]bool{
		{true, true, false}, // FEC 0 protects packets 0 and 1
		{false, true, true}, // FEC 1 protects packets 1 and 2
	}
	mask := NewSimpleMask(protectionMatrix, 3, 2)
	graph := NewRecoveryGraph(mask)

	assert.Equal(t, 32, graph.NumVertices()) // 2^(3+2) = 32 vertices
	assert.Equal(t, 3, graph.N)
	assert.Equal(t, 2, graph.K)

	// Now with N=3, K=2, vertices are 5 bits: [media2][media1][media0][fec1][fec0]
	// Vertex 31 (binary 11111) has all media packets and both FEC packets
	edges := graph.GetEdges(31)
	assert.Len(t, edges, 4)       // 2 edges from FEC 0 + 2 edges from FEC 1
	assert.Contains(t, edges, 29) // Remove packet 1 (FEC 0): 11111 -> 11101
	assert.Contains(t, edges, 30) // Remove packet 0 (FEC 0): 11111 -> 11110
	assert.Contains(t, edges, 27) // Remove packet 2 (FEC 1): 11111 -> 11011
	assert.Contains(t, edges, 29) // Remove packet 1 (FEC 1): 11111 -> 11101 (duplicate)

	// Remove duplicates and check again
	uniqueEdges := make(map[int]bool)
	for _, edge := range edges {
		uniqueEdges[edge] = true
	}
	assert.Len(t, uniqueEdges, 3)   // Should have 3 unique destinations
	assert.True(t, uniqueEdges[27]) // 11111 -> 11011
	assert.True(t, uniqueEdges[29]) // 11111 -> 11101
	assert.True(t, uniqueEdges[30]) // 11111 -> 11110
}

func TestRecoveryGraphBoundaryConditions(t *testing.T) {
	protectionMatrix := [][]bool{
		{true, false},
	}
	mask := NewSimpleMask(protectionMatrix, 2, 1)
	graph := NewRecoveryGraph(mask)

	// Test invalid vertex indices
	assert.Nil(t, graph.GetEdges(-1))
	assert.Nil(t, graph.GetEdges(8)) // 2^(2+1) = 8 vertices (0-7)
}

func TestRecoveryGraphWithBurstyMask(t *testing.T) {
	// Test with actual bursty mask
	factory := &GoogleBurstyMaskFactory{}
	mask, err := factory.CreateMask(3, 2) // N=3, K=2
	require.NoError(t, err)

	graph := NewRecoveryGraph(mask)
	assert.Equal(t, 32, graph.NumVertices()) // 2^(3+2) = 32 vertices
	assert.Equal(t, 3, graph.N)

	// Test that the graph has some edges (exact edges depend on the mask pattern)
	hasEdges := false
	for vertex := 0; vertex < graph.NumVertices(); vertex++ {
		if len(graph.GetEdges(vertex)) > 0 {
			hasEdges = true
			break
		}
	}
	assert.True(t, hasEdges, "Graph should have at least some edges")
}

func TestRecoveryGraphBFS(t *testing.T) {
	// Create a simple mask for testing BFS
	protectionMatrix := [][]bool{
		{true, true, false}, // FEC 0 protects packets 0 and 1
	}
	mask := NewSimpleMask(protectionMatrix, 3, 1)
	graph := NewRecoveryGraph(mask)

	// Test BFS from vertex 15 (binary 1111) - all media packets and FEC 0 present
	reachable := BFS(graph, []int{15})

	// From vertex 15, we should be able to reach vertices with fewer packets
	assert.Contains(t, reachable, 15) // Source vertex
	assert.Contains(t, reachable, 13) // Remove packet 1: 1111 -> 1101
	assert.Contains(t, reachable, 14) // Remove packet 0: 1111 -> 1110

	// Test BFS from vertex with no outgoing edges
	reachable = BFS(graph, []int{0})
	assert.Len(t, reachable, 1)
	assert.Contains(t, reachable, 0)
}

func TestRecoveryGraphInterfaceCompliance(t *testing.T) {
	protectionMatrix := [][]bool{
		{true, true},
	}
	mask := NewSimpleMask(protectionMatrix, 2, 1)

	// Test that RecoveryGraph implements Graph interface
	var graph Graph = NewRecoveryGraph(mask)
	require.NotNil(t, graph)
	assert.Equal(t, 8, graph.NumVertices()) // 2^(2+1) = 8 vertices

	// Test that BFS works with RecoveryGraph
	reachable := BFS(graph, []int{7}) // Vertex with both media packets and FEC 0
	assert.Contains(t, reachable, 7)
}
