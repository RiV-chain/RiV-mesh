package core

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestEncodeAndDecodeIPv6(t *testing.T) {
	tests := []struct {
		name       string
		inputBytes []byte
	}{
		{"Test Case 1", []byte{0x61, 0x62, 0x63, 0x64, 0x65, 0x66}},
		{"Test Case 2", []byte{0x30, 0x31, 0x32, 0x2D, 0x39}},
		{"Test Case 3", []byte{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := encodeToIPv6([1]byte{0xfc}, tt.inputBytes)
			if err != nil {
				t.Fatalf("Error encoding: %v", err)
			}

			decoded, err := decodeIPv6(encoded)
			if err != nil {
				t.Fatalf("Error decoding: %v", err)
			}

			if !bytes.Equal(tt.inputBytes, decoded) {
				t.Errorf("Input and Decoded data mismatch")
			}
		})
	}
}

func TestTruncateTrailingZeros(t *testing.T) {
	tests := []struct {
		input    []byte
		expected []byte
	}{
		{[]byte{0x01, 0x02, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00}, []byte{0x01, 0x02, 0x03}},
		{[]byte{0x00, 0x00, 0x00}, []byte{}},
		{[]byte{0x01, 0x02, 0x03}, []byte{0x01, 0x02, 0x03}},
	}

	for _, tt := range tests {
		t.Run(hex.EncodeToString(tt.input), func(t *testing.T) {
			truncated := truncateTrailingZeros(tt.input)
			if !bytes.Equal(truncated, tt.expected) {
				t.Errorf("Expected: %v, Got: %v", tt.expected, truncated)
			}
		})
	}
}
