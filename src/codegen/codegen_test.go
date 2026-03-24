package codegen

import (
	"strings"
	"testing"

	"github.com/hsqbyte/protospec/src/protocol"
)

func TestGenerateGoUDP(t *testing.T) {
	lib, err := protocol.NewLibrary()
	if err != nil {
		t.Fatal(err)
	}

	gen := NewGenerator(lib)
	code, err := gen.GenerateGo("UDP", "proto")
	if err != nil {
		t.Fatal(err)
	}

	// Should contain struct definition
	if !strings.Contains(code, "type UDP struct") {
		t.Error("expected UDP struct definition")
	}
	// Should contain Marshal method
	if !strings.Contains(code, "func (p *UDP) Marshal()") {
		t.Error("expected Marshal method")
	}
	// Should contain Unmarshal method
	if !strings.Contains(code, "func (p *UDP) Unmarshal(data []byte)") {
		t.Error("expected Unmarshal method")
	}
	// Should contain field
	if !strings.Contains(code, "SrcPort") {
		t.Error("expected SrcPort field")
	}
	// Should be generated code
	if !strings.Contains(code, "DO NOT EDIT") {
		t.Error("expected DO NOT EDIT header")
	}
}

func TestGenerateGoIPv4(t *testing.T) {
	lib, err := protocol.NewLibrary()
	if err != nil {
		t.Fatal(err)
	}

	gen := NewGenerator(lib)
	code, err := gen.GenerateGo("IPv4", "mypackage")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(code, "package mypackage") {
		t.Error("expected custom package name")
	}
	if !strings.Contains(code, "type IPv4 struct") {
		t.Error("expected IPv4 struct")
	}
}

func TestGenerateGoMessage(t *testing.T) {
	lib, err := protocol.NewLibrary()
	if err != nil {
		t.Fatal(err)
	}

	gen := NewGenerator(lib)
	code, err := gen.GenerateGo("JSON_RPC", "proto")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(code, "json:") {
		t.Error("expected json struct tags")
	}
}

func TestGoExportName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"src_port", "SrcPort"},
		{"srcIP", "SrcIP"},
		{"total_length", "TotalLength"},
		{"id", "ID"},
		{"ttl", "TTL"},
	}
	for _, tt := range tests {
		got := goExportName(tt.input)
		if got != tt.expected {
			t.Errorf("goExportName(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestGeneratePython(t *testing.T) {
	lib, err := protocol.NewLibrary()
	if err != nil {
		t.Fatal(err)
	}
	gen := NewGenerator(lib)
	code, err := gen.GeneratePython("UDP")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(code, "@dataclass") {
		t.Error("expected @dataclass")
	}
	if !strings.Contains(code, "def pack(self)") {
		t.Error("expected pack method")
	}
	if !strings.Contains(code, "def unpack(cls") {
		t.Error("expected unpack classmethod")
	}
}

func TestGenerateRust(t *testing.T) {
	lib, err := protocol.NewLibrary()
	if err != nil {
		t.Fatal(err)
	}
	gen := NewGenerator(lib)
	code, err := gen.GenerateRust("UDP")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(code, "pub struct") {
		t.Error("expected pub struct")
	}
	if !strings.Contains(code, "fn encode") {
		t.Error("expected encode method")
	}
	if !strings.Contains(code, "fn decode") {
		t.Error("expected decode method")
	}
}

func TestGenerateC(t *testing.T) {
	lib, err := protocol.NewLibrary()
	if err != nil {
		t.Fatal(err)
	}
	gen := NewGenerator(lib)
	code, err := gen.GenerateC("UDP")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(code, "typedef struct") {
		t.Error("expected typedef struct")
	}
	if !strings.Contains(code, "#pragma pack") {
		t.Error("expected pragma pack")
	}
}

func TestGenerateTypeScript(t *testing.T) {
	lib, err := protocol.NewLibrary()
	if err != nil {
		t.Fatal(err)
	}
	gen := NewGenerator(lib)
	code, err := gen.GenerateTypeScript("UDP")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(code, "export interface") {
		t.Error("expected export interface")
	}
	if !strings.Contains(code, "encode") {
		t.Error("expected encode function")
	}
}
