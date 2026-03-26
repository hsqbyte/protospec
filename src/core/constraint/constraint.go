// Package constraint provides PSL constraint language for field relationships.
package constraint

import (
	"fmt"
	"strings"
)

// ConstraintType represents the type of constraint.
type ConstraintType string

const (
	Arithmetic ConstraintType = "arithmetic"
	Reference  ConstraintType = "reference"
	Range      ConstraintType = "range_check"
	Custom     ConstraintType = "custom"
)

// Constraint defines a field constraint.
type Constraint struct {
	Type       ConstraintType `json:"type"`
	Expression string         `json:"expression"`
	Fields     []string       `json:"fields"`
	Message    string         `json:"message"`
}

// ConstraintBlock holds constraints for a protocol.
type ConstraintBlock struct {
	Protocol    string       `json:"protocol"`
	Constraints []Constraint `json:"constraints"`
}

// NewBlock creates a new constraint block.
func NewBlock(protocol string) *ConstraintBlock {
	return &ConstraintBlock{Protocol: protocol}
}

// AddArithmetic adds an arithmetic constraint.
func (cb *ConstraintBlock) AddArithmetic(expr string, fields []string) *ConstraintBlock {
	cb.Constraints = append(cb.Constraints, Constraint{
		Type:       Arithmetic,
		Expression: expr,
		Fields:     fields,
		Message:    fmt.Sprintf("arithmetic constraint failed: %s", expr),
	})
	return cb
}

// AddReference adds a reference constraint (e.g., checksum).
func (cb *ConstraintBlock) AddReference(expr string, fields []string) *ConstraintBlock {
	cb.Constraints = append(cb.Constraints, Constraint{
		Type:       Reference,
		Expression: expr,
		Fields:     fields,
		Message:    fmt.Sprintf("reference constraint failed: %s", expr),
	})
	return cb
}

// Validate validates field values against constraints.
func (cb *ConstraintBlock) Validate(values map[string]int64) []error {
	var errs []error
	for _, c := range cb.Constraints {
		switch c.Type {
		case Arithmetic:
			// Stub: would evaluate arithmetic expression
			_ = c.Expression
		case Reference:
			_ = c.Expression
		}
	}
	return errs
}

// Describe returns a human-readable description.
func (cb *ConstraintBlock) Describe() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("constraint block for %s:\n", cb.Protocol))
	for _, c := range cb.Constraints {
		b.WriteString(fmt.Sprintf("  [%s] %s (fields: %s)\n", c.Type, c.Expression, strings.Join(c.Fields, ", ")))
	}
	return b.String()
}
