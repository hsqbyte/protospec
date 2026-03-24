package checksum

import (
	"hash/crc32"
)

// CRC16 computes CRC-16/CCITT with polynomial 0x1021 and initial value 0xFFFF.
func CRC16(data []byte) uint64 {
	crc := uint16(0xFFFF)
	for _, b := range data {
		crc ^= uint16(b) << 8
		for i := 0; i < 8; i++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return uint64(crc)
}

// CRC32 computes CRC-32/ISO using Go's standard hash/crc32 with the IEEE polynomial.
func CRC32(data []byte) uint64 {
	return uint64(crc32.ChecksumIEEE(data))
}
