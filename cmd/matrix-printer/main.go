package main

import (
	"fmt"
	"os"
	"path/filepath"

	fec "fec-analysis"
)

func main() {
	fmt.Println("FEC Matrix Pretty Printer")
	fmt.Println("========================")
	fmt.Println()

	// Create output directory if it doesn't exist
	outputDir := "matrices"
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

	// Generate matrices for all combinations N=1..12, K=1..N
	for _, maskType := range maskTypes {
		fmt.Printf("Generating %s matrices...\n", maskType.name)

		filename := filepath.Join(outputDir, fmt.Sprintf("%s_matrices.txt", maskType.name))
		file, err := os.Create(filename)
		if err != nil {
			fmt.Printf("Error creating file %s: %v\n", filename, err)
			continue
		}

		// Write header to file
		fmt.Fprintf(file, "%s FEC Matrices\n", maskType.name)
		fmt.Fprintf(file, "=================%s\n", repeatChar('=', len(maskType.name)))
		fmt.Fprintf(file, "\n")

		matricesGenerated := 0

		for N := 1; N <= 12; N++ {
			for K := 1; K <= N; K++ {
				// Try to create mask
				mask, err := maskType.factory.CreateMask(N, K)
				if err != nil {
					fmt.Fprintf(file, "N=%d, K=%d: Error - %v\n\n", N, K, err)
					continue
				}

				// Pretty print the matrix
				fmt.Fprintf(file, "N=%d, K=%d (Matrix: %dx%d)\n", N, K, K, N)
				fmt.Fprintf(file, "%s\n", repeatChar('-', 30))

				printMatrix(file, mask, N, K)
				fmt.Fprintf(file, "\n")
				matricesGenerated++
			}
		}

		file.Close()
		fmt.Printf("Generated %d matrices in %s\n", matricesGenerated, filename)
	}

	fmt.Println("\nMatrix generation complete!")
}

// printMatrix pretty-prints a FEC mask matrix to the file
func printMatrix(file *os.File, mask fec.Mask, N, K int) {
	// Print column headers (media packet indices)
	fmt.Fprintf(file, "     ")
	for packetIdx := 0; packetIdx < N; packetIdx++ {
		fmt.Fprintf(file, " M%-2d", packetIdx)
	}
	fmt.Fprintf(file, "\n")

	// Print separator line
	fmt.Fprintf(file, "     ")
	for packetIdx := 0; packetIdx < N; packetIdx++ {
		fmt.Fprintf(file, " ---")
	}
	fmt.Fprintf(file, "\n")

	// Print each FEC packet row
	for fecIdx := 0; fecIdx < K; fecIdx++ {
		fmt.Fprintf(file, "F%-2d |", fecIdx)

		for packetIdx := 0; packetIdx < N; packetIdx++ {
			if mask.IsProtected(packetIdx, fecIdx) {
				fmt.Fprintf(file, "  1 ")
			} else {
				fmt.Fprintf(file, "  0 ")
			}
		}
		fmt.Fprintf(file, "\n")
	}
}

// repeatChar repeats a character n times
func repeatChar(char rune, n int) string {
	result := make([]rune, n)
	for i := range result {
		result[i] = char
	}
	return string(result)
}
