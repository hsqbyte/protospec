// Package annotation provides protocol annotation system (@deprecated, @since, etc).
package annotation

import (
	"fmt"
	"strings"
)

// Annotation represents a protocol annotation.
type Annotation struct {
	Type  string `json:"type"`  // "deprecated", "since", "see", "experimental", "custom"
	Field string `json:"field"` // field name, or "" for protocol-level
	Value string `json:"value"`
}

// AnnotationSet holds annotations for a protocol.
type AnnotationSet struct {
	Protocol    string       `json:"protocol"`
	Annotations []Annotation `json:"annotations"`
}

// NewAnnotationSet creates a new annotation set.
func NewAnnotationSet(protocol string) *AnnotationSet {
	return &AnnotationSet{Protocol: protocol}
}

// Add adds an annotation.
func (as *AnnotationSet) Add(a Annotation) {
	as.Annotations = append(as.Annotations, a)
}

// ByType returns annotations filtered by type.
func (as *AnnotationSet) ByType(t string) []Annotation {
	var result []Annotation
	for _, a := range as.Annotations {
		if a.Type == t {
			result = append(result, a)
		}
	}
	return result
}

// ByField returns annotations for a specific field.
func (as *AnnotationSet) ByField(field string) []Annotation {
	var result []Annotation
	for _, a := range as.Annotations {
		if a.Field == field {
			result = append(result, a)
		}
	}
	return result
}

// Deprecated returns all deprecated annotations.
func (as *AnnotationSet) Deprecated() []Annotation {
	return as.ByType("deprecated")
}

// Format formats annotations for display.
func (as *AnnotationSet) Format() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Annotations for %s:\n", as.Protocol))
	for _, a := range as.Annotations {
		target := "protocol"
		if a.Field != "" {
			target = a.Field
		}
		b.WriteString(fmt.Sprintf("  @%s(%s) on %s\n", a.Type, a.Value, target))
	}
	return b.String()
}

// Stats returns annotation statistics.
func (as *AnnotationSet) Stats() map[string]int {
	stats := make(map[string]int)
	for _, a := range as.Annotations {
		stats[a.Type]++
	}
	return stats
}
