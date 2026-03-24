package protocol

import (
	"testing"
)

func TestMVP_CreateCodecAndRoundtrip(t *testing.T) {
	lib, err := NewLibrary()
	if err != nil {
		t.Fatalf("NewLibrary failed: %v", err)
	}

	// Create a simple codec from PDL text
	pdl := `protocol Simple version "1.0" {
  byte_order big-endian;
  field a: uint8;
  field b: uint16;
  field payload: bytes;
}`
	c, err := lib.CreateCodec(pdl)
	if err != nil {
		t.Fatalf("CreateCodec failed: %v", err)
	}

	packet := map[string]any{
		"a":       uint64(42),
		"b":       uint64(1234),
		"payload": []byte{0xDE, 0xAD},
	}

	encoded, err := c.Encode(packet)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	result, err := c.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if result.Packet["a"].(uint64) != 42 {
		t.Errorf("a: got %v, want 42", result.Packet["a"])
	}
	if result.Packet["b"].(uint64) != 1234 {
		t.Errorf("b: got %v, want 1234", result.Packet["b"])
	}
	pl := result.Packet["payload"].([]byte)
	if len(pl) != 2 || pl[0] != 0xDE || pl[1] != 0xAD {
		t.Errorf("payload: got %x, want [DE AD]", pl)
	}
}

func TestMVP_BuiltinProtocolsLoaded(t *testing.T) {
	lib, err := NewLibrary()
	if err != nil {
		t.Fatalf("NewLibrary failed: %v", err)
	}

	protocols := lib.Registry().List()
	if len(protocols) == 0 {
		t.Fatal("no built-in protocols loaded")
	}

	// At minimum, UDP should be available
	if !lib.Registry().Has("UDP") {
		t.Errorf("UDP not found in loaded protocols: %v", protocols)
	}
}

func TestMVP_ProtocolNotFound(t *testing.T) {
	lib, err := NewLibrary()
	if err != nil {
		t.Fatalf("NewLibrary failed: %v", err)
	}

	_, err = lib.Encode("NonExistent", map[string]any{})
	if err == nil {
		t.Fatal("expected error for non-existent protocol")
	}
}
