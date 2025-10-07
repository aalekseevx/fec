package fecanalysis

// RecoveryCharacteristics holds the key recovery metrics for a FEC mask
type RecoveryCharacteristics struct {
	MinLostPacketsForNonRecovery     int // Minimum number of lost packets that results in non-recovery
	MinConsecutiveLostForNonRecovery int // Minimum number of consecutive lost packets that results in non-recovery
}

// CalculateRecoveryCharacteristicsFromReachable computes the recovery characteristics using existing BFS results
func CalculateRecoveryCharacteristicsFromReachable(N, K int, reachable []int) RecoveryCharacteristics {
	totalPackets := N + K

	// Convert reachable slice to set for faster lookup
	reachableSet := make(map[int]bool)
	for _, v := range reachable {
		reachableSet[v] = true
	}

	// Find characteristics
	minLostPackets := findMinLostPacketsForNonRecovery(N, K, totalPackets, reachableSet)
	minConsecutiveLost := findMinConsecutiveLostForNonRecovery(N, K, totalPackets, reachableSet)

	return RecoveryCharacteristics{
		MinLostPacketsForNonRecovery:     minLostPackets,
		MinConsecutiveLostForNonRecovery: minConsecutiveLost,
	}
}

// findMinLostPacketsForNonRecovery finds the minimum number of lost packets that results in non-recovery
func findMinLostPacketsForNonRecovery(N, K, totalPackets int, reachableSet map[int]bool) int {
	// Check all possible loss patterns, starting from 1 lost packet
	for numLost := 1; numLost <= totalPackets; numLost++ {
		// Generate all combinations of numLost lost packets
		if hasNonRecoverablePattern(N, K, totalPackets, numLost, reachableSet) {
			return numLost
		}
	}
	return -1 // No non-recoverable pattern exists (perfect recovery)
}

// findMinConsecutiveLostForNonRecovery finds the minimum number of consecutive lost packets that results in non-recovery
func findMinConsecutiveLostForNonRecovery(N, K, totalPackets int, reachableSet map[int]bool) int {
	// Check consecutive loss patterns of increasing length
	for consecutiveLen := 1; consecutiveLen <= totalPackets; consecutiveLen++ {
		// Try all possible starting positions for consecutive losses
		for startPos := 0; startPos <= totalPackets-consecutiveLen; startPos++ {
			// Create loss pattern: consecutive losses from startPos to startPos+consecutiveLen-1
			lossPattern := 0
			for i := startPos; i < startPos+consecutiveLen; i++ {
				lossPattern |= 1 << i
			}

			// Convert loss pattern to delivery pattern (invert bits)
			deliveryPattern := ((1 << totalPackets) - 1) ^ lossPattern

			// Check if this pattern is non-recoverable
			if !reachableSet[deliveryPattern] {
				return consecutiveLen
			}
		}
	}
	return -1 // No non-recoverable consecutive pattern exists (perfect recovery)
}

// hasNonRecoverablePattern checks if there exists any loss pattern with numLost packets that is non-recoverable
func hasNonRecoverablePattern(N, K, totalPackets, numLost int, reachableSet map[int]bool) bool {
	return generateCombinations(totalPackets, numLost, func(lossPattern int) bool {
		// Convert loss pattern to delivery pattern (invert bits)
		deliveryPattern := ((1 << totalPackets) - 1) ^ lossPattern

		// If this delivery pattern is not reachable, we found a non-recoverable pattern
		return !reachableSet[deliveryPattern]
	})
}

// generateCombinations generates all combinations of k bits set in n positions
// and calls the callback for each combination. Returns true if callback ever returns true.
func generateCombinations(n, k int, callback func(int) bool) bool {
	if k == 0 {
		return callback(0)
	}
	if k > n {
		return false
	}

	// Use iterative approach to generate combinations
	// Start with the first combination: k lowest bits set
	combination := (1 << k) - 1
	maxCombination := combination << (n - k)

	for combination <= maxCombination {
		if callback(combination) {
			return true
		}

		// Generate next combination using bit manipulation
		// Find rightmost set bit that can be moved right
		rightmostMovable := combination & -combination // Isolate rightmost set bit
		temp := combination + rightmostMovable

		if temp > maxCombination {
			break
		}

		// Calculate the next combination
		combination = temp | (((combination ^ temp) / rightmostMovable) >> 2)
	}

	return false
}
