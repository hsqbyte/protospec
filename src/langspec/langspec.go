// Package langspec provides PSL 3.0 language specification and EBNF grammar.
package langspec

// Version is the PSL language specification version.
const Version = "3.0"

// EBNF is the formal grammar of PSL 3.0 in EBNF notation.
const EBNF = `
(* PSL 3.0 Grammar — Extended Backus-Naur Form *)

program        = protocol_def | message_def ;

(* Binary Protocol *)
protocol_def   = "protocol" , identifier , "version" , string_lit , "{" , protocol_body , "}" ;
protocol_body  = { directive | field_def | bitfield_group | const_def | type_alias | import_stmt } ;

directive      = byte_order | extends ;
byte_order     = "byte_order" , ( "big-endian" | "little-endian" ) , ";" ;
extends        = "extends" , identifier , ";" ;
import_stmt    = "import" , string_lit , ";" ;

const_def      = "const" , identifier , "=" , integer , ";" ;
type_alias     = "type" , identifier , "=" , type_spec , ";" ;

field_def      = "field" , identifier , ":" , type_spec , { field_modifier } , ";" ;
type_spec      = base_type , [ "[" , integer , "]" ] ;
base_type      = "uint" , digits | "int" , digits | "bytes" | "string" | "bool" | identifier ;

field_modifier = enum_def | checksum_def | length_ref | condition | display | default_val | range_def ;
enum_def       = "enum" , "{" , enum_entry , { "," , enum_entry } , "}" ;
enum_entry     = integer , "=" , string_lit ;
checksum_def   = "checksum" , identifier , "covers" , field_list ;
length_ref     = "length" , identifier , [ "*" , integer ] , [ "+" , integer ] ;
condition      = "if" , identifier , comp_op , value ;
display        = "display" , "=" , string_lit ;
default_val    = "default" , value ;
range_def      = "range" , integer , ".." , integer ;

bitfield_group = "bitfield" , "{" , { field_def } , "}" ;

(* Message Protocol *)
message_def    = "message" , identifier , "version" , string_lit , "{" , message_body , "}" ;
message_body   = { transport_def | request_def | response_def | notify_def } ;
transport_def  = "transport" , identifier , ";" ;
request_def    = "request" , identifier , "{" , { msg_field } , "}" ;
response_def   = "response" , identifier , "{" , { msg_field } , "}" ;
notify_def     = "notification" , identifier , ( "{" , { msg_field } , "}" | ";" ) ;
msg_field      = "field" , identifier , ":" , msg_type , [ "optional" ] , ";" ;
msg_type       = "string" | "number" | "boolean" | "object" | "array" ;

(* Terminals *)
identifier     = letter , { letter | digit | "_" } ;
string_lit     = '"' , { character } , '"' ;
integer        = [ "0x" ] , digits ;
digits         = digit , { digit } ;
comp_op        = "==" | "!=" | ">" | "<" | ">=" | "<=" ;
value          = integer | string_lit ;
`

// SemanticRules documents the semantic rules of PSL 3.0.
var SemanticRules = []string{
	"Protocol names must be unique within a library",
	"Field names must be unique within a protocol",
	"Bitfield groups must sum to a multiple of 8 bits",
	"Checksum covers fields must reference existing fields",
	"Length references must point to integer fields",
	"Condition fields must reference previously defined fields",
	"Enum values must be unique within a field",
	"Type aliases must reference valid base types",
	"Import paths must resolve to valid PSL files",
	"Extends must reference a registered protocol",
}

// TypeSystem documents the PSL 3.0 type system.
var TypeSystem = map[string]string{
	"uint<N>":   "Unsigned integer of N bits (1-64)",
	"int<N>":    "Signed integer of N bits (1-64)",
	"bytes":     "Variable-length byte sequence",
	"bytes[N]":  "Fixed-length byte sequence of N bytes",
	"string":    "UTF-8 encoded string",
	"string[N]": "Fixed-length string of N bytes",
	"bool":      "Boolean value (1 bit)",
}
