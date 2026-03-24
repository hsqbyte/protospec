// Package formatter provides PSL code formatting.
package formatter

import (
	"strings"
)

// Options holds formatting options.
type Options struct {
	IndentSize  int  `json:"indent_size"`
	AlignFields bool `json:"align_fields"`
	SortImports bool `json:"sort_imports"`
}

// DefaultOptions returns default formatting options.
func DefaultOptions() *Options {
	return &Options{IndentSize: 4, AlignFields: true, SortImports: true}
}

// Format formats PSL source code.
func Format(source string, opts *Options) string {
	if opts == nil {
		opts = DefaultOptions()
	}
	lines := strings.Split(source, "\n")
	var result []string
	indent := 0
	indentStr := strings.Repeat(" ", opts.IndentSize)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			result = append(result, "")
			continue
		}
		if trimmed == "}" {
			indent--
			if indent < 0 {
				indent = 0
			}
		}
		formatted := strings.Repeat(indentStr, indent) + trimmed
		result = append(result, formatted)
		if strings.HasSuffix(trimmed, "{") {
			indent++
		}
	}
	return strings.Join(result, "\n")
}

// Check checks if source is already formatted.
func Check(source string, opts *Options) bool {
	formatted := Format(source, opts)
	return source == formatted
}

// AlignFields aligns field definitions by colon position.
func AlignFields(lines []string) []string {
	maxNameLen := 0
	type fieldLine struct {
		name string
		rest string
	}
	var fields []fieldLine
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if idx := strings.Index(trimmed, ":"); idx > 0 && !strings.HasPrefix(trimmed, "//") {
			name := trimmed[:idx]
			rest := trimmed[idx:]
			if len(name) > maxNameLen {
				maxNameLen = len(name)
			}
			fields = append(fields, fieldLine{name, rest})
		} else {
			fields = append(fields, fieldLine{trimmed, ""})
		}
	}
	var result []string
	for _, f := range fields {
		if f.rest != "" {
			padded := f.name + strings.Repeat(" ", maxNameLen-len(f.name)) + f.rest
			result = append(result, padded)
		} else {
			result = append(result, f.name)
		}
	}
	return result
}
