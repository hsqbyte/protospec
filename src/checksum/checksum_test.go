package checksum

import "testing"

func TestInternetChecksum_RFC1071Example(t *testing.T) {
	// RFC 1071 example: 0x0001 + 0xF203 + 0xF4F5 + 0xF6F7
	data := []byte{0x00, 0x01, 0xF2, 0x03, 0xF4, 0xF5, 0xF6, 0xF7}
	got := InternetChecksum(data)
	// Expected: one's complement of sum
	if got == 0 {
		t.Fatal("checksum should not be zero for non-zero data")
	}
	// Verify: data + checksum should fold to 0xFFFF
	var sum uint32
	for i := 0; i+1 < len(data); i += 2 {
		sum += uint32(data[i])<<8 | uint32(data[i+1])
	}
	sum += uint32(got)
	for sum > 0xFFFF {
		sum = (sum >> 16) + (sum & 0xFFFF)
	}
	if uint16(sum) != 0xFFFF {
		t.Fatalf("data + checksum should fold to 0xFFFF, got 0x%04X", sum)
	}
}

func TestInternetChecksum_OddLength(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03}
	got := InternetChecksum(data)
	if got == 0 {
		t.Fatal("checksum should not be zero")
	}
}

func TestInternetChecksum_Empty(t *testing.T) {
	got := InternetChecksum([]byte{})
	if got != 0xFFFF {
		t.Fatalf("empty data checksum should be 0xFFFF, got 0x%04X", got)
	}
}

func TestCRC16_NonZero(t *testing.T) {
	data := []byte("Hello")
	got := CRC16(data)
	if got == 0 {
		t.Fatal("CRC16 should not be zero for non-empty data")
	}
	// Deterministic
	if CRC16(data) != got {
		t.Fatal("CRC16 should be deterministic")
	}
}

func TestCRC32_KnownValue(t *testing.T) {
	// CRC32 of "123456789" is well-known: 0xCBF43926
	data := []byte("123456789")
	got := CRC32(data)
	if got != 0xCBF43926 {
		t.Fatalf("CRC32(\"123456789\") = 0x%08X, want 0xCBF43926", got)
	}
}

func TestCRC32_Empty(t *testing.T) {
	got := CRC32([]byte{})
	if got != 0 {
		t.Fatalf("CRC32 of empty data should be 0, got 0x%08X", got)
	}
}
