// Package testgen generates test cases for protocol schemas.
package testgen

import (
	"fmt"
	"math"
	"strings"

	"github.com/hsqbyte/protospec/src/schema"
)

// TestCase represents a generated test case.
type TestCase struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Type        string         `json:"type"` // "boundary", "equivalence", "roundtrip", "property"
	Input       map[string]any `json:"input"`
	Expected    string         `json:"expected"` // "success", "error"
}

// Generator generates test cases from protocol schemas.
type Generator struct{}

// NewGenerator creates a new test case generator.
func NewGenerator() *Generator {
	return &Generator{}
}

// Generate generates test cases for a protocol schema.
func (g *Generator) Generate(s *schema.ProtocolSchema) []TestCase {
	var tests []TestCase
	tests = append(tests, g.boundaryTests(s)...)
	tests = append(tests, g.roundtripTest(s))
	tests = append(tests, g.propertyTests(s)...)
	return tests
}

func (g *Generator) boundaryTests(s *schema.ProtocolSchema) []TestCase {
	var tests []TestCase
	var fields []schema.FieldDef
	for _, f := range s.Fields {
		if f.IsBitfieldGroup {
			fields = append(fields, f.BitfieldFields...)
		} else {
			fields = append(fields, f)
		}
	}

	for _, f := range fields {
		if f.Type == schema.Bytes || f.Type == schema.String {
			continue
		}
		if f.BitWidth == 0 {
			continue
		}

		maxVal := int64(math.Pow(2, float64(f.BitWidth))) - 1

		// Min value test
		input := make(map[string]any)
		input[f.Name] = 0
		tests = append(tests, TestCase{
			Name:        fmt.Sprintf("%s_min", f.Name),
			Description: fmt.Sprintf("Minimum value for %s (0)", f.Name),
			Type:        "boundary",
			Input:       input,
			Expected:    "success",
		})

		// Max value test
		input2 := make(map[string]any)
		input2[f.Name] = maxVal
		tests = append(tests, TestCase{
			Name:        fmt.Sprintf("%s_max", f.Name),
			Description: fmt.Sprintf("Maximum value for %s (%d)", f.Name, maxVal),
			Type:        "boundary",
			Input:       input2,
			Expected:    "success",
		})
	}
	return tests
}

func (g *Generator) roundtripTest(s *schema.ProtocolSchema) TestCase {
	input := make(map[string]any)
	var fields []schema.FieldDef
	for _, f := range s.Fields {
		if f.IsBitfieldGroup {
			fields = append(fields, f.BitfieldFields...)
		} else {
			fields = append(fields, f)
		}
	}

	for _, f := range fields {
		switch f.Type {
		case schema.Uint, schema.Int:
			input[f.Name] = 1
		case schema.Bool:
			input[f.Name] = true
		case schema.Bytes:
			input[f.Name] = []byte{0xDE, 0xAD}
		case schema.String:
			input[f.Name] = "test"
		}
	}

	return TestCase{
		Name:        "roundtrip",
		Description: "Encode then decode should produce identical values",
		Type:        "roundtrip",
		Input:       input,
		Expected:    "success",
	}
}

func (g *Generator) propertyTests(s *schema.ProtocolSchema) []TestCase {
	return []TestCase{
		{
			Name:        "encode_decode_identity",
			Description: "For all valid inputs: decode(encode(x)) == x",
			Type:        "property",
			Expected:    "success",
		},
		{
			Name:        "field_independence",
			Description: "Changing one field does not affect other fields after decode",
			Type:        "property",
			Expected:    "success",
		},
	}
}

// FormatGoTest generates Go test code from test cases.
func FormatGoTest(protoName string, tests []TestCase) string {
	var b strings.Builder
	b.WriteString("package test\n\n")
	b.WriteString("import (\n\t\"testing\"\n)\n\n")

	for _, tc := range tests {
		testName := fmt.Sprintf("Test%s_%s", protoName, tc.Name)
		b.WriteString(fmt.Sprintf("func %s(t *testing.T) {\n", testName))
		b.WriteString(fmt.Sprintf("\t// %s\n", tc.Description))
		b.WriteString(fmt.Sprintf("\t// Type: %s, Expected: %s\n", tc.Type, tc.Expected))
		b.WriteString("\tt.Skip(\"auto-generated test stub\")\n")
		b.WriteString("}\n\n")
	}
	return b.String()
}
