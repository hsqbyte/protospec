// Package optimize provides protocol design analysis and optimization suggestions.
package optimize

import (
	"fmt"
	"math"
	"strings"

	"github.com/hsqbyte/protospec/src/core/schema"
)

// Suggestion represents an optimization suggestion.
type Suggestion struct {
	Field    string `json:"field"`
	Type     string `json:"type"`     // "alignment", "redundancy", "bitwidth", "compression"
	Severity string `json:"severity"` // "info", "warning", "error"
	Message  string `json:"message"`
}

// Analyzer analyzes protocol schemas for optimization opportunities.
type Analyzer struct{}

// NewAnalyzer creates a new optimization analyzer.
func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

// Analyze analyzes a protocol schema and returns suggestions.
func (a *Analyzer) Analyze(s *schema.ProtocolSchema) []Suggestion {
	var suggestions []Suggestion
	suggestions = append(suggestions, a.checkAlignment(s)...)
	suggestions = append(suggestions, a.checkBitwidth(s)...)
	suggestions = append(suggestions, a.checkRedundancy(s)...)
	return suggestions
}

func (a *Analyzer) checkAlignment(s *schema.ProtocolSchema) []Suggestion {
	var suggestions []Suggestion
	bitOffset := 0
	for _, f := range s.Fields {
		if f.IsBitfieldGroup {
			for _, bf := range f.BitfieldFields {
				bitOffset += bf.BitWidth
			}
			continue
		}
		if f.BitWidth > 0 && f.BitWidth >= 8 && bitOffset%8 != 0 {
			suggestions = append(suggestions, Suggestion{
				Field:    f.Name,
				Type:     "alignment",
				Severity: "warning",
				Message:  fmt.Sprintf("field %q (%d bits) is not byte-aligned at bit offset %d", f.Name, f.BitWidth, bitOffset),
			})
		}
		bitOffset += f.BitWidth
	}
	return suggestions
}

func (a *Analyzer) checkBitwidth(s *schema.ProtocolSchema) []Suggestion {
	var suggestions []Suggestion
	for _, f := range s.Fields {
		if f.IsBitfieldGroup {
			for _, bf := range f.BitfieldFields {
				if bf.RangeMax != nil {
					needed := int(math.Ceil(math.Log2(float64(*bf.RangeMax + 1))))
					if needed < bf.BitWidth {
						suggestions = append(suggestions, Suggestion{
							Field:    bf.Name,
							Type:     "bitwidth",
							Severity: "info",
							Message:  fmt.Sprintf("field %q uses %d bits but max value %d only needs %d bits", bf.Name, bf.BitWidth, *bf.RangeMax, needed),
						})
					}
				}
			}
		}
		if f.EnumMap != nil && f.BitWidth > 0 {
			maxEnum := 0
			for k := range f.EnumMap {
				if k > maxEnum {
					maxEnum = k
				}
			}
			needed := int(math.Ceil(math.Log2(float64(maxEnum + 1))))
			if needed > 0 && needed < f.BitWidth-8 {
				suggestions = append(suggestions, Suggestion{
					Field:    f.Name,
					Type:     "bitwidth",
					Severity: "info",
					Message:  fmt.Sprintf("field %q uses %d bits but enum max %d only needs %d bits", f.Name, f.BitWidth, maxEnum, needed),
				})
			}
		}
	}
	return suggestions
}

func (a *Analyzer) checkRedundancy(s *schema.ProtocolSchema) []Suggestion {
	var suggestions []Suggestion
	names := make(map[string]int)
	for _, f := range s.Fields {
		if f.IsBitfieldGroup {
			for _, bf := range f.BitfieldFields {
				names[strings.ToLower(bf.Name)]++
			}
		} else {
			names[strings.ToLower(f.Name)]++
		}
	}
	for name, count := range names {
		if count > 1 {
			suggestions = append(suggestions, Suggestion{
				Field:    name,
				Type:     "redundancy",
				Severity: "warning",
				Message:  fmt.Sprintf("field name %q appears %d times", name, count),
			})
		}
	}
	return suggestions
}

// FormatSuggestions formats suggestions as a human-readable string.
func FormatSuggestions(suggestions []Suggestion) string {
	if len(suggestions) == 0 {
		return "No optimization suggestions. Protocol looks good!\n"
	}
	var b strings.Builder
	for _, s := range suggestions {
		icon := "ℹ"
		if s.Severity == "warning" {
			icon = "⚠"
		} else if s.Severity == "error" {
			icon = "✗"
		}
		b.WriteString(fmt.Sprintf("%s [%s] %s\n", icon, s.Type, s.Message))
	}
	return b.String()
}
