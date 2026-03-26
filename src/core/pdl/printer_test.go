package pdl

import (
	"testing"

	"github.com/hsqbyte/protospec/src/core/schema"
)

func TestPrinter_SimpleProtocol(t *testing.T) {
	s := &schema.ProtocolSchema{
		Name:             "UDP",
		Version:          "1.0",
		DefaultByteOrder: schema.BigEndian,
		Fields: []schema.FieldDef{
			{Name: "srcPort", Type: schema.Uint, BitWidth: 16},
			{Name: "dstPort", Type: schema.Uint, BitWidth: 16},
			{Name: "length", Type: schema.Uint, BitWidth: 16},
			{Name: "payload", Type: schema.Bytes},
		},
	}

	printer := &PDLPrinter{}
	output := printer.Print(s)

	// Round-trip: parse the output and compare
	parser := NewPDLParser(nil, nil)
	parsed, err := parser.Parse(output)
	if err != nil {
		t.Fatalf("failed to re-parse printed output: %v\nOutput:\n%s", err, output)
	}

	if parsed.Name != s.Name {
		t.Errorf("name mismatch: got %q, want %q", parsed.Name, s.Name)
	}
	if parsed.Version != s.Version {
		t.Errorf("version mismatch: got %q, want %q", parsed.Version, s.Version)
	}
	if parsed.DefaultByteOrder != s.DefaultByteOrder {
		t.Errorf("byte order mismatch: got %v, want %v", parsed.DefaultByteOrder, s.DefaultByteOrder)
	}
	if len(parsed.Fields) != len(s.Fields) {
		t.Fatalf("field count mismatch: got %d, want %d", len(parsed.Fields), len(s.Fields))
	}
}

func TestPrinter_RoundTrip_UDP(t *testing.T) {
	src := `protocol UDP version "1.0" {
  byte_order big-endian;
  field srcPort: uint16;
  field dstPort: uint16;
  field length: uint16;
  field checksum: uint16
    checksum internet-checksum covers [srcPort, dstPort, length, checksum, payload];
  field payload: bytes
    length_ref length offset -8;
}`
	parser := NewPDLParser(nil, nil)
	s1, err := parser.Parse(src)
	if err != nil {
		t.Fatalf("initial parse failed: %v", err)
	}

	printer := &PDLPrinter{}
	printed := printer.Print(s1)

	s2, err := parser.Parse(printed)
	if err != nil {
		t.Fatalf("re-parse failed: %v\nPrinted:\n%s", err, printed)
	}

	assertSchemasEqual(t, s1, s2)
}

func TestPrinter_RoundTrip_IPv4(t *testing.T) {
	src := `protocol IPv4 version "1.0" {
  byte_order big-endian;

  bitfield {
    field version: uint4;
    field ihl: uint4;
  }
  field tos: uint8;
  field totalLength: uint16;
  field identification: uint16;
  bitfield {
    field flags: uint3;
    field fragmentOffset: uint13;
  }
  field ttl: uint8;
  field protocol: uint8
    enum { 1 = "ICMP", 6 = "TCP", 17 = "UDP" };
  field headerChecksum: uint16
    checksum internet-checksum covers [version..protocol, srcAddr, dstAddr];
  field srcAddr: uint32
    display ipv4;
  field dstAddr: uint32
    display ipv4;
  field options: bytes
    length_ref ihl scale 4 offset -20
    when ihl > 5;
  field payload: bytes;
}`
	parser := NewPDLParser(nil, nil)
	s1, err := parser.Parse(src)
	if err != nil {
		t.Fatalf("initial parse failed: %v", err)
	}

	printer := &PDLPrinter{}
	printed := printer.Print(s1)

	s2, err := parser.Parse(printed)
	if err != nil {
		t.Fatalf("re-parse failed: %v\nPrinted:\n%s", err, printed)
	}

	assertSchemasEqual(t, s1, s2)
}

func TestPrinter_LittleEndian(t *testing.T) {
	s := &schema.ProtocolSchema{
		Name:             "Test",
		Version:          "2.0",
		DefaultByteOrder: schema.LittleEndian,
		Fields: []schema.FieldDef{
			{Name: "val", Type: schema.Uint, BitWidth: 32},
		},
	}

	printer := &PDLPrinter{}
	output := printer.Print(s)

	parser := NewPDLParser(nil, nil)
	parsed, err := parser.Parse(output)
	if err != nil {
		t.Fatalf("re-parse failed: %v\nOutput:\n%s", err, output)
	}
	if parsed.DefaultByteOrder != schema.LittleEndian {
		t.Errorf("expected little-endian, got %v", parsed.DefaultByteOrder)
	}
}

// assertSchemasEqual compares two schemas for equivalence.
func assertSchemasEqual(t *testing.T, a, b *schema.ProtocolSchema) {
	t.Helper()
	if a.Name != b.Name {
		t.Errorf("name: %q != %q", a.Name, b.Name)
	}
	if a.Version != b.Version {
		t.Errorf("version: %q != %q", a.Version, b.Version)
	}
	if a.DefaultByteOrder != b.DefaultByteOrder {
		t.Errorf("byte order: %v != %v", a.DefaultByteOrder, b.DefaultByteOrder)
	}
	if len(a.Fields) != len(b.Fields) {
		t.Fatalf("field count: %d != %d", len(a.Fields), len(b.Fields))
	}
	for i := range a.Fields {
		assertFieldsEqual(t, &a.Fields[i], &b.Fields[i], i)
	}
}

func assertFieldsEqual(t *testing.T, a, b *schema.FieldDef, idx int) {
	t.Helper()
	prefix := a.Name
	if a.IsBitfieldGroup {
		prefix = "bitfield"
	}

	if a.IsBitfieldGroup != b.IsBitfieldGroup {
		t.Errorf("field[%d] %s: IsBitfieldGroup %v != %v", idx, prefix, a.IsBitfieldGroup, b.IsBitfieldGroup)
		return
	}

	if a.IsBitfieldGroup {
		if len(a.BitfieldFields) != len(b.BitfieldFields) {
			t.Errorf("field[%d] bitfield: sub-field count %d != %d", idx, len(a.BitfieldFields), len(b.BitfieldFields))
			return
		}
		for j := range a.BitfieldFields {
			assertFieldsEqual(t, &a.BitfieldFields[j], &b.BitfieldFields[j], j)
		}
		return
	}

	if a.Name != b.Name {
		t.Errorf("field[%d]: name %q != %q", idx, a.Name, b.Name)
	}
	if a.Type != b.Type {
		t.Errorf("field[%d] %s: type %v != %v", idx, prefix, a.Type, b.Type)
	}
	if a.BitWidth != b.BitWidth {
		t.Errorf("field[%d] %s: bitWidth %d != %d", idx, prefix, a.BitWidth, b.BitWidth)
	}
	if a.DisplayFormat != b.DisplayFormat {
		t.Errorf("field[%d] %s: display %q != %q", idx, prefix, a.DisplayFormat, b.DisplayFormat)
	}

	// Checksum
	if (a.Checksum == nil) != (b.Checksum == nil) {
		t.Errorf("field[%d] %s: checksum nil mismatch", idx, prefix)
	} else if a.Checksum != nil {
		if a.Checksum.Algorithm != b.Checksum.Algorithm {
			t.Errorf("field[%d] %s: checksum algo %q != %q", idx, prefix, a.Checksum.Algorithm, b.Checksum.Algorithm)
		}
		if len(a.Checksum.CoverFields) != len(b.Checksum.CoverFields) {
			t.Errorf("field[%d] %s: checksum covers count %d != %d", idx, prefix, len(a.Checksum.CoverFields), len(b.Checksum.CoverFields))
		} else {
			for k := range a.Checksum.CoverFields {
				if a.Checksum.CoverFields[k] != b.Checksum.CoverFields[k] {
					t.Errorf("field[%d] %s: checksum cover[%d] %q != %q", idx, prefix, k, a.Checksum.CoverFields[k], b.Checksum.CoverFields[k])
				}
			}
		}
	}

	// LengthRef
	if (a.LengthRef == nil) != (b.LengthRef == nil) {
		t.Errorf("field[%d] %s: lengthRef nil mismatch", idx, prefix)
	} else if a.LengthRef != nil {
		if a.LengthRef.FieldName != b.LengthRef.FieldName {
			t.Errorf("field[%d] %s: lengthRef field %q != %q", idx, prefix, a.LengthRef.FieldName, b.LengthRef.FieldName)
		}
		if a.LengthRef.Scale != b.LengthRef.Scale {
			t.Errorf("field[%d] %s: lengthRef scale %d != %d", idx, prefix, a.LengthRef.Scale, b.LengthRef.Scale)
		}
		if a.LengthRef.Offset != b.LengthRef.Offset {
			t.Errorf("field[%d] %s: lengthRef offset %d != %d", idx, prefix, a.LengthRef.Offset, b.LengthRef.Offset)
		}
	}

	// Enum
	if len(a.EnumMap) != len(b.EnumMap) {
		t.Errorf("field[%d] %s: enum count %d != %d", idx, prefix, len(a.EnumMap), len(b.EnumMap))
	} else {
		for k, v := range a.EnumMap {
			if b.EnumMap[k] != v {
				t.Errorf("field[%d] %s: enum[%d] %q != %q", idx, prefix, k, v, b.EnumMap[k])
			}
		}
	}

	// Condition
	if (a.Condition == nil) != (b.Condition == nil) {
		t.Errorf("field[%d] %s: condition nil mismatch", idx, prefix)
	} else if a.Condition != nil {
		if a.Condition.FieldName != b.Condition.FieldName {
			t.Errorf("field[%d] %s: condition field %q != %q", idx, prefix, a.Condition.FieldName, b.Condition.FieldName)
		}
		if a.Condition.Operator != b.Condition.Operator {
			t.Errorf("field[%d] %s: condition op %q != %q", idx, prefix, a.Condition.Operator, b.Condition.Operator)
		}
	}
}
