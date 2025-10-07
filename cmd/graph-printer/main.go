package main

import (
	"fmt"
	"os"
	"path/filepath"

	fec "fec-analysis"
)

func main() {
	fmt.Println("FEC Graph Printer")
	fmt.Println("=================")
	fmt.Println()

	// Create output directory if it doesn't exist
	outputDir := "graphs"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}

	// Define mask types to generate
	maskTypes := []struct {
		name    string
		factory fec.MaskFactory
	}{
		{"Bursty", &fec.GoogleBurstyMaskFactory{}},
		{"Random", &fec.GoogleRandomMaskFactory{}},
		{"Interleaved", &fec.InterleavedMaskFactory{}},
	}

	// Generate graphs for all combinations N=1..6, K=1..N (limited for reasonable output size)
	for _, maskType := range maskTypes {
		fmt.Printf("Generating %s graphs...\n", maskType.name)

		filename := filepath.Join(outputDir, fmt.Sprintf("%s_graphs.txt", maskType.name))
		file, err := os.Create(filename)
		if err != nil {
			fmt.Printf("Error creating file %s: %v\n", filename, err)
			continue
		}

		// Write header to file
		fmt.Fprintf(file, "%s FEC Graphs\n", maskType.name)
		fmt.Fprintf(file, "================%s\n", repeatChar('=', len(maskType.name)))
		fmt.Fprintf(file, "\n")

		graphsGenerated := 0

		for N := 1; N <= 6; N++ {
			for K := 1; K <= N; K++ {
				// Try to create mask
				mask, err := maskType.factory.CreateMask(N, K)
				if err != nil {
					fmt.Fprintf(file, "N=%d, K=%d: Error - %v\n\n", N, K, err)
					continue
				}

				// Create recovery graph
				graph := fec.NewRecoveryGraph(mask)
				
				// Print graph representation
				printGraph(file, graph, N, K)
				graphsGenerated++
			}
		}

		file.Close()
		fmt.Printf("Generated %d graphs in %s\n", graphsGenerated, filename)
	}

	fmt.Println("\nGraph generation complete!")
}

// repeatChar repeats a character n times
func repeatChar(char rune, n int) string {
	result := make([]rune, n)
	for i := range result {
		result[i] = char
	}
	return string(result)
}