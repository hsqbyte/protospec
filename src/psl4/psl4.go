// Package psl4 provides PSL 4.0 language features: dependent types, effects, macros, type inference.
package psl4

import (
	"fmt"
	"strings"
)

// DependentType represents a dependent type where field type depends on another field's value.
type DependentType struct {
	Field     string            `json:"field"`
	DependsOn string            `json:"depends_on"`
	TypeMap   map[string]string `json:"type_map"` // value → type
}

// Effect represents a side effect annotation.
type Effect struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Pure        bool   `json:"pure"`
}

// Macro represents a compile-time macro.
type Macro struct {
	Name       string   `json:"name"`
	Parameters []string `json:"parameters"`
	Body       string   `json:"body"`
}

// TypeInference represents a type inference result.
type TypeInference struct {
	Field        string  `json:"field"`
	InferredType string  `json:"inferred_type"`
	Confidence   float64 `json:"confidence"`
}

// PSL4Features holds all PSL 4.0 language features.
type PSL4Features struct {
	DependentTypes []DependentType `json:"dependent_types"`
	Effects        []Effect        `json:"effects"`
	Macros         []Macro         `json:"macros"`
}

// NewFeatures creates PSL 4.0 features with examples.
func NewFeatures() *PSL4Features {
	return &PSL4Features{
		DependentTypes: []DependentType{
			{Field: "payload", DependsOn: "type", TypeMap: map[string]string{"1": "TCPPayload", "2": "UDPPayload", "3": "ICMPPayload"}},
		},
		Effects: []Effect{
			{Name: "pure", Description: "No side effects", Pure: true},
			{Name: "io", Description: "Performs I/O operations", Pure: false},
			{Name: "network", Description: "Network access", Pure: false},
		},
		Macros: []Macro{
			{Name: "checksum_field", Parameters: []string{"algo", "scope"}, Body: "field checksum: uint16 = ${algo}(${scope});"},
			{Name: "length_prefixed", Parameters: []string{"type"}, Body: "field length: uint16;\nfield data: ${type}[length];"},
		},
	}
}

// InferTypes performs type inference on field definitions.
func InferTypes(fields map[string]string) []TypeInference {
	var results []TypeInference
	for name, hint := range fields {
		inferred := "uint32"
		confidence := 0.8
		if strings.Contains(name, "port") {
			inferred = "uint16"
			confidence = 0.95
		} else if strings.Contains(name, "addr") || strings.Contains(name, "ip") {
			inferred = "bytes"
			confidence = 0.9
		} else if hint != "" {
			inferred = hint
			confidence = 1.0
		}
		results = append(results, TypeInference{Field: name, InferredType: inferred, Confidence: confidence})
	}
	return results
}

// Describe returns a description of PSL 4.0 features.
func (f *PSL4Features) Describe() string {
	var b strings.Builder
	b.WriteString("=== PSL 4.0 Language Features ===\n\n")
	b.WriteString("Dependent Types:\n")
	for _, dt := range f.DependentTypes {
		b.WriteString(fmt.Sprintf("  %s depends on %s (%d variants)\n", dt.Field, dt.DependsOn, len(dt.TypeMap)))
	}
	b.WriteString("\nEffect System:\n")
	for _, e := range f.Effects {
		pure := "impure"
		if e.Pure {
			pure = "pure"
		}
		b.WriteString(fmt.Sprintf("  @%s [%s] — %s\n", e.Name, pure, e.Description))
	}
	b.WriteString("\nMacros:\n")
	for _, m := range f.Macros {
		b.WriteString(fmt.Sprintf("  macro %s(%s)\n", m.Name, strings.Join(m.Parameters, ", ")))
	}
	return b.String()
}
