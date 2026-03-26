// Package coverage provides protocol field coverage analysis.
package coverage

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FieldCoverage tracks coverage for a single field.
type FieldCoverage struct {
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	ObservedVals []string `json:"observed_values"`
	TotalValues  int      `json:"total_values,omitempty"` // for enums
	Covered      bool     `json:"covered"`
}

// Report holds coverage analysis results.
type Report struct {
	Protocol string          `json:"protocol"`
	Fields   []FieldCoverage `json:"fields"`
	Total    int             `json:"total_fields"`
	Covered  int             `json:"covered_fields"`
	Percent  float64         `json:"coverage_percent"`
}

// NewReport creates a new coverage report.
func NewReport(protocol string) *Report {
	return &Report{Protocol: protocol}
}

// AddField adds a field coverage entry.
func (r *Report) AddField(name, typ string, observed []string, totalEnum int) {
	fc := FieldCoverage{
		Name:         name,
		Type:         typ,
		ObservedVals: observed,
		TotalValues:  totalEnum,
		Covered:      len(observed) > 0,
	}
	r.Fields = append(r.Fields, fc)
	r.Total++
	if fc.Covered {
		r.Covered++
	}
	if r.Total > 0 {
		r.Percent = float64(r.Covered) / float64(r.Total) * 100
	}
}

// ToJSON returns the report as JSON.
func (r *Report) ToJSON() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}

// ToHTML returns the report as HTML.
func (r *Report) ToHTML() string {
	var b strings.Builder
	b.WriteString("<html><head><style>")
	b.WriteString(".covered{background:#d4edda} .uncovered{background:#f8d7da} table{border-collapse:collapse;width:100%} td,th{border:1px solid #ddd;padding:8px}")
	b.WriteString("</style></head><body>")
	b.WriteString(fmt.Sprintf("<h2>%s Coverage: %.1f%%</h2>", r.Protocol, r.Percent))
	b.WriteString("<table><tr><th>Field</th><th>Type</th><th>Observed</th><th>Status</th></tr>")
	for _, f := range r.Fields {
		cls := "uncovered"
		if f.Covered {
			cls = "covered"
		}
		b.WriteString(fmt.Sprintf("<tr class=\"%s\"><td>%s</td><td>%s</td><td>%d</td><td>%s</td></tr>",
			cls, f.Name, f.Type, len(f.ObservedVals), cls))
	}
	b.WriteString("</table></body></html>")
	return b.String()
}

// ToText returns a text summary.
func (r *Report) ToText() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Protocol: %s — Coverage: %.1f%% (%d/%d fields)\n\n", r.Protocol, r.Percent, r.Covered, r.Total))
	for _, f := range r.Fields {
		status := "✓"
		if !f.Covered {
			status = "✗"
		}
		b.WriteString(fmt.Sprintf("  %s %s (%s) — %d values observed\n", status, f.Name, f.Type, len(f.ObservedVals)))
	}
	return b.String()
}
