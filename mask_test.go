package fecanalysis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaskWithSpecificPattern(t *testing.T) {
	// Test the specific pattern: []byte{0xff, 0xf0}
	// This mask has one FEC packet which protects each of 12 packets
	// 0xff = 11111111 (first 8 packets protected)
	// 0xf0 = 11110000 (next 4 packets protected, last 4 not protected)
	// Total: 12 packets protected out of 16 possible

	mask := &bitMask{
		data: []byte{0xff, 0xf0},
		n:    12,
		k:    1,
	}

	// Test that first 12 packets are protected by FEC 0
	for i := 0; i < 12; i++ {
		assert.True(t, mask.IsProtected(i, 0), "Packet %d should be protected", i)
	}

	// Test that packets 12-15 are not protected by FEC 0
	for i := 12; i < 16; i++ {
		assert.False(t, mask.IsProtected(i, 0), "Packet %d should not be protected", i)
	}
}

func TestMaskBitInterpretation(t *testing.T) {
	// Test with a simple pattern to verify bit interpretation
	// 0x80, 0x00 = 10000000 00000000 (only first packet protected)
	mask := &bitMask{
		data: []byte{0x80, 0x00},
		n:    1,
		k:    1,
	}

	// Only first packet should be protected by FEC 0
	assert.True(t, mask.IsProtected(0, 0), "First packet should be protected")

	// All other packets should not be protected by FEC 0
	for i := 1; i < 16; i++ {
		assert.False(t, mask.IsProtected(i, 0), "Packet %d should not be protected", i)
	}
}

func TestMaskMultipleFECPackets(t *testing.T) {
	// Test with multiple FEC packets
	// First FEC packet: 0xff, 0x00 = 11111111 00000000 (first 8 packets protected)
	// Second FEC packet: 0x00, 0xff = 00000000 11111111 (last 8 packets protected)
	mask := &bitMask{
		data: []byte{0xff, 0x00, 0x00, 0xff},
		n:    16,
		k:    2,
	}

	// First FEC packet (index 0) protects packets 0-7
	for i := 0; i < 8; i++ {
		assert.True(t, mask.IsProtected(i, 0), "Packet %d should be protected by first FEC", i)
	}
	for i := 8; i < 16; i++ {
		assert.False(t, mask.IsProtected(i, 0), "Packet %d should not be protected by first FEC", i)
	}

	// Second FEC packet (index 1) protects packets 8-15
	for i := 0; i < 8; i++ {
		assert.False(t, mask.IsProtected(i, 1), "Packet %d should not be protected by second FEC", i)
	}
	for i := 8; i < 16; i++ {
		assert.True(t, mask.IsProtected(i, 1), "Packet %d should be protected by second FEC", i)
	}
}
