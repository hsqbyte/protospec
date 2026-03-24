// Package compliance provides RFC compliance testing for protocols.
package compliance

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hsqbyte/protospec/src/protocol"
	"github.com/hsqbyte/protospec/src/schema"
)

// TestLevel represents the compliance test result level.
type TestLevel string

const (
	Pass TestLevel = "PASS"
	Warn TestLevel = "WARN"
	Fail TestLevel = "FAIL"
)

// TestCase defines a single compliance test.
type TestCase struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"` // "required_fields", "value_range", "state_machine"
	Protocol    string `json:"protocol"`
}

// TestResult holds the result of a compliance test.
type TestResult struct {
	TestCase TestCase  `json:"test_case"`
	Level    TestLevel `json:"level"`
	Message  string    `json:"message"`
	Duration float64   `json:"duration_ms"`
}

// ComplianceReport holds the full compliance report.
type ComplianceReport struct {
	Protocol  string       `json:"protocol"`
	Timestamp time.Time    `json:"timestamp"`
	Results   []TestResult `json:"results"`
	Summary   Summary      `json:"summary"`
}

// Summary holds pass/warn/fail counts.
type Summary struct {
	Total  int `json:"total"`
	Passed int `json:"passed"`
	Warned int `json:"warned"`
	Failed int `json:"failed"`
}

// Runner executes compliance tests.
type Runner struct {
	lib *protocol.Library
}

// NewRunner creates a new compliance test runner.
func NewRunner(lib *protocol.Library) *Runner {
	return &Runner{lib: lib}
}

// RunCompliance runs compliance tests for a protocol against sample data.
func (r *Runner) RunCompliance(protoName string, samples [][]byte) (*ComplianceReport, error) {
	report := &ComplianceReport{
		Protocol:  protoName,
		Timestamp: time.Now(),
	}

	// Check if message protocol
	if ms := r.lib.Message(protoName); ms != nil {
		r.runMessageCompliance(ms, report)
		r.summarize(report)
		return report, nil
	}

	s, err := r.lib.Registry().GetSchema(protoName)
	if err != nil {
		return nil, err
	}

	// Required fields check
	r.checkRequiredFields(s, samples, report)

	// Value range check
	r.checkValueRanges(s, samples, report)

	// Decode validity check
	r.checkDecodeValidity(protoName, samples, report)

	r.summarize(report)
	return report, nil
}

func (r *Runner) checkRequiredFields(s *schema.ProtocolSchema, samples [][]byte, report *ComplianceReport) {
	tc := TestCase{
		Name:        "required_fields",
		Description: "All required fields must be present in decoded output",
		Category:    "required_fields",
		Protocol:    s.Name,
	}

	start := time.Now()
	allPresent := true
	var missing []string

	for _, data := range samples {
		result, err := r.lib.Decode(s.Name, data)
		if err != nil {
			continue
		}
		for _, f := range flatFields(s.Fields) {
			if f.Condition != nil {
				continue // skip conditional fields
			}
			if _, ok := result.Packet[f.Name]; !ok {
				allPresent = false
				missing = append(missing, f.Name)
			}
		}
	}

	level := Pass
	msg := "All required fields present"
	if !allPresent {
		level = Fail
		msg = fmt.Sprintf("Missing fields: %s", strings.Join(missing, ", "))
	}
	if len(samples) == 0 {
		level = Warn
		msg = "No sample data provided"
	}

	report.Results = append(report.Results, TestResult{
		TestCase: tc,
		Level:    level,
		Message:  msg,
		Duration: float64(time.Since(start).Microseconds()) / 1000,
	})
}

func (r *Runner) checkValueRanges(s *schema.ProtocolSchema, samples [][]byte, report *ComplianceReport) {
	tc := TestCase{
		Name:        "value_ranges",
		Description: "Field values must be within defined ranges",
		Category:    "value_range",
		Protocol:    s.Name,
	}

	start := time.Now()
	violations := 0

	for _, data := range samples {
		result, err := r.lib.Decode(s.Name, data)
		if err != nil {
			continue
		}
		for _, f := range flatFields(s.Fields) {
			if f.RangeMin == nil && f.RangeMax == nil {
				continue
			}
			val, ok := result.Packet[f.Name]
			if !ok {
				continue
			}
			if v, ok := val.(uint64); ok {
				if f.RangeMin != nil && int64(v) < *f.RangeMin {
					violations++
				}
				if f.RangeMax != nil && int64(v) > *f.RangeMax {
					violations++
				}
			}
		}
	}

	level := Pass
	msg := "All field values within range"
	if violations > 0 {
		level = Fail
		msg = fmt.Sprintf("%d range violations found", violations)
	}

	report.Results = append(report.Results, TestResult{
		TestCase: tc,
		Level:    level,
		Message:  msg,
		Duration: float64(time.Since(start).Microseconds()) / 1000,
	})
}

func (r *Runner) checkDecodeValidity(protoName string, samples [][]byte, report *ComplianceReport) {
	tc := TestCase{
		Name:        "decode_validity",
		Description: "All samples must decode without errors",
		Category:    "decode",
		Protocol:    protoName,
	}

	start := time.Now()
	errors := 0
	total := len(samples)

	for _, data := range samples {
		_, err := r.lib.Decode(protoName, data)
		if err != nil {
			errors++
		}
	}

	level := Pass
	msg := fmt.Sprintf("%d/%d samples decoded successfully", total-errors, total)
	if errors > 0 {
		level = Warn
		msg = fmt.Sprintf("%d/%d decode errors", errors, total)
	}

	report.Results = append(report.Results, TestResult{
		TestCase: tc,
		Level:    level,
		Message:  msg,
		Duration: float64(time.Since(start).Microseconds()) / 1000,
	})
}

func (r *Runner) runMessageCompliance(ms *schema.MessageSchema, report *ComplianceReport) {
	tc := TestCase{
		Name:        "message_structure",
		Description: "Message protocol structure validation",
		Category:    "structure",
		Protocol:    ms.Name,
	}

	hasRequest := false
	hasResponse := false
	for _, msg := range ms.Messages {
		if msg.Kind == "request" {
			hasRequest = true
		}
		if msg.Kind == "response" {
			hasResponse = true
		}
	}

	level := Pass
	msg := fmt.Sprintf("%d messages defined", len(ms.Messages))
	if !hasRequest && !hasResponse {
		level = Warn
		msg = "No request or response messages defined"
	}

	report.Results = append(report.Results, TestResult{
		TestCase: tc,
		Level:    level,
		Message:  msg,
	})
}

func (r *Runner) summarize(report *ComplianceReport) {
	for _, res := range report.Results {
		report.Summary.Total++
		switch res.Level {
		case Pass:
			report.Summary.Passed++
		case Warn:
			report.Summary.Warned++
		case Fail:
			report.Summary.Failed++
		}
	}
}

// ToHTML generates an HTML compliance report.
func (report *ComplianceReport) ToHTML() string {
	var b strings.Builder
	b.WriteString("<!DOCTYPE html>\n<html><head><meta charset=\"utf-8\">\n")
	b.WriteString(fmt.Sprintf("<title>Compliance Report — %s</title>\n", report.Protocol))
	b.WriteString("<style>body{font-family:sans-serif;max-width:800px;margin:0 auto;padding:20px}")
	b.WriteString(".pass{color:green}.warn{color:orange}.fail{color:red}")
	b.WriteString("table{border-collapse:collapse;width:100%}th,td{border:1px solid #ddd;padding:8px;text-align:left}th{background:#f5f5f5}</style>\n")
	b.WriteString("</head><body>\n")
	b.WriteString(fmt.Sprintf("<h1>Compliance Report: %s</h1>\n", report.Protocol))
	b.WriteString(fmt.Sprintf("<p>Date: %s</p>\n", report.Timestamp.Format(time.RFC3339)))
	b.WriteString(fmt.Sprintf("<p>Total: %d | <span class=\"pass\">PASS: %d</span> | <span class=\"warn\">WARN: %d</span> | <span class=\"fail\">FAIL: %d</span></p>\n",
		report.Summary.Total, report.Summary.Passed, report.Summary.Warned, report.Summary.Failed))
	b.WriteString("<table><tr><th>Test</th><th>Category</th><th>Result</th><th>Message</th></tr>\n")
	for _, res := range report.Results {
		cls := strings.ToLower(string(res.Level))
		b.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%s</td><td class=\"%s\">%s</td><td>%s</td></tr>\n",
			res.TestCase.Name, res.TestCase.Category, cls, res.Level, res.Message))
	}
	b.WriteString("</table></body></html>\n")
	return b.String()
}

// ToJSON generates a JSON compliance report.
func (report *ComplianceReport) ToJSON() (string, error) {
	data, err := json.MarshalIndent(report, "", "  ")
	return string(data), err
}

func flatFields(fields []schema.FieldDef) []schema.FieldDef {
	var out []schema.FieldDef
	for _, f := range fields {
		if f.IsBitfieldGroup {
			out = append(out, f.BitfieldFields...)
		} else {
			out = append(out, f)
		}
	}
	return out
}
