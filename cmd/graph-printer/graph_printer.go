package main

import (
	"fmt"
	"os"

	fec "fec-analysis"
)

// printGraph prints the complete graph representation with vertices and edges
func printGraph(file *os.File, graph *fec.RecoveryGraph, N, K int) {
	fmt.Fprintf(file, "N=%d, K=%d\n", N, K)
	fmt.Fprintf(file, "%s\n", repeatChar('-', 40))
	fmt.Fprintf(file, "Vertices: %d\n", graph.NumVertices())
	
	// Collect all edges as pairs
	var edgePairs []struct{ from, to int }
	for vertex := 0; vertex < graph.NumVertices(); vertex++ {
		edges := graph.GetEdges(vertex)
		for _, edge := range edges {
			edgePairs = append(edgePairs, struct{ from, to int }{vertex, edge})
		}
	}
	fmt.Fprintf(file, "Edges: %d\n", len(edgePairs))
	fmt.Fprintf(file, "\n")

	// Print vertex details with binary representation
	fmt.Fprintf(file, "Vertices (binary representation):\n")
	fmt.Fprintf(file, "Format: [Media packets M0-M%d]|[FEC packets F0-F%d]\n", N-1, K-1)
	fmt.Fprintf(file, "\n")

	for vertex := 0; vertex < graph.NumVertices(); vertex++ {
		fmt.Fprintf(file, "Vertex %3d: %s\n", vertex, formatBinaryMask(vertex, N, K))
	}
	fmt.Fprintf(file, "\n")

	// Print edges as pairs with binary representation (transposed: swap from/to)
	fmt.Fprintf(file, "Edges (transposed - from -> to):\n")
	for _, edge := range edgePairs {
		fromBinary := formatBinaryMask(edge.from, N, K)
		toBinary := formatBinaryMask(edge.to, N, K)
		fmt.Fprintf(file, "%s -> %s\n", toBinary, fromBinary)
	}
	fmt.Fprintf(file, "\n")
}