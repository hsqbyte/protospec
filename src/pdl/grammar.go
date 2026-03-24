package pdl

import (
	"strconv"
)

// TokenType represents the type of a lexical token in PDL.
type TokenType int

const (
	// Special tokens
	TokenEOF TokenType = iota

	// Literals
	TokenIdent  // identifier
	TokenString // string literal (e.g. "1.0")
	TokenInt    // integer literal (e.g. 42)

	// Keywords
	TokenProtocol  // protocol
	TokenVersion   // version
	TokenByteOrder // byte_order
	TokenField     // field
	TokenBitfield  // bitfield
	TokenChecksum  // checksum
	TokenCovers    // covers
	TokenWhen      // when
	TokenDisplay   // display
	TokenEnum      // enum
	TokenLengthRef // length_ref
	TokenScale     // scale
	TokenOffset    // offset
	TokenConst     // const
	TokenRange     // range
	TokenDefault   // = (default value, context-dependent)
	TokenMessage   // message
	TokenRequest   // request
	TokenResponse  // response
	TokenNotify    // notification
	TokenTransport // transport
	TokenOptional  // optional
	TokenObject    // object
	TokenArray     // array
	TokenImport    // import
	TokenEmbed     // embed
	TokenExtends   // extends
	TokenTypeAlias // type (alias)

	// Byte order values (treated as keywords)
	TokenBigEndian    // big-endian
	TokenLittleEndian // little-endian

	// Punctuation
	TokenLBrace       // {
	TokenRBrace       // }
	TokenLBracket     // [
	TokenRBracket     // ]
	TokenColon        // :
	TokenSemicolon    // ;
	TokenComma        // ,
	TokenEquals       // =
	TokenDotDot       // ..
	TokenGreater      // >
	TokenLess         // <
	TokenGreaterEqual // >=
	TokenLessEqual    // <=
	TokenNotEqual     // !=
	TokenEqualEqual   // ==
	TokenMinus        // -
)

// String returns a human-readable name for the token type.
func (t TokenType) String() string {
	switch t {
	case TokenEOF:
		return "EOF"
	case TokenIdent:
		return "Ident"
	case TokenString:
		return "String"
	case TokenInt:
		return "Int"
	case TokenProtocol:
		return "protocol"
	case TokenVersion:
		return "version"
	case TokenByteOrder:
		return "byte_order"
	case TokenField:
		return "field"
	case TokenBitfield:
		return "bitfield"
	case TokenChecksum:
		return "checksum"
	case TokenCovers:
		return "covers"
	case TokenWhen:
		return "when"
	case TokenDisplay:
		return "display"
	case TokenEnum:
		return "enum"
	case TokenLengthRef:
		return "length_ref"
	case TokenScale:
		return "scale"
	case TokenOffset:
		return "offset"
	case TokenConst:
		return "const"
	case TokenRange:
		return "range"
	case TokenMessage:
		return "message"
	case TokenRequest:
		return "request"
	case TokenResponse:
		return "response"
	case TokenNotify:
		return "notification"
	case TokenTransport:
		return "transport"
	case TokenOptional:
		return "optional"
	case TokenObject:
		return "object"
	case TokenArray:
		return "array"
	case TokenImport:
		return "import"
	case TokenEmbed:
		return "embed"
	case TokenExtends:
		return "extends"
	case TokenTypeAlias:
		return "type"
	case TokenBigEndian:
		return "big-endian"
	case TokenLittleEndian:
		return "little-endian"
	case TokenLBrace:
		return "{"
	case TokenRBrace:
		return "}"
	case TokenLBracket:
		return "["
	case TokenRBracket:
		return "]"
	case TokenColon:
		return ":"
	case TokenSemicolon:
		return ";"
	case TokenComma:
		return ","
	case TokenEquals:
		return "="
	case TokenDotDot:
		return ".."
	case TokenGreater:
		return ">"
	case TokenLess:
		return "<"
	case TokenGreaterEqual:
		return ">="
	case TokenLessEqual:
		return "<="
	case TokenNotEqual:
		return "!="
	case TokenEqualEqual:
		return "=="
	case TokenMinus:
		return "-"
	default:
		return "Unknown"
	}
}

// Token represents a single lexical token produced by the PDL lexer.
type Token struct {
	Type   TokenType // The token type.
	Value  string    // The literal text of the token.
	Line   int       // 1-based line number where the token starts.
	Column int       // 1-based column number where the token starts.
}

// keywords maps keyword strings to their corresponding TokenType.
var keywords = map[string]TokenType{
	"protocol":      TokenProtocol,
	"version":       TokenVersion,
	"byte_order":    TokenByteOrder,
	"field":         TokenField,
	"bitfield":      TokenBitfield,
	"checksum":      TokenChecksum,
	"covers":        TokenCovers,
	"when":          TokenWhen,
	"display":       TokenDisplay,
	"enum":          TokenEnum,
	"length_ref":    TokenLengthRef,
	"scale":         TokenScale,
	"offset":        TokenOffset,
	"const":         TokenConst,
	"range":         TokenRange,
	"message":       TokenMessage,
	"request":       TokenRequest,
	"response":      TokenResponse,
	"notification":  TokenNotify,
	"transport":     TokenTransport,
	"optional":      TokenOptional,
	"object":        TokenObject,
	"array":         TokenArray,
	"import":        TokenImport,
	"embed":         TokenEmbed,
	"extends":       TokenExtends,
	"type":          TokenTypeAlias,
	"big-endian":    TokenBigEndian,
	"little-endian": TokenLittleEndian,
}

// LookupKeyword returns the keyword TokenType for the given identifier,
// or TokenIdent if it is not a keyword.
func LookupKeyword(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return TokenIdent
}

// ParseIntLiteral parses a decimal or hexadecimal integer literal string.
func ParseIntLiteral(s string) (int64, error) {
	if len(s) > 2 && s[0] == '0' && (s[1] == 'x' || s[1] == 'X') {
		v, err := strconv.ParseInt(s[2:], 16, 64)
		return v, err
	}
	v, err := strconv.ParseInt(s, 10, 64)
	return v, err
}
