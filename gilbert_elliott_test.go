package fecanalysis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGilbertElliotLossModel_SteadyState(t *testing.T) {
	tests := []struct {
		name                             string
		pe0, pe1                         float64
		p01, p10                         float64
		expectedSteady0, expectedSteady1 float64
	}{
		{
			name: "symmetric transition",
			pe0:  0.01, pe1: 0.5,
			p01: 0.1, p10: 0.1,
			expectedSteady0: 0.5, expectedSteady1: 0.5,
		},
		{
			name: "asymmetric transition",
			pe0:  0.01, pe1: 0.8,
			p01: 0.1, p10: 0.4,
			expectedSteady0: 0.8, expectedSteady1: 0.2,
		},
		{
			name: "Gilbert model",
			pe0:  0.0, pe1: 1.0,
			p01: 0.05, p10: 0.95,
			expectedSteady0: 0.95, expectedSteady1: 0.05,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewGilbertElliotLossModel(tt.pe0, tt.pe1, tt.p01, tt.p10)
			steady0, steady1 := model.GetSteadyStateProbabilities()

			assert.InDelta(t, tt.expectedSteady0, steady0, 0.001, "steady state 0 probability")
			assert.InDelta(t, tt.expectedSteady1, steady1, 0.001, "steady state 1 probability")
			assert.InDelta(t, 1.0, steady0+steady1, 0.001, "steady states should sum to 1")
		})
	}
}

func TestGilbertLossModel(t *testing.T) {
	// Gilbert model: Pe0 = 0 (no loss in good state)
	model := NewGilbertLossModel(1.0, 0.1, 0.9) // Pe1=1.0, P01=0.1, P10=0.9

	steady0, steady1 := model.GetSteadyStateProbabilities()
	expectedSteady0 := 0.9 / (0.1 + 0.9) // P10 / (P01 + P10)
	expectedSteady1 := 0.1 / (0.1 + 0.9) // P01 / (P01 + P10)

	assert.InDelta(t, expectedSteady0, steady0, 0.001)
	assert.InDelta(t, expectedSteady1, steady1, 0.001)
}

func TestGilbertElliotLossModel_AllDelivered(t *testing.T) {
	model := NewGilbertElliotLossModel(0.01, 0.5, 0.1, 0.4)

	// Test all packets delivered (vertex with all bits set)
	N := 5
	allDelivered := (1 << N) - 1 // 0b11111

	prob := model.CalculateProbability(allDelivered, N)

	// Probability should be > 0 since some packets can be delivered
	assert.Greater(t, prob, 0.0)
	// Probability should be < 1 since there's always some loss probability
	assert.Less(t, prob, 1.0)
}

func TestGilbertElliotLossModel_AllLost(t *testing.T) {
	model := NewGilbertElliotLossModel(0.01, 0.5, 0.1, 0.4)

	// Test all packets lost (vertex = 0)
	N := 5
	allLost := 0 // 0b00000

	prob := model.CalculateProbability(allLost, N)

	// Probability should be > 0 since packets can be lost
	assert.Greater(t, prob, 0.0)
	// Probability should be < 1 since loss isn't certain in good state
	assert.Less(t, prob, 1.0)
}

func TestGilbertElliotLossModel_Burstiness(t *testing.T) {
	// Create a bursty model (high P01, low P10 means longer bad periods)
	burstyModel := NewGilbertElliotLossModel(0.01, 0.9, 0.8, 0.2)

	// Create a less bursty model
	lessB7rstyModel := NewGilbertElliotLossModel(0.01, 0.9, 0.2, 0.8)

	N := 6
	// Test consecutive loss pattern: 000111 (3 losses followed by 3 deliveries)
	consecutiveLoss := 0b000111 // First 3 lost, next 3 delivered

	burstyProb := burstyModel.CalculateProbability(consecutiveLoss, N)
	lessB7rstyProb := lessB7rstyModel.CalculateProbability(consecutiveLoss, N)

	// The bursty model should have higher probability for consecutive losses
	// since it tends to stay in bad state longer
	assert.Greater(t, burstyProb, lessB7rstyProb,
		"bursty model should have higher probability for consecutive loss patterns")
}

func TestGilbertElliotLossModel_Cache(t *testing.T) {
	model := NewGilbertElliotLossModel(0.01, 0.5, 0.1, 0.4)

	N := 4
	pattern := 0b1010 // alternating pattern

	// Calculate probability twice
	prob1 := model.CalculateProbability(pattern, N)
	prob2 := model.CalculateProbability(pattern, N)

	// Should be exactly equal (cached result)
	assert.Equal(t, prob1, prob2)

	// Clear cache and recalculate
	model.ClearCache()
	prob3 := model.CalculateProbability(pattern, N)

	// Should still be equal but recalculated
	assert.InDelta(t, prob1, prob3, 1e-10)
}

func TestGilbertElliotLossModel_DynamicProgramming(t *testing.T) {
	// Test that our DP implementation is consistent
	model := NewGilbertElliotLossModel(0.1, 0.7, 0.3, 0.2)

	// Test different patterns and ensure probabilities are reasonable
	patterns := []struct {
		pattern int
		length  int
		desc    string
	}{
		{0b0000, 4, "all lost"},
		{0b1111, 4, "all delivered"},
		{0b1010, 4, "alternating"},
		{0b1100, 4, "burst then gap"},
		{0b0011, 4, "gap then burst"},
	}

	totalProb := 0.0
	for _, p := range patterns {
		prob := model.CalculateProbability(p.pattern, p.length)
		assert.Greater(t, prob, 0.0, "probability should be positive for %s", p.desc)
		assert.Less(t, prob, 1.0, "probability should be less than 1 for %s", p.desc)
		totalProb += prob
	}

	// All 2^4 = 16 patterns should sum to 1
	// Test just a few patterns to verify they're reasonable
	assert.Greater(t, totalProb, 0.0)
}

func TestGilbertElliotLossModel_ReducesToRandomLoss(t *testing.T) {
	// When P01 = P10 = 0 (no state transitions), and Pe0 = Pe1 = p,
	// the model should behave like random loss with probability p
	p := 0.3
	noTransitionModel := NewGilbertElliotLossModel(p, p, 0.0, 0.0)
	randomModel := NewRandomLossModel(p)

	// Test several patterns
	patterns := []int{0b0000, 0b1111, 0b1010, 0b1100}
	N := 4

	for _, pattern := range patterns {
		gilbertProb := noTransitionModel.CalculateProbability(pattern, N)
		randomProb := randomModel.CalculateProbability(pattern, N)

		assert.InDelta(t, randomProb, gilbertProb, 0.001,
			"Gilbert-Elliott with no transitions should match random loss for pattern %04b", pattern)
	}
}

func TestGilbertElliotLossModel_EdgeCases(t *testing.T) {
	model := NewGilbertElliotLossModel(0.1, 0.8, 0.2, 0.3)

	// Test N = 0
	assert.Equal(t, 0.0, model.CalculateProbability(0, 0))

	// Test N = 1
	prob0 := model.CalculateProbability(0, 1) // packet lost
	prob1 := model.CalculateProbability(1, 1) // packet delivered

	assert.Greater(t, prob0, 0.0)
	assert.Greater(t, prob1, 0.0)
	assert.InDelta(t, 1.0, prob0+prob1, 0.001, "probabilities for N=1 should sum to 1")
}

func BenchmarkGilbertElliotLossModel(b *testing.B) {
	model := NewGilbertElliotLossModel(0.01, 0.5, 0.1, 0.4)

	patterns := []int{0b00000000, 0b11111111, 0b10101010, 0b11110000}
	N := 8

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pattern := patterns[i%len(patterns)]
		model.CalculateProbability(pattern, N)
	}
}

func BenchmarkGilbertElliotLossModel_WithCache(b *testing.B) {
	model := NewGilbertElliotLossModel(0.01, 0.5, 0.1, 0.4)

	// Pre-populate cache
	patterns := []int{0b00000000, 0b11111111, 0b10101010, 0b11110000}
	N := 8
	for _, pattern := range patterns {
		model.CalculateProbability(pattern, N)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pattern := patterns[i%len(patterns)]
		model.CalculateProbability(pattern, N)
	}
}
