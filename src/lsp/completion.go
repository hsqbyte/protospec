// PSL LSP completion provider — enhanced code completion for PSL files.
package lsp

// CompletionItem represents a code completion item.
type CompletionItem struct {
	Label      string `json:"label"`
	Kind       int    `json:"kind"` // 1=Text, 2=Method, 3=Function, 6=Variable, 7=Class, 14=Keyword
	Detail     string `json:"detail"`
	InsertText string `json:"insertText"`
}

// KeywordCompletions returns PSL keyword completions.
func KeywordCompletions() []CompletionItem {
	return []CompletionItem{
		{Label: "protocol", Kind: 14, Detail: "Binary protocol definition", InsertText: "protocol ${1:Name} version \"${2:1.0}\" {\n    byte_order ${3:big-endian};\n    $0\n}"},
		{Label: "message", Kind: 14, Detail: "Message protocol definition", InsertText: "message ${1:Name} version \"${2:1.0}\" {\n    transport ${3:rest};\n    $0\n}"},
		{Label: "field", Kind: 14, Detail: "Field definition", InsertText: "field ${1:name}: ${2:uint8};"},
		{Label: "bitfield", Kind: 14, Detail: "Bitfield group", InsertText: "bitfield {\n    field ${1:name}: ${2:uint4};\n    $0\n}"},
		{Label: "byte_order", Kind: 14, Detail: "Byte order directive", InsertText: "byte_order ${1|big-endian,little-endian|};"},
		{Label: "request", Kind: 14, Detail: "Request message", InsertText: "request ${1:Name} {\n    field ${2:name}: ${3:string};\n    $0\n}"},
		{Label: "response", Kind: 14, Detail: "Response message", InsertText: "response ${1:Name} {\n    field ${2:name}: ${3:string};\n    $0\n}"},
		{Label: "enum", Kind: 14, Detail: "Enum values", InsertText: "enum {\n    ${1:0} = \"${2:value}\"\n}"},
		{Label: "checksum", Kind: 14, Detail: "Checksum field", InsertText: "checksum ${1:internet-checksum} covers ${2:fields}"},
		{Label: "uint8", Kind: 6, Detail: "8-bit unsigned integer"},
		{Label: "uint16", Kind: 6, Detail: "16-bit unsigned integer"},
		{Label: "uint32", Kind: 6, Detail: "32-bit unsigned integer"},
		{Label: "bytes", Kind: 6, Detail: "Variable-length byte sequence"},
		{Label: "string", Kind: 6, Detail: "UTF-8 string"},
	}
}

// HoverInfo returns hover information for a PSL token.
func HoverInfo(token string) string {
	info := map[string]string{
		"protocol":   "Defines a binary protocol with byte-level field definitions",
		"message":    "Defines a message protocol with structured request/response messages",
		"field":      "Declares a named field with a type and optional modifiers",
		"bitfield":   "Groups sub-byte fields that share byte boundaries",
		"byte_order": "Sets the default byte order (big-endian or little-endian)",
		"enum":       "Maps integer values to human-readable names",
		"checksum":   "Defines a checksum field with algorithm and coverage",
		"transport":  "Specifies the transport mechanism (rest, jsonrpc, graphql)",
		"optional":   "Marks a message field as optional",
	}
	return info[token]
}
