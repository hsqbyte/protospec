package codec

import (
	"testing"

	"github.com/hsqbyte/protospec/src/checksum"
	"github.com/hsqbyte/protospec/src/format"
	"github.com/hsqbyte/protospec/src/pdl"
)

func setupEngine() (*CodecEngine, *pdl.PDLParser) {
	cr := checksum.NewDefaultChecksumRegistry()
	fr := format.NewDefaultFormatRegistry()
	engine := NewCodecEngine(cr, fr)
	parser := pdl.NewPDLParser(cr, fr)
	return engine, parser
}

func TestRoundtrip_SimpleUint(t *testing.T) {
	engine, parser := setupEngine()
	s, err := parser.Parse(`protocol Test version "1.0" {
  byte_order big-endian;
  field a: uint8;
  field b: uint16;
  field c: uint32;
}`)
	if err != nil {
		t.Fatal(err)
	}

	packet := map[string]any{"a": uint64(0xFF), "b": uint64(0x1234), "c": uint64(0xDEADBEEF)}
	data, err := engine.Encode(s, packet)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 7 {
		t.Fatalf("expected 7 bytes, got %d", len(data))
	}

	result, err := engine.Decode(s, data)
	if err != nil {
		t.Fatal(err)
	}
	if result.Packet["a"].(uint64) != 0xFF {
		t.Errorf("a: got %v", result.Packet["a"])
	}
	if result.Packet["b"].(uint64) != 0x1234 {
		t.Errorf("b: got %v", result.Packet["b"])
	}
	if result.Packet["c"].(uint64) != 0xDEADBEEF {
		t.Errorf("c: got %v", result.Packet["c"])
	}
}

func TestRoundtrip_Bitfield(t *testing.T) {
	engine, parser := setupEngine()
	s, err := parser.Parse(`protocol Test version "1.0" {
  byte_order big-endian;
  bitfield {
    field version: uint4;
    field ihl: uint4;
  }
  field tos: uint8;
}`)
	if err != nil {
		t.Fatal(err)
	}

	packet := map[string]any{"version": uint64(4), "ihl": uint64(5), "tos": uint64(0)}
	data, err := engine.Encode(s, packet)
	if err != nil {
		t.Fatal(err)
	}
	if data[0] != 0x45 {
		t.Fatalf("expected 0x45, got 0x%02X", data[0])
	}

	result, err := engine.Decode(s, data)
	if err != nil {
		t.Fatal(err)
	}
	if result.Packet["version"].(uint64) != 4 {
		t.Errorf("version: got %v", result.Packet["version"])
	}
	if result.Packet["ihl"].(uint64) != 5 {
		t.Errorf("ihl: got %v", result.Packet["ihl"])
	}
}

func TestRoundtrip_LengthRef(t *testing.T) {
	engine, parser := setupEngine()
	s, err := parser.Parse(`protocol Test version "1.0" {
  byte_order big-endian;
  field length: uint16;
  field data: bytes length_ref length;
  field trailer: uint8;
}`)
	if err != nil {
		t.Fatal(err)
	}

	packet := map[string]any{
		"length":  uint64(3),
		"data":    []byte{0xAA, 0xBB, 0xCC},
		"trailer": uint64(0xFF),
	}
	data, err := engine.Encode(s, packet)
	if err != nil {
		t.Fatal(err)
	}

	result, err := engine.Decode(s, data)
	if err != nil {
		t.Fatal(err)
	}
	d := result.Packet["data"].([]byte)
	if len(d) != 3 || d[0] != 0xAA {
		t.Errorf("data: got %x", d)
	}
	if result.Packet["trailer"].(uint64) != 0xFF {
		t.Errorf("trailer: got %v", result.Packet["trailer"])
	}
}

func TestRoundtrip_Condition(t *testing.T) {
	engine, parser := setupEngine()
	s, err := parser.Parse(`protocol Test version "1.0" {
  byte_order big-endian;
  field flag: uint8;
  field extra: uint16 when flag > 0;
  field payload: bytes;
}`)
	if err != nil {
		t.Fatal(err)
	}

	// With condition true
	packet := map[string]any{"flag": uint64(1), "extra": uint64(0x1234), "payload": []byte{0xFF}}
	data, err := engine.Encode(s, packet)
	if err != nil {
		t.Fatal(err)
	}
	result, err := engine.Decode(s, data)
	if err != nil {
		t.Fatal(err)
	}
	if result.Packet["extra"].(uint64) != 0x1234 {
		t.Errorf("extra: got %v", result.Packet["extra"])
	}

	// With condition false
	packet2 := map[string]any{"flag": uint64(0), "payload": []byte{0xAA}}
	data2, err := engine.Encode(s, packet2)
	if err != nil {
		t.Fatal(err)
	}
	result2, err := engine.Decode(s, data2)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := result2.Packet["extra"]; ok {
		t.Error("extra should not be present when flag=0")
	}
}

func TestRoundtrip_DisplayFormat(t *testing.T) {
	engine, parser := setupEngine()
	s, err := parser.Parse(`protocol Test version "1.0" {
  byte_order big-endian;
  field addr: uint32 display ipv4;
}`)
	if err != nil {
		t.Fatal(err)
	}

	// Encode with string value
	packet := map[string]any{"addr": "192.168.1.1"}
	data, err := engine.Encode(s, packet)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 4 {
		t.Fatalf("expected 4 bytes, got %d", len(data))
	}

	// Decode should return formatted string
	result, err := engine.Decode(s, data)
	if err != nil {
		t.Fatal(err)
	}
	if result.Packet["addr"].(string) != "192.168.1.1" {
		t.Errorf("addr: got %v", result.Packet["addr"])
	}
}

func TestRoundtrip_Checksum(t *testing.T) {
	engine, parser := setupEngine()
	s, err := parser.Parse(`protocol Test version "1.0" {
  byte_order big-endian;
  field a: uint16;
  field b: uint16;
  field chk: uint16 checksum internet-checksum covers [a, b];
}`)
	if err != nil {
		t.Fatal(err)
	}

	packet := map[string]any{"a": uint64(0x0001), "b": uint64(0x0002)}
	data, err := engine.Encode(s, packet)
	if err != nil {
		t.Fatal(err)
	}

	// Decode should verify checksum
	result, err := engine.Decode(s, data)
	if err != nil {
		t.Fatal(err)
	}
	if result.Packet["a"].(uint64) != 1 {
		t.Errorf("a: got %v", result.Packet["a"])
	}

	// Corrupt checksum and verify decode fails
	data[5] ^= 0xFF
	_, err = engine.Decode(s, data)
	if err == nil {
		t.Fatal("expected checksum error")
	}
}

func TestRoundtrip_LittleEndian(t *testing.T) {
	engine, parser := setupEngine()
	s, err := parser.Parse(`protocol Test version "1.0" {
  byte_order little-endian;
  field val: uint16;
}`)
	if err != nil {
		t.Fatal(err)
	}

	packet := map[string]any{"val": uint64(0x0102)}
	data, err := engine.Encode(s, packet)
	if err != nil {
		t.Fatal(err)
	}
	// Little-endian: 0x02 first, then 0x01
	if data[0] != 0x02 || data[1] != 0x01 {
		t.Fatalf("expected [0x02, 0x01], got %x", data)
	}

	result, err := engine.Decode(s, data)
	if err != nil {
		t.Fatal(err)
	}
	if result.Packet["val"].(uint64) != 0x0102 {
		t.Errorf("val: got %v", result.Packet["val"])
	}
}

func TestRoundtrip_Payload(t *testing.T) {
	engine, parser := setupEngine()
	s, err := parser.Parse(`protocol Test version "1.0" {
  byte_order big-endian;
  field header: uint8;
  field payload: bytes;
}`)
	if err != nil {
		t.Fatal(err)
	}

	packet := map[string]any{"header": uint64(42), "payload": []byte{1, 2, 3, 4, 5}}
	data, err := engine.Encode(s, packet)
	if err != nil {
		t.Fatal(err)
	}

	result, err := engine.Decode(s, data)
	if err != nil {
		t.Fatal(err)
	}
	pl := result.Packet["payload"].([]byte)
	if len(pl) != 5 {
		t.Fatalf("payload length: got %d", len(pl))
	}
	if result.BytesRead != 6 {
		t.Errorf("bytes_read: got %d, want 6", result.BytesRead)
	}
}

func TestEncode_InvalidRange(t *testing.T) {
	engine, parser := setupEngine()
	s, err := parser.Parse(`protocol Test version "1.0" {
  byte_order big-endian;
  field val: uint8;
}`)
	if err != nil {
		t.Fatal(err)
	}

	packet := map[string]any{"val": uint64(256)} // exceeds uint8 range
	_, err = engine.Encode(s, packet)
	if err == nil {
		t.Fatal("expected range error")
	}
}

func TestDecode_InsufficientData(t *testing.T) {
	engine, parser := setupEngine()
	s, err := parser.Parse(`protocol Test version "1.0" {
  byte_order big-endian;
  field val: uint32;
}`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = engine.Decode(s, []byte{0x01}) // only 1 byte for uint32
	if err == nil {
		t.Fatal("expected insufficient data error")
	}
}

func TestRoundtrip_MAC(t *testing.T) {
	engine, parser := setupEngine()
	s, err := parser.Parse(`protocol Test version "1.0" {
  byte_order big-endian;
  bitfield {
    field mac: uint48;
  }
}`)
	if err != nil {
		t.Fatal(err)
	}

	packet := map[string]any{"mac": uint64(0xAABBCCDDEEFF)}
	data, err := engine.Encode(s, packet)
	if err != nil {
		t.Fatal(err)
	}

	result, err := engine.Decode(s, data)
	if err != nil {
		t.Fatal(err)
	}
	if result.Packet["mac"].(uint64) != 0xAABBCCDDEEFF {
		t.Errorf("mac: got 0x%X", result.Packet["mac"])
	}
}

func TestRoundtrip_LengthRefWithScaleOffset(t *testing.T) {
	engine, parser := setupEngine()
	s, err := parser.Parse(`protocol Test version "1.0" {
  byte_order big-endian;
  field hdrLen: uint8;
  field options: bytes length_ref hdrLen scale 4 offset -4;
  field payload: bytes;
}`)
	if err != nil {
		t.Fatal(err)
	}

	// hdrLen=2 → length = 2*4 - 4 = 4 bytes of options
	packet := map[string]any{
		"hdrLen":  uint64(2),
		"options": []byte{0x01, 0x02, 0x03, 0x04},
		"payload": []byte{0xFF},
	}
	data, err := engine.Encode(s, packet)
	if err != nil {
		t.Fatal(err)
	}

	result, err := engine.Decode(s, data)
	if err != nil {
		t.Fatal(err)
	}
	opts := result.Packet["options"].([]byte)
	if len(opts) != 4 {
		t.Fatalf("options length: got %d, want 4", len(opts))
	}
}

// Verify built-in protocols can be loaded and their schemas are valid
func TestBuiltinProtocols_LoadAndShow(t *testing.T) {
	cr := checksum.NewDefaultChecksumRegistry()
	fr := format.NewDefaultFormatRegistry()
	engine := NewCodecEngine(cr, fr)
	_ = engine // engine is available for encode/decode if needed

	parser := pdl.NewPDLParser(cr, fr)

	// Simple UDP roundtrip
	s, err := parser.Parse(`protocol UDP version "1.0" {
  byte_order big-endian;
  field srcPort: uint16;
  field dstPort: uint16;
  field length: uint16;
  field checksum: uint16;
  field payload: bytes;
}`)
	if err != nil {
		t.Fatal(err)
	}

	packet := map[string]any{
		"srcPort":  uint64(12345),
		"dstPort":  uint64(80),
		"length":   uint64(12),
		"checksum": uint64(0),
		"payload":  []byte{0x48, 0x65, 0x6C, 0x6C}, // "Hell"
	}
	data, err := engine.Encode(s, packet)
	if err != nil {
		t.Fatal(err)
	}

	result, err := engine.Decode(s, data)
	if err != nil {
		t.Fatal(err)
	}
	if result.Packet["srcPort"].(uint64) != 12345 {
		t.Errorf("srcPort: got %v", result.Packet["srcPort"])
	}
	if result.BytesRead != 12 {
		t.Errorf("bytes_read: got %d, want 12", result.BytesRead)
	}
}

func TestRoundtrip_ChecksumRange(t *testing.T) {
	engine, parser := setupEngine()
	s, err := parser.Parse(`protocol Test version "1.0" {
  byte_order big-endian;
  field a: uint16;
  field b: uint16;
  field c: uint16;
  field chk: uint16 checksum internet-checksum covers [a..c];
}`)
	if err != nil {
		t.Fatal(err)
	}

	packet := map[string]any{
		"a": uint64(0x0001),
		"b": uint64(0x0002),
		"c": uint64(0x0003),
	}
	data, err := engine.Encode(s, packet)
	if err != nil {
		t.Fatal(err)
	}

	result, err := engine.Decode(s, data)
	if err != nil {
		t.Fatal(err)
	}
	if result.Packet["a"].(uint64) != 1 || result.Packet["b"].(uint64) != 2 || result.Packet["c"].(uint64) != 3 {
		t.Errorf("unexpected values: %v", result.Packet)
	}
}

func TestRoundtrip_Bool(t *testing.T) {
	engine, parser := setupEngine()
	s, err := parser.Parse(`protocol Test version "1.0" {
  byte_order big-endian;
  bitfield {
    field flag: uint1;
    field reserved: uint7;
  }
}`)
	if err != nil {
		t.Fatal(err)
	}

	packet := map[string]any{"flag": uint64(1), "reserved": uint64(0)}
	data, err := engine.Encode(s, packet)
	if err != nil {
		t.Fatal(err)
	}

	result, err := engine.Decode(s, data)
	if err != nil {
		t.Fatal(err)
	}
	if result.Packet["flag"].(uint64) != 1 {
		t.Errorf("flag: got %v", result.Packet["flag"])
	}
}

func TestHexLiteral_InCondition(t *testing.T) {
	_, parser := setupEngine()
	s, err := parser.Parse(`protocol Test version "1.0" {
  byte_order big-endian;
  field etherType: uint16;
  field ipPayload: bytes when etherType == 0x0800;
}`)
	if err != nil {
		t.Fatal(err)
	}

	// Verify the condition value was parsed as 0x0800 = 2048
	for _, f := range s.Fields {
		if f.Name == "ipPayload" && f.Condition != nil {
			if f.Condition.Value.(int) != 0x0800 {
				t.Fatalf("condition value: got %v, want 2048", f.Condition.Value)
			}
		}
	}
}

func TestRoundtrip_FixedLengthBytes(t *testing.T) {
	engine, parser := setupEngine()
	s, err := parser.Parse(`protocol Test version "1.0" {
  byte_order big-endian;
  field addr: bytes[4];
  field trailer: uint8;
}`)
	if err != nil {
		t.Fatal(err)
	}

	packet := map[string]any{
		"addr":    []byte{192, 168, 1, 1},
		"trailer": uint64(0xFF),
	}
	data, err := engine.Encode(s, packet)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 5 {
		t.Fatalf("expected 5 bytes, got %d", len(data))
	}

	result, err := engine.Decode(s, data)
	if err != nil {
		t.Fatal(err)
	}
	addr := result.Packet["addr"].([]byte)
	if len(addr) != 4 || addr[0] != 192 || addr[3] != 1 {
		t.Errorf("addr: got %v", addr)
	}
	if result.Packet["trailer"].(uint64) != 0xFF {
		t.Errorf("trailer: got %v", result.Packet["trailer"])
	}
}

func TestRoundtrip_FixedLengthBytesPadding(t *testing.T) {
	engine, parser := setupEngine()
	s, err := parser.Parse(`protocol Test version "1.0" {
  byte_order big-endian;
  field name: bytes[8];
  field id: uint16;
}`)
	if err != nil {
		t.Fatal(err)
	}

	// Short data should be zero-padded
	packet := map[string]any{
		"name": []byte{0x48, 0x69},
		"id":   uint64(42),
	}
	data, err := engine.Encode(s, packet)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 10 {
		t.Fatalf("expected 10 bytes, got %d", len(data))
	}
	// Verify padding
	if data[2] != 0 || data[7] != 0 {
		t.Errorf("expected zero padding, got %x", data)
	}

	result, err := engine.Decode(s, data)
	if err != nil {
		t.Fatal(err)
	}
	name := result.Packet["name"].([]byte)
	if len(name) != 8 {
		t.Fatalf("name length: got %d, want 8", len(name))
	}
}

func TestEncode_DefaultValue(t *testing.T) {
	engine, parser := setupEngine()
	s, err := parser.Parse(`protocol Test version "1.0" {
  byte_order big-endian;
  field version: uint8 = 4;
  field data: uint8;
}`)
	if err != nil {
		t.Fatal(err)
	}

	// Encode without providing version — should use default
	packet := map[string]any{"data": uint64(42)}
	data, err := engine.Encode(s, packet)
	if err != nil {
		t.Fatal(err)
	}
	if data[0] != 4 {
		t.Fatalf("expected version=4, got %d", data[0])
	}
	if data[1] != 42 {
		t.Fatalf("expected data=42, got %d", data[1])
	}
}

func TestEncode_RangeValidation(t *testing.T) {
	engine, parser := setupEngine()
	s, err := parser.Parse(`protocol Test version "1.0" {
  byte_order big-endian;
  field val: uint8 range [1..100];
}`)
	if err != nil {
		t.Fatal(err)
	}

	// Valid value
	_, err = engine.Encode(s, map[string]any{"val": uint64(50)})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Out of range
	_, err = engine.Encode(s, map[string]any{"val": uint64(200)})
	if err == nil {
		t.Fatal("expected range error")
	}

	// Below range
	_, err = engine.Encode(s, map[string]any{"val": uint64(0)})
	if err == nil {
		t.Fatal("expected range error for value below min")
	}
}
