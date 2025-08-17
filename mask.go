package fecanalysis

import "fmt"

// Mask represents a FEC protection mask that determines which symbols are protected
type Mask interface {
	// IsProtected returns true if the packet at packetIndex is protected by FEC at fecIndex
	IsProtected(packetIndex, fecIndex int) bool
	// N returns the number of media packets
	N() int
	// K returns the number of FEC packets
	K() int
}

// MaskFactory creates masks with specified parameters
type MaskFactory interface {
	// CreateMask creates a mask with N total symbols and K protection symbols
	CreateMask(N, K int) (Mask, error)
}

// bitMask represents a mask implementation using bit patterns
type bitMask struct {
	data []byte
	n    int // number of media packets
	k    int // number of FEC packets
}

// IsProtected checks if the packet at packetIndex is protected by FEC at fecIndex
func (m *bitMask) IsProtected(packetIndex, fecIndex int) bool {
	// Check bounds
	if packetIndex < 0 || packetIndex >= 16 || fecIndex < 0 {
		return false
	}

	// Check if we have enough bytes for this FEC index
	// Each FEC packet is 2 bytes
	if fecIndex*2 > len(m.data) {
		return false
	}

	// Calculate byte and bit position within the FEC packet
	byteOffset := fecIndex * 2
	if packetIndex < 8 {
		// First byte of the FEC packet
		bitPos := 7 - packetIndex // MSB first
		return (m.data[byteOffset] & (1 << bitPos)) != 0
	} else {
		// Second byte of the FEC packet
		bitPos := 7 - (packetIndex - 8) // MSB first
		return (m.data[byteOffset+1] & (1 << bitPos)) != 0
	}
}

// N returns the number of media packets
func (m *bitMask) N() int {
	return m.n
}

// K returns the number of FEC packets
func (m *bitMask) K() int {
	return m.k
}

// InterleavedMask implements interleaved protection where each packet is protected by one FEC packet
// The FEC packet index is determined by media_packet % K
type InterleavedMask struct {
	n int // number of media packets
	k int // number of FEC packets
}

// IsProtected returns true if the packet at packetIndex is protected by FEC at fecIndex
func (m *InterleavedMask) IsProtected(packetIndex, fecIndex int) bool {
	if packetIndex < 0 || packetIndex >= m.n || fecIndex < 0 || fecIndex >= m.k {
		return false
	}
	// Each packet is protected by exactly one FEC packet: media_packet % K
	return packetIndex%m.k == fecIndex
}

// N returns the number of media packets
func (m *InterleavedMask) N() int {
	return m.n
}

// K returns the number of FEC packets
func (m *InterleavedMask) K() int {
	return m.k
}

// InterleavedMaskFactory creates interleaved protection masks
type InterleavedMaskFactory struct{}

// CreateMask creates an interleaved mask with N media packets and K FEC packets
func (f *InterleavedMaskFactory) CreateMask(N, K int) (Mask, error) {
	if N <= 0 || K <= 0 || K > N {
		return nil, fmt.Errorf("invalid parameters for interleaved mask: N=%d, K=%d", N, K)
	}

	return &InterleavedMask{
		n: N,
		k: K,
	}, nil
}
