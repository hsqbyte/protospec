package format

import (
	"github.com/hsqbyte/protospec/src/errors"
)

// DisplayFormatter converts between raw values and human-readable strings.
type DisplayFormatter interface {
	// Encode converts a raw value to a human-readable string.
	Encode(value any) (string, error)
	// Decode converts a human-readable string back to a raw value.
	Decode(display string) (any, error)
}

// FormatRegistry manages registered display format converters.
type FormatRegistry struct {
	formatters map[string]DisplayFormatter
}

// NewFormatRegistry creates a new empty FormatRegistry.
func NewFormatRegistry() *FormatRegistry {
	return &FormatRegistry{
		formatters: make(map[string]DisplayFormatter),
	}
}

// Register adds a display format converter to the registry.
func (r *FormatRegistry) Register(name string, formatter DisplayFormatter) {
	r.formatters[name] = formatter
}

// Get retrieves a display format converter by name.
// Returns FormatNotFoundError if the format is not registered.
func (r *FormatRegistry) Get(name string) (DisplayFormatter, error) {
	f, ok := r.formatters[name]
	if !ok {
		return nil, &errors.FormatNotFoundError{Name: name}
	}
	return f, nil
}

// Has checks whether a display format is registered.
func (r *FormatRegistry) Has(name string) bool {
	_, ok := r.formatters[name]
	return ok
}
