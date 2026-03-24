package pdl

import (
	"strings"
	"testing"

	"github.com/hsqbyte/protospec/src/checksum"
	"github.com/hsqbyte/protospec/src/errors"
	"github.com/hsqbyte/protospec/src/format"
)

func TestValidate_ValidProtocol(t *testing.T) {
	src := `protocol UDP version "1.0" {
  byte_order big-endian;
  field srcPort: uint16;
  field dstPort: uint16;
  field length: uint16;
  field payload: bytes length_ref length offset -8;
}`
	parser := NewPDLParser(nil, nil)
	ps, err := parser.Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ps.Name != "UDP" {
		t.Errorf("expected name UDP, got %s", ps.Name)
	}
}

func TestValidate_BitWidthOutOfRange(t *testing.T) {
	src := `protocol Test version "1.0" {
  byte_order big-endian;
  field bad: uint65;
}`
	parser := NewPDLParser(nil, nil)
	_, err := parser.Parse(src)
	if err == nil {
		t.Fatal("expected error for bit width 65")
	}
	var semErr *errors.PDLSemanticError
	if se, ok := err.(*errors.PDLSemanticError); ok {
		semErr = se
	} else {
		t.Fatalf("expected PDLSemanticError, got %T: %v", err, err)
	}
	if semErr.FieldName != "bad" {
		t.Errorf("expected field name 'bad', got %q", semErr.FieldName)
	}
	if !strings.Contains(semErr.Message, "bit width 65 out of valid range 1-64") {
		t.Errorf("unexpected message: %s", semErr.Message)
	}
}

func TestValidate_UndefinedLengthRefField(t *testing.T) {
	src := `protocol Test version "1.0" {
  byte_order big-endian;
  field payload: bytes length_ref noSuchField;
}`
	parser := NewPDLParser(nil, nil)
	_, err := parser.Parse(src)
	if err == nil {
		t.Fatal("expected error for undefined length_ref field")
	}
	semErr, ok := err.(*errors.PDLSemanticError)
	if !ok {
		t.Fatalf("expected PDLSemanticError, got %T: %v", err, err)
	}
	if !strings.Contains(semErr.Message, "noSuchField") {
		t.Errorf("expected message to mention 'noSuchField', got: %s", semErr.Message)
	}
}

func TestValidate_UndefinedChecksumCoverField(t *testing.T) {
	src := `protocol Test version "1.0" {
  byte_order big-endian;
  field srcPort: uint16;
  field chk: uint16 checksum internet-checksum covers [srcPort, missing];
}`
	parser := NewPDLParser(nil, nil)
	_, err := parser.Parse(src)
	if err == nil {
		t.Fatal("expected error for undefined checksum cover field")
	}
	semErr, ok := err.(*errors.PDLSemanticError)
	if !ok {
		t.Fatalf("expected PDLSemanticError, got %T: %v", err, err)
	}
	if !strings.Contains(semErr.Message, "missing") {
		t.Errorf("expected message to mention 'missing', got: %s", semErr.Message)
	}
}

func TestValidate_UndefinedChecksumCoverRangeField(t *testing.T) {
	src := `protocol Test version "1.0" {
  byte_order big-endian;
  field srcPort: uint16;
  field dstPort: uint16;
  field chk: uint16 checksum internet-checksum covers [srcPort..noEnd];
}`
	parser := NewPDLParser(nil, nil)
	_, err := parser.Parse(src)
	if err == nil {
		t.Fatal("expected error for undefined range endpoint")
	}
	semErr, ok := err.(*errors.PDLSemanticError)
	if !ok {
		t.Fatalf("expected PDLSemanticError, got %T: %v", err, err)
	}
	if !strings.Contains(semErr.Message, "noEnd") {
		t.Errorf("expected message to mention 'noEnd', got: %s", semErr.Message)
	}
}

func TestValidate_UndefinedWhenField(t *testing.T) {
	src := `protocol Test version "1.0" {
  byte_order big-endian;
  field data: bytes when ghost > 5;
}`
	parser := NewPDLParser(nil, nil)
	_, err := parser.Parse(src)
	if err == nil {
		t.Fatal("expected error for undefined when field")
	}
	semErr, ok := err.(*errors.PDLSemanticError)
	if !ok {
		t.Fatalf("expected PDLSemanticError, got %T: %v", err, err)
	}
	if !strings.Contains(semErr.Message, "ghost") {
		t.Errorf("expected message to mention 'ghost', got: %s", semErr.Message)
	}
}

func TestValidate_UnknownChecksumAlgorithm(t *testing.T) {
	src := `protocol Test version "1.0" {
  byte_order big-endian;
  field data: uint16;
  field chk: uint16 checksum bogus-algo covers [data];
}`
	cr := checksum.NewChecksumRegistry()
	parser := NewPDLParser(cr, nil)
	_, err := parser.Parse(src)
	if err == nil {
		t.Fatal("expected error for unknown checksum algorithm")
	}
	semErr, ok := err.(*errors.PDLSemanticError)
	if !ok {
		t.Fatalf("expected PDLSemanticError, got %T: %v", err, err)
	}
	if !strings.Contains(semErr.Message, "unknown checksum algorithm: bogus-algo") {
		t.Errorf("unexpected message: %s", semErr.Message)
	}
}

func TestValidate_UnknownDisplayFormat(t *testing.T) {
	src := `protocol Test version "1.0" {
  byte_order big-endian;
  field addr: uint32 display foobar;
}`
	fr := format.NewFormatRegistry()
	parser := NewPDLParser(nil, fr)
	_, err := parser.Parse(src)
	if err == nil {
		t.Fatal("expected error for unknown display format")
	}
	semErr, ok := err.(*errors.PDLSemanticError)
	if !ok {
		t.Fatalf("expected PDLSemanticError, got %T: %v", err, err)
	}
	if !strings.Contains(semErr.Message, "unknown display format: foobar") {
		t.Errorf("unexpected message: %s", semErr.Message)
	}
}

func TestValidate_NilRegistriesSkipAlgoAndFormatChecks(t *testing.T) {
	// When registries are nil, algorithm and format checks are skipped.
	src := `protocol Test version "1.0" {
  byte_order big-endian;
  field data: uint16;
  field chk: uint16 checksum whatever covers [data];
  field addr: uint32 display anything;
}`
	parser := NewPDLParser(nil, nil)
	_, err := parser.Parse(src)
	if err != nil {
		t.Fatalf("expected no error when registries are nil, got: %v", err)
	}
}

func TestValidate_RegisteredAlgorithmPasses(t *testing.T) {
	src := `protocol Test version "1.0" {
  byte_order big-endian;
  field data: uint16;
  field chk: uint16 checksum my-algo covers [data];
}`
	cr := checksum.NewChecksumRegistry()
	cr.Register("my-algo", func(data []byte) uint64 { return 0 })
	parser := NewPDLParser(cr, nil)
	_, err := parser.Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidate_BitfieldSubFieldReferences(t *testing.T) {
	// Bitfield sub-fields should be in the defined set.
	src := `protocol Test version "1.0" {
  byte_order big-endian;
  bitfield {
    field version: uint4;
    field ihl: uint4;
  }
  field options: bytes length_ref ihl;
}`
	parser := NewPDLParser(nil, nil)
	_, err := parser.Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParse_FixedLengthBytes(t *testing.T) {
	src := `protocol Test version "1.0" {
  byte_order big-endian;
  field addr: bytes[16];
  field payload: bytes;
}`
	parser := NewPDLParser(nil, nil)
	ps, err := parser.Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ps.Fields[0].FixedLength != 16 {
		t.Errorf("expected FixedLength=16, got %d", ps.Fields[0].FixedLength)
	}
}

func TestParse_HexLiteral(t *testing.T) {
	src := `protocol Test version "1.0" {
  byte_order big-endian;
  field etherType: uint16;
  field data: bytes when etherType == 0x0800;
}`
	parser := NewPDLParser(nil, nil)
	ps, err := parser.Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, f := range ps.Fields {
		if f.Name == "data" && f.Condition != nil {
			if f.Condition.Value.(int) != 0x0800 {
				t.Fatalf("expected 2048, got %v", f.Condition.Value)
			}
		}
	}
}

func TestParse_Const(t *testing.T) {
	src := `protocol Test version "1.0" {
  byte_order big-endian;
  const PORT_HTTP = 80;
  const ETHERTYPE_IPV4 = 0x0800;
  field port: uint16;
}`
	parser := NewPDLParser(nil, nil)
	ps, err := parser.Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ps.Constants["PORT_HTTP"] != 80 {
		t.Errorf("PORT_HTTP: got %d", ps.Constants["PORT_HTTP"])
	}
	if ps.Constants["ETHERTYPE_IPV4"] != 0x0800 {
		t.Errorf("ETHERTYPE_IPV4: got %d", ps.Constants["ETHERTYPE_IPV4"])
	}
}

func TestParse_DefaultValue(t *testing.T) {
	src := `protocol Test version "1.0" {
  byte_order big-endian;
  field version: uint4 = 4;
  field ihl: uint4;
}`
	parser := NewPDLParser(nil, nil)
	ps, err := parser.Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, f := range ps.Fields {
		if f.IsBitfieldGroup {
			continue
		}
		if f.Name == "version" {
			if f.DefaultValue == nil {
				t.Fatal("expected default value for version")
			}
			if f.DefaultValue.(int64) != 4 {
				t.Errorf("default value: got %v", f.DefaultValue)
			}
		}
	}
}

func TestParse_Range(t *testing.T) {
	src := `protocol Test version "1.0" {
  byte_order big-endian;
  field version: uint4 range [4..15];
}`
	parser := NewPDLParser(nil, nil)
	ps, err := parser.Parse(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	f := ps.Fields[0]
	if f.RangeMin == nil || *f.RangeMin != 4 {
		t.Errorf("RangeMin: got %v", f.RangeMin)
	}
	if f.RangeMax == nil || *f.RangeMax != 15 {
		t.Errorf("RangeMax: got %v", f.RangeMax)
	}
}

func TestParseMessage_Basic(t *testing.T) {
	src := `message TestProto version "2.0" {
  transport jsonrpc;

  request hello {
    field name: string;
    field count: number;
  }

  response hello {
    field result: string;
    field ok: boolean;
  }

  notification ping;
}`
	parser := NewPDLParser(nil, nil)
	ms, err := parser.ParseMessage(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ms.Name != "TestProto" {
		t.Errorf("name: got %q", ms.Name)
	}
	if ms.Version != "2.0" {
		t.Errorf("version: got %q", ms.Version)
	}
	if ms.Transport != "jsonrpc" {
		t.Errorf("transport: got %q", ms.Transport)
	}
	if len(ms.Messages) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(ms.Messages))
	}
	// request hello
	req := ms.Messages[0]
	if req.Kind != "request" || req.Name != "hello" {
		t.Errorf("msg[0]: got %s %s", req.Kind, req.Name)
	}
	if len(req.Fields) != 2 {
		t.Errorf("request fields: got %d", len(req.Fields))
	}
	// response hello
	resp := ms.Messages[1]
	if resp.Kind != "response" || resp.Name != "hello" {
		t.Errorf("msg[1]: got %s %s", resp.Kind, resp.Name)
	}
	// notification ping (no fields)
	notif := ms.Messages[2]
	if notif.Kind != "notification" || notif.Name != "ping" {
		t.Errorf("msg[2]: got %s %s", notif.Kind, notif.Name)
	}
	if len(notif.Fields) != 0 {
		t.Errorf("notification fields: got %d", len(notif.Fields))
	}
}

func TestParseMessage_NestedObject(t *testing.T) {
	src := `message MCP version "2024-11-05" {
  transport jsonrpc;

  request initialize {
    field protocolVersion: string;
    field capabilities: object {
      field roots: object optional;
      field sampling: object optional;
    };
    field clientInfo: object {
      field name: string;
      field version: string;
    };
  }
}`
	parser := NewPDLParser(nil, nil)
	ms, err := parser.ParseMessage(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ms.Name != "MCP" {
		t.Errorf("name: got %q", ms.Name)
	}
	req := ms.Messages[0]
	if len(req.Fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(req.Fields))
	}
	// capabilities has 2 nested optional fields
	caps := req.Fields[1]
	if caps.Name != "capabilities" {
		t.Errorf("field[1] name: got %q", caps.Name)
	}
	if len(caps.Fields) != 2 {
		t.Errorf("capabilities sub-fields: got %d", len(caps.Fields))
	}
	if !caps.Fields[0].Optional {
		t.Error("roots should be optional")
	}
	// clientInfo has 2 required fields
	ci := req.Fields[2]
	if len(ci.Fields) != 2 {
		t.Errorf("clientInfo sub-fields: got %d", len(ci.Fields))
	}
	if ci.Fields[0].Optional {
		t.Error("name should not be optional")
	}
}

func TestParseMessage_ArrayField(t *testing.T) {
	src := `message Test version "1.0" {
  transport jsonrpc;

  response list {
    field items: array;
    field count: number optional;
  }
}`
	parser := NewPDLParser(nil, nil)
	ms, err := parser.ParseMessage(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp := ms.Messages[0]
	if resp.Fields[0].Type.String() != "array" {
		t.Errorf("expected array, got %s", resp.Fields[0].Type.String())
	}
	if !resp.Fields[1].Optional {
		t.Error("count should be optional")
	}
}
