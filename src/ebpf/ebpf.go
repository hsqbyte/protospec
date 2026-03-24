// Package ebpf provides eBPF program generation for protocol parsing.
package ebpf

import (
	"fmt"
	"strings"
)

// ProgramType represents an eBPF program type.
type ProgramType string

const (
	ProgXDP ProgramType = "xdp"
	ProgTC  ProgramType = "tc"
)

// MapDef represents an eBPF map definition.
type MapDef struct {
	Name       string `json:"name"`
	Type       string `json:"type"` // hash, array, percpu_hash
	KeySize    int    `json:"key_size"`
	ValueSize  int    `json:"value_size"`
	MaxEntries int    `json:"max_entries"`
}

// Program represents a generated eBPF program.
type Program struct {
	Protocol string      `json:"protocol"`
	Type     ProgramType `json:"type"`
	Maps     []MapDef    `json:"maps"`
}

// Generate generates an eBPF C program for protocol parsing.
func Generate(protocol string, progType ProgramType) *Program {
	prog := &Program{Protocol: protocol, Type: progType}
	prog.Maps = []MapDef{
		{Name: fmt.Sprintf("%s_stats", strings.ToLower(protocol)), Type: "percpu_hash", KeySize: 4, ValueSize: 8, MaxEntries: 1024},
	}
	return prog
}

// ToC generates the eBPF C source code.
func (p *Program) ToC() string {
	var b strings.Builder
	b.WriteString("#include <linux/bpf.h>\n")
	b.WriteString("#include <bpf/bpf_helpers.h>\n\n")
	for _, m := range p.Maps {
		b.WriteString(fmt.Sprintf("struct {\n"))
		b.WriteString(fmt.Sprintf("    __uint(type, BPF_MAP_TYPE_%s);\n", strings.ToUpper(m.Type)))
		b.WriteString(fmt.Sprintf("    __uint(key_size, %d);\n", m.KeySize))
		b.WriteString(fmt.Sprintf("    __uint(value_size, %d);\n", m.ValueSize))
		b.WriteString(fmt.Sprintf("    __uint(max_entries, %d);\n", m.MaxEntries))
		b.WriteString(fmt.Sprintf("} %s SEC(\".maps\");\n\n", m.Name))
	}
	section := "xdp"
	if p.Type == ProgTC {
		section = "tc"
	}
	b.WriteString(fmt.Sprintf("SEC(\"%s\")\n", section))
	b.WriteString(fmt.Sprintf("int parse_%s(struct xdp_md *ctx) {\n", strings.ToLower(p.Protocol)))
	b.WriteString("    void *data = (void *)(long)ctx->data;\n")
	b.WriteString("    void *data_end = (void *)(long)ctx->data_end;\n")
	b.WriteString("    // Protocol parsing logic\n")
	b.WriteString("    return XDP_PASS;\n")
	b.WriteString("}\n\n")
	b.WriteString("char LICENSE[] SEC(\"license\") = \"GPL\";\n")
	return b.String()
}

// Describe returns a program description.
func (p *Program) Describe() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("eBPF Program: %s (%s)\n", p.Protocol, p.Type))
	for _, m := range p.Maps {
		b.WriteString(fmt.Sprintf("  map: %s (%s, %d entries)\n", m.Name, m.Type, m.MaxEntries))
	}
	return b.String()
}
