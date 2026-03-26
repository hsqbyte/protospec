package pdl

import "github.com/hsqbyte/protospec/src/core/schema"

// DynamicKeywordRegistry holds transport-derived keywords for a parsing session.
// It indexes MessageTypeDef names from a TransportDef for O(1) lookup.
type DynamicKeywordRegistry struct {
	messageTypes map[string]*schema.MessageTypeDef
	transport    *schema.TransportDef
}

// NewDynamicKeywordRegistry creates a registry from a TransportDef,
// indexing all MessageTypeDef names in a map for O(1) lookup.
func NewDynamicKeywordRegistry(td *schema.TransportDef) *DynamicKeywordRegistry {
	r := &DynamicKeywordRegistry{
		messageTypes: make(map[string]*schema.MessageTypeDef, len(td.MessageTypes)),
		transport:    td,
	}
	for i := range td.MessageTypes {
		r.messageTypes[td.MessageTypes[i].Name] = &td.MessageTypes[i]
	}
	return r
}

// IsMessageType checks if a token is a transport-defined message type.
func (r *DynamicKeywordRegistry) IsMessageType(name string) bool {
	_, ok := r.messageTypes[name]
	return ok
}

// GetMessageType returns the MessageTypeDef for a message type name.
func (r *DynamicKeywordRegistry) GetMessageType(name string) (*schema.MessageTypeDef, bool) {
	mtd, ok := r.messageTypes[name]
	return mtd, ok
}

// GetFieldDef returns the TransportFieldDef for a field within a message type.
func (r *DynamicKeywordRegistry) GetFieldDef(messageType, fieldName string) (*schema.TransportFieldDef, bool) {
	mtd, ok := r.messageTypes[messageType]
	if !ok {
		return nil, false
	}
	for i := range mtd.Fields {
		if mtd.Fields[i].Name == fieldName {
			return &mtd.Fields[i], true
		}
	}
	return nil, false
}

// MessageTypeNames returns all registered message type names.
func (r *DynamicKeywordRegistry) MessageTypeNames() []string {
	names := make([]string, 0, len(r.messageTypes))
	for name := range r.messageTypes {
		names = append(names, name)
	}
	return names
}
