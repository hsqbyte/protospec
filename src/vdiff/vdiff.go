// Package vdiff provides visual protocol diff tools.
package vdiff

import (
	"fmt"
	"strings"
)

// FieldDiff represents a difference in a single field.
type FieldDiff struct {
	Name    string `json:"name"`
	Type    string `json:"type"` // added, removed, changed, unchanged
	OldType string `json:"old_type,omitempty"`
	NewType string `json:"new_type,omitempty"`
	OldBits int    `json:"old_bits,omitempty"`
	NewBits int    `json:"new_bits,omitempty"`
}

// ProtocolDiff represents the diff between two protocol versions.
type ProtocolDiff struct {
	ProtocolA string      `json:"protocol_a"`
	ProtocolB string      `json:"protocol_b"`
	Fields    []FieldDiff `json:"fields"`
}

// Compare compares two field lists and produces a diff.
func Compare(nameA, nameB string, fieldsA, fieldsB []string) *ProtocolDiff {
	diff := &ProtocolDiff{ProtocolA: nameA, ProtocolB: nameB}
	setB := make(map[string]bool)
	for _, f := range fieldsB {
		setB[f] = true
	}
	setA := make(map[string]bool)
	for _, f := range fieldsA {
		setA[f] = true
		if setB[f] {
			diff.Fields = append(diff.Fields, FieldDiff{Name: f, Type: "unchanged"})
		} else {
			diff.Fields = append(diff.Fields, FieldDiff{Name: f, Type: "removed"})
		}
	}
	for _, f := range fieldsB {
		if !setA[f] {
			diff.Fields = append(diff.Fields, FieldDiff{Name: f, Type: "added"})
		}
	}
	return diff
}

// ToHTML generates an HTML side-by-side diff view.
func (d *ProtocolDiff) ToHTML() string {
	var b strings.Builder
	b.WriteString("<html><head><style>")
	b.WriteString(".added{background:#d4edda} .removed{background:#f8d7da} .changed{background:#fff3cd} table{border-collapse:collapse;width:100%} td,th{border:1px solid #ddd;padding:8px}")
	b.WriteString("</style></head><body>")
	b.WriteString(fmt.Sprintf("<h2>%s vs %s</h2>", d.ProtocolA, d.ProtocolB))
	b.WriteString("<table><tr><th>Field</th><th>Status</th></tr>")
	for _, f := range d.Fields {
		cls := f.Type
		b.WriteString(fmt.Sprintf("<tr class=\"%s\"><td>%s</td><td>%s</td></tr>", cls, f.Name, f.Type))
	}
	b.WriteString("</table></body></html>")
	return b.String()
}

// ToText generates a text diff view.
func (d *ProtocolDiff) ToText() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("--- %s\n+++ %s\n", d.ProtocolA, d.ProtocolB))
	for _, f := range d.Fields {
		switch f.Type {
		case "added":
			b.WriteString(fmt.Sprintf("+ %s\n", f.Name))
		case "removed":
			b.WriteString(fmt.Sprintf("- %s\n", f.Name))
		case "changed":
			b.WriteString(fmt.Sprintf("~ %s\n", f.Name))
		default:
			b.WriteString(fmt.Sprintf("  %s\n", f.Name))
		}
	}
	return b.String()
}

// Summary returns a summary of changes.
func (d *ProtocolDiff) Summary() string {
	added, removed, changed := 0, 0, 0
	for _, f := range d.Fields {
		switch f.Type {
		case "added":
			added++
		case "removed":
			removed++
		case "changed":
			changed++
		}
	}
	return fmt.Sprintf("%d added, %d removed, %d changed", added, removed, changed)
}
