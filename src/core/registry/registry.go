package registry

import (
	"os"
	"sort"

	"github.com/hsqbyte/protospec/src/core/checksum"
	"github.com/hsqbyte/protospec/src/core/errors"
	"github.com/hsqbyte/protospec/src/core/format"
	"github.com/hsqbyte/protospec/src/core/pdl"
	"github.com/hsqbyte/protospec/src/core/schema"
)

// ProtocolRegistry manages loaded protocol schemas.
type ProtocolRegistry struct {
	parser  *pdl.PDLParser
	schemas map[string]*schema.ProtocolSchema
}

// NewProtocolRegistry creates a new ProtocolRegistry with the given registries
// for semantic validation during PSL parsing.
func NewProtocolRegistry(cr *checksum.ChecksumRegistry, fr *format.FormatRegistry) *ProtocolRegistry {
	return &ProtocolRegistry{
		parser:  pdl.NewPDLParser(cr, fr),
		schemas: make(map[string]*schema.ProtocolSchema),
	}
}

// RegisterFromPSL parses PSL text and registers the resulting schema.
func (r *ProtocolRegistry) RegisterFromPSL(pslText string) error {
	_, err := r.RegisterFromPSLReturnName(pslText)
	return err
}

// RegisterFromPSLReturnName parses PSL text, registers the schema, and returns the protocol name.
func (r *ProtocolRegistry) RegisterFromPSLReturnName(pslText string) (string, error) {
	s, err := r.parser.Parse(pslText)
	if err != nil {
		return "", err
	}
	if _, exists := r.schemas[s.Name]; exists {
		return "", &errors.ProtocolConflictError{Name: s.Name}
	}
	r.schemas[s.Name] = s
	return s.Name, nil
}

// RegisterFromFile reads a PSL file and registers the protocol.
func (r *ProtocolRegistry) RegisterFromFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	return r.RegisterFromPSL(string(data))
}

// GetSchema returns the schema for the given protocol name.
func (r *ProtocolRegistry) GetSchema(name string) (*schema.ProtocolSchema, error) {
	s, ok := r.schemas[name]
	if !ok {
		return nil, &errors.ProtocolNotFoundError{Name: name}
	}
	return s, nil
}

// Has checks whether a protocol is registered.
func (r *ProtocolRegistry) Has(name string) bool {
	_, ok := r.schemas[name]
	return ok
}

// List returns all registered protocol names in sorted order.
func (r *ProtocolRegistry) List() []string {
	names := make([]string, 0, len(r.schemas))
	for name := range r.schemas {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
