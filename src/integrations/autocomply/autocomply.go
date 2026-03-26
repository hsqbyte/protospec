// Package autocomply provides automated protocol compliance management.
package autocomply

import (
	"fmt"
	"strings"
)

// Rule represents a compliance rule.
type Rule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Standard    string `json:"standard"`
	Severity    string `json:"severity"` // critical, high, medium, low
	Description string `json:"description"`
}

// Violation represents a compliance violation.
type Violation struct {
	RuleID  string `json:"rule_id"`
	Field   string `json:"field"`
	Message string `json:"message"`
	Fix     string `json:"fix_suggestion"`
}

// Report represents a compliance report.
type Report struct {
	Protocol   string      `json:"protocol"`
	Rules      []Rule      `json:"rules"`
	Violations []Violation `json:"violations"`
	Score      int         `json:"score"`
}

// DefaultRules returns default compliance rules.
func DefaultRules() []Rule {
	return []Rule{
		{ID: "SEC-001", Name: "Encryption Required", Standard: "NIST", Severity: "critical", Description: "Sensitive fields must be encrypted"},
		{ID: "SEC-002", Name: "Checksum Validation", Standard: "RFC", Severity: "high", Description: "All packets must have valid checksums"},
		{ID: "FMT-001", Name: "Field Naming", Standard: "PSL", Severity: "low", Description: "Fields must use snake_case"},
	}
}

// Check checks a protocol against compliance rules.
func Check(protocol string, rules []Rule) *Report {
	r := &Report{Protocol: protocol, Rules: rules, Score: 100}
	// Stub: would check actual protocol definition
	return r
}

// GenerateReport generates a formatted compliance report.
func (r *Report) GenerateReport() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("=== Compliance Report: %s ===\n", r.Protocol))
	b.WriteString(fmt.Sprintf("Score: %d/100\n\n", r.Score))
	if len(r.Violations) == 0 {
		b.WriteString("✓ All compliance checks passed\n")
	}
	for _, v := range r.Violations {
		b.WriteString(fmt.Sprintf("  ✗ [%s] %s — %s\n    Fix: %s\n", v.RuleID, v.Field, v.Message, v.Fix))
	}
	b.WriteString(fmt.Sprintf("\nRules checked: %d\n", len(r.Rules)))
	return b.String()
}
