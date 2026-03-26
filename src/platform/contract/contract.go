// Package contract provides protocol contract testing.
package contract

import (
	"fmt"
	"strings"
)

// Contract represents a protocol contract between provider and consumer.
type Contract struct {
	Provider string   `json:"provider"`
	Consumer string   `json:"consumer"`
	Fields   []string `json:"required_fields"`
	Version  string   `json:"version"`
}

// Result represents a contract verification result.
type Result struct {
	Compatible bool     `json:"compatible"`
	Missing    []string `json:"missing_fields"`
	Extra      []string `json:"extra_fields"`
}

// Verify verifies compatibility between provider and consumer fields.
func Verify(provider, consumer []string) *Result {
	r := &Result{Compatible: true}
	pSet := make(map[string]bool)
	for _, f := range provider {
		pSet[f] = true
	}
	for _, f := range consumer {
		if !pSet[f] {
			r.Missing = append(r.Missing, f)
			r.Compatible = false
		}
	}
	cSet := make(map[string]bool)
	for _, f := range consumer {
		cSet[f] = true
	}
	for _, f := range provider {
		if !cSet[f] {
			r.Extra = append(r.Extra, f)
		}
	}
	return r
}

// FormatResult formats a verification result.
func FormatResult(c *Contract, r *Result) string {
	var b strings.Builder
	status := "✓ compatible"
	if !r.Compatible {
		status = "✗ incompatible"
	}
	b.WriteString(fmt.Sprintf("Contract: %s → %s [%s]\n", c.Provider, c.Consumer, status))
	if len(r.Missing) > 0 {
		b.WriteString(fmt.Sprintf("  missing: %s\n", strings.Join(r.Missing, ", ")))
	}
	if len(r.Extra) > 0 {
		b.WriteString(fmt.Sprintf("  extra: %s\n", strings.Join(r.Extra, ", ")))
	}
	return b.String()
}
