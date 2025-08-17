package fecanalysis

// RecoveryGraph implements the Graph interface for FEC recovery analysis
// Each vertex represents a bitset of delivered/recovered packets
// Edges represent possible recovery operations using FEC packets
type RecoveryGraph struct {
	numVertices int  // 2^(N+K) vertices
	N           int  // number of media packets
	K           int  // number of FEC packets (derived from mask)
	mask        Mask // FEC protection mask
}

// NewRecoveryGraph creates a new recovery graph with the given mask
func NewRecoveryGraph(mask Mask) *RecoveryGraph {
	N := mask.N()
	K := mask.K()
	numVertices := 1 << (N + K) // 2^(N+K) vertices

	return &RecoveryGraph{
		numVertices: numVertices,
		N:           N,
		K:           K,
		mask:        mask,
	}
}

// NumVertices returns the total number of vertices in the graph (2^(N+K))
func (g *RecoveryGraph) NumVertices() int {
	return g.numVertices
}

// GetEdges returns a list of edges from the given vertex, calculated on demand
func (g *RecoveryGraph) GetEdges(vertex int) []int {
	if vertex < 0 || vertex >= g.numVertices {
		return nil
	}

	var edges []int

	// For each FEC packet
	for fecIndex := 0; fecIndex < g.K; fecIndex++ {
		// Check if all packets protected by this FEC packet are present in current vertex
		if g.canUseFECPacket(vertex, fecIndex) {
			// Add edges to vertices where we can recover missing packets
			edges = g.addRecoveryEdges(edges, vertex, fecIndex)
		}
	}

	return edges
}

// canUseFECPacket checks if the FEC packet is delivered and all packets protected by it are present in the vertex
func (g *RecoveryGraph) canUseFECPacket(vertex int, fecIndex int) bool {
	// Check if the FEC packet itself is delivered (bit N+fecIndex)
	fecBitPosition := g.N + fecIndex
	if (vertex & (1 << fecBitPosition)) == 0 {
		return false // FEC packet is not delivered
	}

	// Check if all protected packets are present
	for packetIndex := 0; packetIndex < g.N; packetIndex++ {
		if g.mask.IsProtected(packetIndex, fecIndex) {
			// Check if this packet is present in the vertex (bit is set)
			if (vertex & (1 << packetIndex)) == 0 {
				return false // This protected packet is missing
			}
		}
	}
	return true
}

// addRecoveryEdges adds edges from the current vertex to vertices with recovered packets
func (g *RecoveryGraph) addRecoveryEdges(edges []int, vertex int, fecIndex int) []int {
	// For each protected packet, create an edge to a vertex where that packet is removed
	for packetIndex := 0; packetIndex < g.N; packetIndex++ {
		if g.mask.IsProtected(packetIndex, fecIndex) {
			// Create destination vertex by removing this packet (clearing the bit)
			destVertex := vertex &^ (1 << packetIndex)

			// Only add edge if destination vertex is different (packet was actually present)
			if destVertex != vertex {
				edges = append(edges, destVertex)
			}
		}
	}

	return edges
}
