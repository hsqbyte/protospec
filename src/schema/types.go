package schema

// FieldType represents the data type of a protocol field.
type FieldType int

const (
	Uint   FieldType = iota // Unsigned integer
	Int                     // Signed integer
	Bytes                   // Byte sequence
	String                  // UTF-8 string
	Bool                    // Boolean (1 bit)
)

// String returns the string representation of a FieldType.
func (ft FieldType) String() string {
	switch ft {
	case Uint:
		return "uint"
	case Int:
		return "int"
	case Bytes:
		return "bytes"
	case String:
		return "string"
	case Bool:
		return "bool"
	default:
		return "unknown"
	}
}

// ByteOrder represents the byte order (endianness) for a field or protocol.
type ByteOrder int

const (
	BigEndian    ByteOrder = iota // Network byte order
	LittleEndian                  // Little-endian byte order
)

// String returns the string representation of a ByteOrder.
func (bo ByteOrder) String() string {
	switch bo {
	case BigEndian:
		return "big-endian"
	case LittleEndian:
		return "little-endian"
	default:
		return "unknown"
	}
}

// ChecksumConfig describes a checksum field's algorithm and coverage.
type ChecksumConfig struct {
	// Algorithm is the name of the checksum algorithm (e.g. "internet-checksum", "crc16").
	Algorithm string
	// CoverFields lists the fields covered by this checksum.
	// Supports both individual field names and range notation like "field1..field2".
	CoverFields []string
}

// LengthRef describes a variable-length field's length reference.
type LengthRef struct {
	// FieldName is the name of the field whose value determines the length.
	FieldName string
	// Scale is the multiplier applied to the referenced field's value.
	Scale int
	// Offset is added to the scaled value to compute the final byte length.
	Offset int
}

// ConditionExpr describes a conditional expression for optional fields.
type ConditionExpr struct {
	// FieldName is the field whose value is evaluated.
	FieldName string
	// Operator is the comparison operator (e.g. ">", "<", "==", "!=", ">=", "<=").
	Operator string
	// Value is the value to compare against.
	Value any
}

// FieldDef describes a single field in a protocol schema.
// A FieldDef with IsBitfieldGroup=true acts as a bitfield group container,
// with its sub-fields stored in BitfieldFields.
type FieldDef struct {
	// Name is the field name.
	Name string
	// Type is the data type of the field.
	Type FieldType
	// BitWidth is the width of the field in bits (1-64).
	BitWidth int
	// ByteOrder optionally overrides the protocol-level default byte order.
	ByteOrder *ByteOrder
	// EnumMap maps integer values to human-readable names.
	EnumMap map[int]string
	// Checksum holds the checksum configuration if this is a checksum field.
	Checksum *ChecksumConfig
	// Condition holds the condition expression for conditional fields.
	Condition *ConditionExpr
	// FixedLength is the fixed byte length for bytes/string fields (e.g. bytes[16]).
	// Zero means variable length (determined by LengthRef or remaining bytes).
	FixedLength int
	// LengthRef holds the length reference for variable-length fields.
	LengthRef *LengthRef
	// DisplayFormat is the name of the display formatter (e.g. "ipv4", "mac").
	DisplayFormat string
	// DefaultValue is the default value for this field (used during encoding if not provided).
	DefaultValue any
	// RangeMin and RangeMax define the valid value range for this field.
	RangeMin *int64
	RangeMax *int64
	// IsBitfieldGroup indicates whether this FieldDef represents a bitfield group.
	IsBitfieldGroup bool
	// BitfieldFields contains the sub-fields when IsBitfieldGroup is true.
	BitfieldFields []FieldDef
}

// ProtocolSchema is the internal representation of a protocol defined by PDL.
type ProtocolSchema struct {
	// Name is the protocol name (e.g. "IPv4", "UDP").
	Name string
	// Version is the protocol version string.
	Version string
	// DefaultByteOrder is the protocol-level default byte order.
	DefaultByteOrder ByteOrder
	// Constants maps constant names to their values.
	Constants map[string]int64
	// Fields is the ordered list of field definitions.
	Fields []FieldDef
	// Imports lists imported protocol/file references.
	Imports []string
	// Extends is the parent protocol name (if any).
	Extends string
	// TypeAliases maps alias names to their definitions.
	TypeAliases map[string]*TypeAlias
}

// TypeAlias defines a named type alias.
type TypeAlias struct {
	Name          string
	BaseType      FieldType
	BitWidth      int
	FixedLength   int
	DisplayFormat string
}

// EmbedDef describes an embedded protocol reference.
type EmbedDef struct {
	ProtocolName string
	Condition    *ConditionExpr
}

// MessageFieldType represents the type of a field in a message protocol.
type MessageFieldType int

const (
	MsgString  MessageFieldType = iota // string
	MsgNumber                          // number
	MsgBoolean                         // boolean
	MsgObject                          // object (nested fields)
	MsgArray                           // array
)

// String returns the string representation of a MessageFieldType.
func (ft MessageFieldType) String() string {
	switch ft {
	case MsgString:
		return "string"
	case MsgNumber:
		return "number"
	case MsgBoolean:
		return "boolean"
	case MsgObject:
		return "object"
	case MsgArray:
		return "array"
	default:
		return "unknown"
	}
}

// MessageFieldDef describes a field in a message definition.
type MessageFieldDef struct {
	Name     string
	Type     MessageFieldType
	Optional bool
	// Fields holds nested fields when Type is MsgObject.
	Fields []MessageFieldDef
	// ItemType holds the element type when Type is MsgArray.
	ItemType *MessageFieldDef
}

// MessageDef describes a single request, response, or notification.
type MessageDef struct {
	Name     string
	Kind     string // "request", "response", "notification"
	Fields   []MessageFieldDef
	Response *MessageDef // nested response (only if transport allows)
}

// MessageSchema is the internal representation of a message protocol.
type MessageSchema struct {
	Name         string
	Version      string
	Transport    string        // "jsonrpc", "rest", "graphql"
	TransportVer *string       // optional pinned version (e.g., "2.0")
	TransportDef *TransportDef // resolved transport definition
	Messages     []MessageDef
}

// TransportFieldDef describes a field defined in a transport message type.
type TransportFieldDef struct {
	Name         string
	Type         MessageFieldType
	Optional     bool
	DefaultValue any                 // nil means required from upper layer
	AutoValue    bool                // true for "default auto" (engine-generated)
	Fields       []TransportFieldDef // nested fields for object type
}

// MessageTypeDef describes a message type defined by a transport (e.g., request, notification, command).
type MessageTypeDef struct {
	Name        string
	Fields      []TransportFieldDef
	ResponseDef *MessageTypeDef // non-nil if this type supports nested response
}

// TransportDef holds the full transport definition extracted from a transport PSL.
type TransportDef struct {
	Name         string
	Version      string
	MessageTypes []MessageTypeDef
}
