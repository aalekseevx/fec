package fecanalysis

// LossModel represents a packet loss model that calculates scenario probabilities
type LossModel interface {
	// CalculateProbability calculates the probability of a given scenario (vertex)
	// vertex represents the delivery state where bit i indicates if packet i was delivered
	// N is the total number of packets (media + FEC)
	CalculateProbability(vertex int, N int) float64

	// GetAverageLossProbability returns the average loss probability for this model
	GetAverageLossProbability() float64
}
