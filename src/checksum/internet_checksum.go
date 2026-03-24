package checksum

// InternetChecksum computes the RFC 1071 internet checksum over data.
// It sums all 16-bit words, folds the carry bits, and returns the one's complement as uint64.
func InternetChecksum(data []byte) uint64 {
	var sum uint32
	length := len(data)
	i := 0

	// Sum all 16-bit words
	for i+1 < length {
		sum += uint32(data[i])<<8 | uint32(data[i+1])
		i += 2
	}

	// If odd number of bytes, pad with zero byte
	if i < length {
		sum += uint32(data[i]) << 8
	}

	// Fold 32-bit sum into 16 bits
	for sum > 0xFFFF {
		sum = (sum >> 16) + (sum & 0xFFFF)
	}

	// Return one's complement
	return uint64(^uint16(sum))
}
