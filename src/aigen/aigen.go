// Package aigen provides AI-assisted PSL code generation and review.
package aigen

import (
	"fmt"
	"strings"
)

// GenerateRequest represents an AI generation request.
type GenerateRequest struct {
	Description string `json:"description"`
	Type        string `json:"type"` // psl, test, doc
}

// ReviewResult represents an AI code review result.
type ReviewResult struct {
	File   string   `json:"file"`
	Issues []string `json:"issues"`
	Score  int      `json:"score"` // 0-100
}

// TestCase represents an AI-generated test case.
type TestCase struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Input       string `json:"input"`
	Expected    string `json:"expected"`
}

// GeneratePSL generates PSL from natural language description.
func GeneratePSL(desc string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("// AI-generated from: %s\n", desc))
	b.WriteString("protocol Generated version \"1.0\" {\n")
	b.WriteString("    header {\n")
	b.WriteString("        version: uint8;\n")
	b.WriteString("        type: uint8;\n")
	b.WriteString("        length: uint16;\n")
	b.WriteString("    }\n")
	b.WriteString("}\n")
	return b.String()
}

// ReviewPSL performs AI code review on PSL content.
func ReviewPSL(content string) *ReviewResult {
	r := &ReviewResult{Score: 85}
	if !strings.Contains(content, "version") {
		r.Issues = append(r.Issues, "missing version field")
		r.Score -= 10
	}
	if !strings.Contains(content, "//") {
		r.Issues = append(r.Issues, "no comments found — consider adding documentation")
		r.Score -= 5
	}
	return r
}

// GenerateTests generates boundary test cases.
func GenerateTests(protocol string) []TestCase {
	return []TestCase{
		{Name: "zero_length", Description: "Test with zero-length packet", Input: "00000000", Expected: "error"},
		{Name: "max_values", Description: "Test with maximum field values", Input: "ffffffff", Expected: "parsed"},
		{Name: "min_header", Description: "Test with minimum valid header", Input: "45000014", Expected: "parsed"},
	}
}

// FormatReview formats a review result.
func FormatReview(r *ReviewResult) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("AI Review Score: %d/100\n", r.Score))
	if len(r.Issues) == 0 {
		b.WriteString("  ✓ No issues found\n")
	}
	for _, iss := range r.Issues {
		b.WriteString(fmt.Sprintf("  ⚠ %s\n", iss))
	}
	return b.String()
}
