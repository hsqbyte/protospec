// Package p4gen provides P4 programmable switch code generation.
package p4gen

import (
	"fmt"
	"strings"
)

// Header represents a P4 header definition.
type Header struct {
	Name   string         `json:"name"`
	Fields map[string]int `json:"fields"` // field name → bit width
}

// Table represents a P4 match-action table.
type Table struct {
	Name    string   `json:"name"`
	Keys    []string `json:"keys"`
	Actions []string `json:"actions"`
	Size    int      `json:"size"`
}

// Program represents a P4 program.
type Program struct {
	Protocol string   `json:"protocol"`
	Headers  []Header `json:"headers"`
	Tables   []Table  `json:"tables"`
}

// Generate generates a P4 program from protocol definition.
func Generate(protocol string, fields map[string]int) *Program {
	prog := &Program{Protocol: protocol}
	prog.Headers = []Header{{Name: strings.ToLower(protocol) + "_t", Fields: fields}}
	prog.Tables = []Table{{
		Name:    strings.ToLower(protocol) + "_table",
		Keys:    []string{"hdr." + strings.ToLower(protocol) + ".version"},
		Actions: []string{"forward", "drop", "NoAction"},
		Size:    1024,
	}}
	return prog
}

// ToP4 generates P4_16 source code.
func (p *Program) ToP4() string {
	var b strings.Builder
	b.WriteString("#include <core.p4>\n#include <v1model.p4>\n\n")
	for _, h := range p.Headers {
		b.WriteString(fmt.Sprintf("header %s {\n", h.Name))
		for name, bits := range h.Fields {
			b.WriteString(fmt.Sprintf("    bit<%d> %s;\n", bits, name))
		}
		b.WriteString("}\n\n")
	}
	for _, t := range p.Tables {
		b.WriteString(fmt.Sprintf("table %s {\n", t.Name))
		b.WriteString("    key = {\n")
		for _, k := range t.Keys {
			b.WriteString(fmt.Sprintf("        %s : exact;\n", k))
		}
		b.WriteString("    }\n    actions = {\n")
		for _, a := range t.Actions {
			b.WriteString(fmt.Sprintf("        %s;\n", a))
		}
		b.WriteString(fmt.Sprintf("    }\n    size = %d;\n}\n", t.Size))
	}
	return b.String()
}

// Describe returns a program description.
func (p *Program) Describe() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("P4 Program: %s\n", p.Protocol))
	for _, h := range p.Headers {
		b.WriteString(fmt.Sprintf("  header %s (%d fields)\n", h.Name, len(h.Fields)))
	}
	for _, t := range p.Tables {
		b.WriteString(fmt.Sprintf("  table %s (size=%d)\n", t.Name, t.Size))
	}
	return b.String()
}
