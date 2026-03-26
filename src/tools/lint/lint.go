// Package lint provides PSL code style checking.
package lint

import (
	"fmt"
	"io/fs"
	"github.com/hsqbyte/protospec/src/core/pdl"
	"strings"
	"unicode"
)

// Severity represents lint issue severity.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// Issue represents a lint issue.
type Issue struct {
	File     string   `json:"file"`
	Line     int      `json:"line"`
	Severity Severity `json:"severity"`
	Rule     string   `json:"rule"`
	Message  string   `json:"message"`
}

// Rule represents a lint rule.
type Rule struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

// Config holds linter configuration.
type Config struct {
	Rules []Rule `json:"rules"`
}

// DefaultConfig returns the default linter configuration.
func DefaultConfig() *Config {
	return &Config{
		Rules: []Rule{
			{Name: "snake_case_fields", Enabled: true},
			{Name: "unused_constants", Enabled: true},
			{Name: "field_order", Enabled: true},
			{Name: "naming_convention", Enabled: true},
		},
	}
}

// Linter performs PSL code style checks.
type Linter struct {
	Config          *Config
	TransportLoader *pdl.TransportLoader
}

// NewLinter creates a new linter.
func NewLinter(cfg *Config) *Linter {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Linter{Config: cfg}
}

// NewLinterWithTransport creates a new linter with transport support.
func NewLinterWithTransport(cfg *Config, embedFS fs.FS) *Linter {
	l := NewLinter(cfg)
	if embedFS != nil {
		l.TransportLoader = pdl.NewTransportLoader(embedFS, nil)
	}
	return l
}

// CheckSnakeCase checks if a name follows snake_case convention.
func CheckSnakeCase(name string) bool {
	for _, r := range name {
		if unicode.IsUpper(r) {
			return false
		}
	}
	return true
}

// LintContent lints PSL content and returns issues.
func (l *Linter) LintContent(filename, content string) []Issue {
	var issues []Issue
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Check field naming
		if strings.Contains(trimmed, ":") && !strings.HasPrefix(trimmed, "//") && !strings.HasPrefix(trimmed, "protocol") {
			parts := strings.SplitN(trimmed, ":", 2)
			fieldName := strings.TrimSpace(parts[0])
			// Strip "field " prefix if present
			fieldName = strings.TrimPrefix(fieldName, "field ")
			if !CheckSnakeCase(fieldName) && l.ruleEnabled("snake_case_fields") {
				issues = append(issues, Issue{
					File:     filename,
					Line:     i + 1,
					Severity: SeverityWarning,
					Rule:     "snake_case_fields",
					Message:  fmt.Sprintf("field '%s' should use snake_case", fieldName),
				})
			}
		}
	}

	// Transport-aware validation: parse message protocols with transport loader
	if l.TransportLoader != nil && strings.Contains(content, "message ") && strings.Contains(content, "transport ") {
		parser := pdl.NewPDLParser(nil, nil)
		parser.SetTransportLoader(l.TransportLoader)
		_, err := parser.ParseMessage(content)
		if err != nil {
			issues = append(issues, Issue{
				File:     filename,
				Line:     1,
				Severity: SeverityError,
				Rule:     "transport_validation",
				Message:  err.Error(),
			})
		}
	}

	return issues
}

func (l *Linter) ruleEnabled(name string) bool {
	for _, r := range l.Config.Rules {
		if r.Name == name {
			return r.Enabled
		}
	}
	return false
}

// FormatIssues formats lint issues for display.
func FormatIssues(issues []Issue) string {
	var b strings.Builder
	for _, iss := range issues {
		b.WriteString(fmt.Sprintf("%s:%d [%s] %s — %s\n", iss.File, iss.Line, iss.Severity, iss.Rule, iss.Message))
	}
	if len(issues) == 0 {
		b.WriteString("no issues found\n")
	} else {
		b.WriteString(fmt.Sprintf("\n%d issue(s) found\n", len(issues)))
	}
	return b.String()
}
