package fecanalysis

import "sync"

// GilbertElliotLossModel implements a Gilbert-Elliott 2-state Markov chain loss model
// State 0: Good state (low loss probability Pe0)
// State 1: Bad state (high loss probability Pe1)
// Transition probabilities: P01 (good->bad), P10 (bad->good)
type GilbertElliotLossModel struct {
	Pe0 float64 // packet loss probability in good state (0)
	Pe1 float64 // packet loss probability in bad state (1)
	P01 float64 // transition probability from good (0) to bad (1)
	P10 float64 // transition probability from bad (1) to good (0)

	// Pre-computed probability cache for dynamic programming
	cache map[cacheKey]float64
	mutex sync.RWMutex

	// Steady-state probabilities
	steadyState0 float64 // steady-state probability of being in state 0
	steadyState1 float64 // steady-state probability of being in state 1
}

// cacheKey represents the key for memoizing probability calculations
type cacheKey struct {
	pattern   int // loss pattern as bitmask
	length    int // number of packets
	initState int // initial state (0 or 1)
}

// NewGilbertElliotLossModel creates a new Gilbert-Elliott loss model
func NewGilbertElliotLossModel(pe0, pe1, p01, p10 float64) *GilbertElliotLossModel {
	model := &GilbertElliotLossModel{
		Pe0:   pe0,
		Pe1:   pe1,
		P01:   p01,
		P10:   p10,
		cache: make(map[cacheKey]float64),
	}

	// Calculate steady-state probabilities
	// π₀ = P₁₀ / (P₀₁ + P₁₀)
	// π₁ = P₀₁ / (P₀₁ + P₁₀)
	denominator := p01 + p10
	if denominator > 0 {
		model.steadyState0 = p10 / denominator
		model.steadyState1 = p01 / denominator
	} else {
		// If no transitions, assume equal steady-state probabilities
		model.steadyState0 = 0.5
		model.steadyState1 = 0.5
	}

	return model
}

// NewGilbertLossModel creates a Gilbert model (Pe0 = 0)
func NewGilbertLossModel(pe1, p01, p10 float64) *GilbertElliotLossModel {
	return NewGilbertElliotLossModel(0.0, pe1, p01, p10)
}

// CalculateProbability calculates the probability of a loss pattern using dynamic programming
func (m *GilbertElliotLossModel) CalculateProbability(vertex int, N int) float64 {
	if N <= 0 {
		return 0.0
	}

	// Calculate probability starting from steady-state distribution
	prob0 := m.calculatePatternProbability(vertex, N, 0) // starting in good state
	prob1 := m.calculatePatternProbability(vertex, N, 1) // starting in bad state

	return m.steadyState0*prob0 + m.steadyState1*prob1
}

// calculatePatternProbability calculates probability of a specific loss pattern starting from a given state
func (m *GilbertElliotLossModel) calculatePatternProbability(pattern int, length int, initState int) float64 {
	if length == 0 {
		return 1.0
	}

	key := cacheKey{pattern: pattern, length: length, initState: initState}

	// Check cache first
	m.mutex.RLock()
	if prob, exists := m.cache[key]; exists {
		m.mutex.RUnlock()
		return prob
	}
	m.mutex.RUnlock()

	// Dynamic programming computation
	prob := m.computePatternProbabilityDP(pattern, length, initState)

	// Cache the result
	m.mutex.Lock()
	m.cache[key] = prob
	m.mutex.Unlock()

	return prob
}

// computePatternProbabilityDP computes pattern probability using dynamic programming
func (m *GilbertElliotLossModel) computePatternProbabilityDP(pattern int, length int, initState int) float64 {
	// DP table: dp[i][state] = probability of observing pattern[0..i-1] and ending in state
	dp := make([][2]float64, length+1)

	// Base case: start with probability 1 in the initial state
	if initState == 0 {
		dp[0][0] = 1.0
		dp[0][1] = 0.0
	} else {
		dp[0][0] = 0.0
		dp[0][1] = 1.0
	}

	// Fill DP table
	for i := 1; i <= length; i++ {
		packetIndex := i - 1
		packetDelivered := (pattern & (1 << packetIndex)) != 0

		// Calculate probabilities for each state transition
		// Transitions to state 0 (good)
		if packetDelivered {
			// Packet delivered: probability (1 - Pe_state)
			dp[i][0] = dp[i-1][0]*(1.0-m.P01)*(1.0-m.Pe0) + dp[i-1][1]*m.P10*(1.0-m.Pe0)
		} else {
			// Packet lost: probability Pe_state
			dp[i][0] = dp[i-1][0]*(1.0-m.P01)*m.Pe0 + dp[i-1][1]*m.P10*m.Pe0
		}

		// Transitions to state 1 (bad)
		if packetDelivered {
			// Packet delivered: probability (1 - Pe_state)
			dp[i][1] = dp[i-1][0]*m.P01*(1.0-m.Pe1) + dp[i-1][1]*(1.0-m.P10)*(1.0-m.Pe1)
		} else {
			// Packet lost: probability Pe_state
			dp[i][1] = dp[i-1][0]*m.P01*m.Pe1 + dp[i-1][1]*(1.0-m.P10)*m.Pe1
		}
	}

	// Return total probability (sum over all ending states)
	return dp[length][0] + dp[length][1]
}

// GetSteadyStateProbabilities returns the steady-state probabilities
func (m *GilbertElliotLossModel) GetSteadyStateProbabilities() (float64, float64) {
	return m.steadyState0, m.steadyState1
}

// ClearCache clears the probability cache (useful for testing or memory management)
func (m *GilbertElliotLossModel) ClearCache() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.cache = make(map[cacheKey]float64)
}

// GetAverageLossProbability returns the steady-state average loss probability
func (m *GilbertElliotLossModel) GetAverageLossProbability() float64 {
	return m.steadyState0*m.Pe0 + m.steadyState1*m.Pe1
}
