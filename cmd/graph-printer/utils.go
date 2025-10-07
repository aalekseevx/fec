package main

import "fmt"

// formatBinaryMask converts a vertex number to binary representation showing packet status
func formatBinaryMask(vertex, N, K int) string {
	binaryStr := ""
	
	// Build media packets first (bits 0 to N-1) in order M0, M1, M2, ...
	for bit := 0; bit < N; bit++ {
		if (vertex & (1 << bit)) != 0 {
			binaryStr += "1"
		} else {
			binaryStr += "0"
		}
	}
	
	// Add separator between media and FEC packets
	binaryStr += "|"
	
	// Build FEC packets (bits N to N+K-1) in order F0, F1, F2, ...
	for bit := N; bit < N + K; bit++ {
		if (vertex & (1 << bit)) != 0 {
			binaryStr += "1"
		} else {
			binaryStr += "0"
		}
	}
	
	return fmt.Sprintf("%s (%s)", binaryStr, explainBits(vertex, N, K))
}

// explainBits provides a human-readable explanation of which packets are present
func explainBits(vertex, N, K int) string {
	explanation := ""
	
	// Check media packets (bits 0 to N-1)
	mediaPresent := []int{}
	for i := 0; i < N; i++ {
		if (vertex & (1 << i)) != 0 {
			mediaPresent = append(mediaPresent, i)
		}
	}
	
	// Check FEC packets (bits N to N+K-1)
	fecPresent := []int{}
	for i := 0; i < K; i++ {
		if (vertex & (1 << (N + i))) != 0 {
			fecPresent = append(fecPresent, i)
		}
	}
	
	if len(mediaPresent) > 0 {
		explanation += "M:"
		for i, m := range mediaPresent {
			if i > 0 {
				explanation += ","
			}
			explanation += fmt.Sprintf("%d", m)
		}
	}
	
	if len(fecPresent) > 0 {
		if explanation != "" {
			explanation += " "
		}
		explanation += "F:"
		for i, f := range fecPresent {
			if i > 0 {
				explanation += ","
			}
			explanation += fmt.Sprintf("%d", f)
		}
	}
	
	if explanation == "" {
		explanation = "none"
	}
	
	return explanation
}