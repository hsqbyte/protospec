package schema

// PSL 2.0 language extensions: generics, union types, expressions, modules.

// GenericParam represents a generic type parameter.
type GenericParam struct {
	Name       string
	Constraint string // optional constraint
}

// GenericType represents a generic type definition.
type GenericType struct {
	Name   string
	Params []GenericParam
	Fields []FieldDef
}

// UnionType represents a union of multiple types.
type UnionType struct {
	Name    string
	Options []string // protocol names or type names
}

// Expression represents a field value expression.
type Expression struct {
	Kind     ExprKind
	Left     *Expression
	Right    *Expression
	Operator string // +, -, *, /, &, |, ^, <<, >>
	FieldRef string // field reference
	Value    int64  // literal value
}

// ExprKind represents the kind of expression.
type ExprKind int

const (
	ExprLiteral ExprKind = iota
	ExprFieldRef
	ExprBinary
	ExprUnary
)

// ModuleDef represents a PSL module declaration.
type ModuleDef struct {
	Name      string
	Public    bool
	Imports   []UseStatement
	Constants map[string]int64
	Types     map[string]*TypeAlias
	Protocols []string
}

// UseStatement represents a selective import.
type UseStatement struct {
	Module string
	Names  []string // specific names to import, empty = import all
}

// Visibility represents field/type visibility.
type Visibility int

const (
	Private Visibility = iota
	Public
)

// EnumDef represents a standalone enum type definition.
type EnumDef struct {
	Name   string
	Values map[int64]string
}

// PSL2Extensions holds all PSL 2.0 extensions for a protocol.
type PSL2Extensions struct {
	Generics    []GenericType
	Unions      []UnionType
	Enums       []EnumDef
	Module      *ModuleDef
	Expressions map[string]*Expression // field name -> computed expression
}
