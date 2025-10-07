package fecanalysis

import "math"

// RandomLossModel implements a random loss model with uniform packet loss probability
type RandomLossModel struct {
	P float64 // packet loss probability (0.0 to 1.0)
}

// NewRandomLossModel creates a new random loss model with the given packet loss probability
func NewRandomLossModel(p float64) *RandomLossModel {
	return &RandomLossModel{P: p}
}

// CalculateProbability calculates the probability of a scenario under random loss
// Probability = p^(number of zeros) * (1-p)^(number of ones)
func (m *RandomLossModel) CalculateProbability(vertex int, N int) float64 {
	if N <= 0 {
		return 0.0
	}

	onesCount := 0
	zerosCount := 0

	// Count ones and zeros in the vertex (delivered vs lost packets)
	for i := 0; i < N; i++ {
		if (vertex & (1 << i)) != 0 {
			onesCount++ // packet delivered
		} else {
			zerosCount++ // packet lost
		}
	}

	// Probability = p^(zeros) * (1-p)^(ones)
	return math.Pow(m.P, float64(zerosCount)) * math.Pow(1.0-m.P, float64(onesCount))
}

// GetAverageLossProbability returns the average loss probability for this model
func (m *RandomLossModel) GetAverageLossProbability() float64 {
	return m.P
}
