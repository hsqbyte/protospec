package format

import "testing"

func TestIPv4_Roundtrip(t *testing.T) {
	f := &IPv4Formatter{}
	encoded, err := f.Encode(uint64(0xC0A80101)) // 192.168.1.1
	if err != nil {
		t.Fatal(err)
	}
	if encoded != "192.168.1.1" {
		t.Fatalf("got %q, want 192.168.1.1", encoded)
	}
	decoded, err := f.Decode(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.(uint64) != 0xC0A80101 {
		t.Fatalf("got 0x%X, want 0xC0A80101", decoded)
	}
}

func TestIPv4_Zero(t *testing.T) {
	f := &IPv4Formatter{}
	s, err := f.Encode(uint64(0))
	if err != nil {
		t.Fatal(err)
	}
	if s != "0.0.0.0" {
		t.Fatalf("got %q, want 0.0.0.0", s)
	}
}

func TestIPv4_InvalidDecode(t *testing.T) {
	f := &IPv4Formatter{}
	_, err := f.Decode("not-an-ip")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMAC_Roundtrip(t *testing.T) {
	f := &MACFormatter{}
	encoded, err := f.Encode(uint64(0xAABBCCDDEEFF))
	if err != nil {
		t.Fatal(err)
	}
	if encoded != "aa:bb:cc:dd:ee:ff" {
		t.Fatalf("got %q", encoded)
	}
	decoded, err := f.Decode(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.(uint64) != 0xAABBCCDDEEFF {
		t.Fatalf("got 0x%X", decoded)
	}
}

func TestMAC_InvalidDecode(t *testing.T) {
	f := &MACFormatter{}
	_, err := f.Decode("not-a-mac")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestIPv6_Roundtrip(t *testing.T) {
	f := &IPv6Formatter{}
	// ::1 (loopback)
	input := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	encoded, err := f.Encode(input)
	if err != nil {
		t.Fatal(err)
	}
	if encoded != "0000:0000:0000:0000:0000:0000:0000:0001" {
		t.Fatalf("got %q", encoded)
	}
	decoded, err := f.Decode("::1")
	if err != nil {
		t.Fatal(err)
	}
	if decoded == nil {
		t.Fatal("decoded is nil")
	}
}

func TestIPv6_InvalidDecode(t *testing.T) {
	f := &IPv6Formatter{}
	_, err := f.Decode("not-ipv6")
	if err == nil {
		t.Fatal("expected error")
	}
}
