package errors

import "fmt"

// PDLSyntaxError represents a syntax error encountered during PDL parsing.
type PDLSyntaxError struct {
	Line    int
	Column  int
	Message string
	Source  string // optional: the source line for context
}

func (e *PDLSyntaxError) Error() string {
	base := fmt.Sprintf("syntax error at line %d, column %d: %s", e.Line, e.Column, e.Message)
	if e.Source != "" {
		pointer := ""
		if e.Column > 0 {
			for i := 1; i < e.Column; i++ {
				pointer += " "
			}
			pointer += "^"
		}
		base += fmt.Sprintf("\n  %s\n  %s", e.Source, pointer)
	}
	return base
}

// PDLSemanticError represents a semantic error in a PDL definition.
type PDLSemanticError struct {
	FieldName string
	Message   string
}

func (e *PDLSemanticError) Error() string {
	return fmt.Sprintf("semantic error for field %q: %s", e.FieldName, e.Message)
}

// InvalidFieldError indicates a field value is outside its valid range.
type InvalidFieldError struct {
	FieldName   string
	ValidRange  string
	ActualValue any
}

func (e *InvalidFieldError) Error() string {
	return fmt.Sprintf("invalid value for field %q: got %v, valid range is %s", e.FieldName, e.ActualValue, e.ValidRange)
}

// ChecksumError indicates a checksum verification failure.
type ChecksumError struct {
	FieldName string
	Expected  uint64
	Actual    uint64
}

func (e *ChecksumError) Error() string {
	return fmt.Sprintf("checksum mismatch for field %q: expected 0x%X, got 0x%X", e.FieldName, e.Expected, e.Actual)
}

// InsufficientDataError indicates not enough bytes to decode.
type InsufficientDataError struct {
	ExpectedMin int
	ActualLen   int
}

func (e *InsufficientDataError) Error() string {
	return fmt.Sprintf("insufficient data: expected at least %d bytes, got %d", e.ExpectedMin, e.ActualLen)
}

// ProtocolNotFoundError indicates a protocol name was not found in the registry.
type ProtocolNotFoundError struct {
	Name string
}

func (e *ProtocolNotFoundError) Error() string {
	return fmt.Sprintf("protocol not found: %q", e.Name)
}

// ProtocolConflictError indicates a protocol name is already registered.
type ProtocolConflictError struct {
	Name string
}

func (e *ProtocolConflictError) Error() string {
	return fmt.Sprintf("protocol conflict: %q is already registered", e.Name)
}

// AlgorithmNotFoundError indicates a checksum algorithm was not found.
type AlgorithmNotFoundError struct {
	Name string
}

func (e *AlgorithmNotFoundError) Error() string {
	return fmt.Sprintf("checksum algorithm not found: %q", e.Name)
}

// FormatNotFoundError indicates a display format was not found.
type FormatNotFoundError struct {
	Name string
}

func (e *FormatNotFoundError) Error() string {
	return fmt.Sprintf("display format not found: %q", e.Name)
}
