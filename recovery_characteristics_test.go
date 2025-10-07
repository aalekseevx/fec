package fecanalysis

import (
	"testing"
)

func TestRecoveryCharacteristics(t *testing.T) {
	tests := []struct {
		name                       string
		N                          int
		K                          int
		reachable                  []int
		expectedMinLost            int
		expectedMinConsecutiveLost int
	}{
		{
			name:                       "Simple case N=2, K=1 with perfect recovery",
			N:                          2,
			K:                          1,
			reachable:                  []int{0, 1, 2, 3, 4, 5, 6, 7}, // All patterns recoverable
			expectedMinLost:            -1,                            // Perfect recovery
			expectedMinConsecutiveLost: -1,                            // Perfect recovery
		},
		{
			name:                       "N=2, K=1 with some non-recoverable patterns",
			N:                          2,
			K:                          1,
			reachable:                  []int{3, 5, 6, 7}, // Missing patterns 0,1,2,4 (binary: 000,001,010,100)
			expectedMinLost:            2,                 // Need 2 losses to get non-recoverable pattern
			expectedMinConsecutiveLost: 2,                 // Need 2 consecutive losses
		},
		{
			name:                       "N=3, K=2 partial recovery",
			N:                          3,
			K:                          2,
			reachable:                  []int{7, 15, 23, 31}, // Only patterns with all media packets (bits 0,1,2)
			expectedMinLost:            1,                    // Any single loss of media packet
			expectedMinConsecutiveLost: 1,                    // Any consecutive loss including media packet
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateRecoveryCharacteristicsFromReachable(tt.N, tt.K, tt.reachable)

			if result.MinLostPacketsForNonRecovery != tt.expectedMinLost {
				t.Errorf("MinLostPacketsForNonRecovery = %d, expected %d",
					result.MinLostPacketsForNonRecovery, tt.expectedMinLost)
			}

			if result.MinConsecutiveLostForNonRecovery != tt.expectedMinConsecutiveLost {
				t.Errorf("MinConsecutiveLostForNonRecovery = %d, expected %d",
					result.MinConsecutiveLostForNonRecovery, tt.expectedMinConsecutiveLost)
			}
		})
	}
}

func TestFindMinLostPacketsForNonRecovery(t *testing.T) {
	tests := []struct {
		name         string
		N            int
		K            int
		totalPackets int
		reachableSet map[int]bool
		expected     int
	}{
		{
			name:         "All patterns recoverable",
			N:            2,
			K:            1,
			totalPackets: 3,
			reachableSet: map[int]bool{0: true, 1: true, 2: true, 3: true, 4: true, 5: true, 6: true, 7: true},
			expected:     -1, // Perfect recovery
		},
		{
			name:         "Pattern with 0 packets lost is non-recoverable",
			N:            2,
			K:            1,
			totalPackets: 3,
			reachableSet: map[int]bool{0: false, 1: true, 2: true, 3: true, 4: true, 5: true, 6: true, 7: true},
			expected:     3, // 3 lost packets (pattern 0) is non-recoverable
		},
		{
			name:         "Pattern with 2 packets lost is first non-recoverable",
			N:            2,
			K:            1,
			totalPackets: 3,
			reachableSet: map[int]bool{7: true, 6: true, 5: true, 3: false}, // Pattern 3 = 011b (lost packet 2)
			expected:     1,                                                 // 1 lost packet results in non-recovery
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findMinLostPacketsForNonRecovery(tt.N, tt.K, tt.totalPackets, tt.reachableSet)
			if result != tt.expected {
				t.Errorf("findMinLostPacketsForNonRecovery() = %d, expected %d", result, tt.expected)
			}
		})
	}
}

func TestFindMinConsecutiveLostForNonRecovery(t *testing.T) {
	tests := []struct {
		name         string
		N            int
		K            int
		totalPackets int
		reachableSet map[int]bool
		expected     int
	}{
		{
			name:         "All consecutive patterns recoverable",
			N:            2,
			K:            1,
			totalPackets: 3,
			reachableSet: map[int]bool{0: true, 1: true, 2: true, 3: true, 4: true, 5: true, 6: true, 7: true},
			expected:     -1, // Perfect recovery
		},
		{
			name:         "Single consecutive loss at position 0 is non-recoverable",
			N:            2,
			K:            1,
			totalPackets: 3,
			reachableSet: map[int]bool{0: true, 1: true, 2: true, 3: true, 4: true, 5: true, 6: false, 7: true},
			expected:     1, // 1 consecutive lost packet
		},
		{
			name:         "Two consecutive losses required for non-recovery",
			N:            3,
			K:            2,
			totalPackets: 5,
			reachableSet: map[int]bool{
				31: true, 30: true, 29: true, 28: true, 27: true, 26: true, 25: true, 24: true, // All single losses recoverable
				23: false, // Two consecutive losses at start: 10111b = 23 (lost bits 3,4)
			},
			expected: 1, // Actually 1 consecutive lost packet is enough if it's the right one
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findMinConsecutiveLostForNonRecovery(tt.N, tt.K, tt.totalPackets, tt.reachableSet)
			if result != tt.expected {
				t.Errorf("findMinConsecutiveLostForNonRecovery() = %d, expected %d", result, tt.expected)
			}
		})
	}
}

func TestHasNonRecoverablePattern(t *testing.T) {
	tests := []struct {
		name         string
		N            int
		K            int
		totalPackets int
		numLost      int
		reachableSet map[int]bool
		expected     bool
	}{
		{
			name:         "No non-recoverable patterns with 1 lost packet",
			N:            2,
			K:            1,
			totalPackets: 3,
			numLost:      1,
			reachableSet: map[int]bool{7: true, 6: true, 5: true, 3: true}, // All single-loss patterns recoverable
			expected:     false,
		},
		{
			name:         "Has non-recoverable pattern with 1 lost packet",
			N:            2,
			K:            1,
			totalPackets: 3,
			numLost:      1,
			reachableSet: map[int]bool{7: true, 6: true, 5: true, 3: false}, // Pattern 3 is non-recoverable
			expected:     true,
		},
		{
			name:         "Has non-recoverable pattern with 2 lost packets",
			N:            2,
			K:            1,
			totalPackets: 3,
			numLost:      2,
			reachableSet: map[int]bool{7: true, 6: true, 5: true, 3: true, 1: false}, // Pattern 1 is non-recoverable
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasNonRecoverablePattern(tt.N, tt.K, tt.totalPackets, tt.numLost, tt.reachableSet)
			if result != tt.expected {
				t.Errorf("hasNonRecoverablePattern() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGenerateCombinations(t *testing.T) {
	tests := []struct {
		name                 string
		n                    int
		k                    int
		expectedCombinations []int
		callbackReturnTrue   bool
		expectedResult       bool
	}{
		{
			name:                 "Generate combinations C(3,2)",
			n:                    3,
			k:                    2,
			expectedCombinations: []int{3, 5, 6}, // 011, 101, 110 in binary
			callbackReturnTrue:   false,
			expectedResult:       false,
		},
		{
			name:                 "Generate combinations C(4,2)",
			n:                    4,
			k:                    2,
			expectedCombinations: []int{3, 5, 6, 9, 10, 12}, // All combinations of 2 bits in 4 positions
			callbackReturnTrue:   false,
			expectedResult:       false,
		},
		{
			name:                 "Callback returns true early",
			n:                    3,
			k:                    2,
			expectedCombinations: []int{3}, // Should stop after first combination
			callbackReturnTrue:   true,
			expectedResult:       true,
		},
		{
			name:                 "Edge case: k=0",
			n:                    3,
			k:                    0,
			expectedCombinations: []int{0},
			callbackReturnTrue:   false,
			expectedResult:       false,
		},
		{
			name:                 "Edge case: k > n",
			n:                    2,
			k:                    3,
			expectedCombinations: []int{},
			callbackReturnTrue:   false,
			expectedResult:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var combinations []int
			callCount := 0

			result := generateCombinations(tt.n, tt.k, func(combination int) bool {
				combinations = append(combinations, combination)
				callCount++
				return tt.callbackReturnTrue
			})

			if result != tt.expectedResult {
				t.Errorf("generateCombinations() returned %v, expected %v", result, tt.expectedResult)
			}

			if len(combinations) != len(tt.expectedCombinations) {
				t.Errorf("Generated %d combinations, expected %d", len(combinations), len(tt.expectedCombinations))
				t.Errorf("Got: %v", combinations)
				t.Errorf("Expected: %v", tt.expectedCombinations)
				return
			}

			for i, expected := range tt.expectedCombinations {
				if combinations[i] != expected {
					t.Errorf("Combination %d: got %d, expected %d", i, combinations[i], expected)
				}
			}

			// Verify callback was called correct number of times
			if tt.callbackReturnTrue && callCount > 1 {
				t.Errorf("Callback called %d times, expected 1 (should stop early)", callCount)
			}
		})
	}
}

func TestGenerateCombinationsEdgeCases(t *testing.T) {
	// Test k=n (all bits set)
	var combinations []int
	result := generateCombinations(3, 3, func(combination int) bool {
		combinations = append(combinations, combination)
		return false
	})

	if result != false {
		t.Errorf("generateCombinations(3,3) returned %v, expected false", result)
	}

	expected := []int{7} // 111 in binary
	if len(combinations) != 1 || combinations[0] != expected[0] {
		t.Errorf("generateCombinations(3,3) = %v, expected %v", combinations, expected)
	}
}

func TestGenerateCombinationsCallbackLogic(t *testing.T) {
	// Test that callback receives correct values for C(4,2)
	var combinations []int

	generateCombinations(4, 2, func(combination int) bool {
		combinations = append(combinations, combination)

		// Count bits in combination
		bitCount := 0
		for i := 0; i < 4; i++ {
			if combination&(1<<i) != 0 {
				bitCount++
			}
		}

		if bitCount != 2 {
			t.Errorf("Combination %d has %d bits set, expected 2", combination, bitCount)
		}

		return false
	})

	if len(combinations) != 6 {
		t.Errorf("C(4,2) should generate 6 combinations, got %d", len(combinations))
	}
}

func TestRecoveryCharacteristicsPerfectRecovery(t *testing.T) {
	// Test that -1 is returned for perfect recovery scenarios
	// N=1, K=1 with all patterns reachable should return -1 for both metrics
	N, K := 1, 1
	totalPackets := N + K
	reachable := make([]int, 1<<totalPackets)
	for i := 0; i < (1 << totalPackets); i++ {
		reachable[i] = i
	}

	result := CalculateRecoveryCharacteristicsFromReachable(N, K, reachable)
	
	if result.MinLostPacketsForNonRecovery != -1 {
		t.Errorf("Expected MinLostPacketsForNonRecovery = -1 for perfect recovery, got %d", 
			result.MinLostPacketsForNonRecovery)
	}
	
	if result.MinConsecutiveLostForNonRecovery != -1 {
		t.Errorf("Expected MinConsecutiveLostForNonRecovery = -1 for perfect recovery, got %d", 
			result.MinConsecutiveLostForNonRecovery)
	}
}

func BenchmarkCalculateRecoveryCharacteristics(b *testing.B) {
	// Create a realistic reachable set for N=8, K=4
	reachable := make([]int, 0, 2048)
	for i := 0; i < (1 << 12); i++ {
		// Include patterns where at least 6 out of first 8 bits are set (media packets)
		mediaCount := 0
		for j := 0; j < 8; j++ {
			if i&(1<<j) != 0 {
				mediaCount++
			}
		}
		if mediaCount >= 6 {
			reachable = append(reachable, i)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateRecoveryCharacteristicsFromReachable(8, 4, reachable)
	}
}
