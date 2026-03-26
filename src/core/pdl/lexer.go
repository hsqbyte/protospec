package pdl

import (
	"fmt"
	"github.com/hsqbyte/protospec/src/core/errors"
	"strings"
)

// Lexer performs lexical analysis on PDL source text, producing tokens one at a time.
type Lexer struct {
	source string
	pos    int
	line   int
	col    int
}

// NewLexer creates a new Lexer for the given PDL source text.
func NewLexer(source string) *Lexer {
	return &Lexer{
		source: source,
		pos:    0,
		line:   1,
		col:    1,
	}
}

// peek returns the current character without advancing, or 0 if at end.
func (l *Lexer) peek() byte {
	if l.pos >= len(l.source) {
		return 0
	}
	return l.source[l.pos]
}

// advance moves forward one character, updating line and column tracking.
func (l *Lexer) advance() byte {
	ch := l.source[l.pos]
	l.pos++
	if ch == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return ch
}

// skipWhitespaceAndComments skips spaces, tabs, newlines, and // comments.
func (l *Lexer) skipWhitespaceAndComments() {
	for l.pos < len(l.source) {
		ch := l.source[l.pos]
		if ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' {
			l.advance()
			continue
		}
		// Single-line comment
		if ch == '/' && l.pos+1 < len(l.source) && l.source[l.pos+1] == '/' {
			l.advance() // skip first /
			l.advance() // skip second /
			for l.pos < len(l.source) && l.source[l.pos] != '\n' {
				l.advance()
			}
			continue
		}
		break
	}
}

// NextToken returns the next token from the source.
// Returns TokenEOF at end of input, or a PDLSyntaxError for illegal characters.
func (l *Lexer) NextToken() (Token, error) {
	l.skipWhitespaceAndComments()

	if l.pos >= len(l.source) {
		return Token{Type: TokenEOF, Value: "", Line: l.line, Column: l.col}, nil
	}

	startLine := l.line
	startCol := l.col
	ch := l.peek()

	// String literal
	if ch == '"' {
		return l.scanString()
	}

	// Integer literal (digits or 0x hex)
	if ch >= '0' && ch <= '9' {
		return l.scanNumber(startLine, startCol), nil
	}

	// Identifier or keyword
	if isIdentStart(ch) {
		return l.scanIdentOrKeyword(startLine, startCol), nil
	}

	// Two-character tokens and single-character punctuation
	switch ch {
	case '{':
		l.advance()
		return Token{Type: TokenLBrace, Value: "{", Line: startLine, Column: startCol}, nil
	case '}':
		l.advance()
		return Token{Type: TokenRBrace, Value: "}", Line: startLine, Column: startCol}, nil
	case '[':
		l.advance()
		return Token{Type: TokenLBracket, Value: "[", Line: startLine, Column: startCol}, nil
	case ']':
		l.advance()
		return Token{Type: TokenRBracket, Value: "]", Line: startLine, Column: startCol}, nil
	case ':':
		l.advance()
		return Token{Type: TokenColon, Value: ":", Line: startLine, Column: startCol}, nil
	case ';':
		l.advance()
		return Token{Type: TokenSemicolon, Value: ";", Line: startLine, Column: startCol}, nil
	case ',':
		l.advance()
		return Token{Type: TokenComma, Value: ",", Line: startLine, Column: startCol}, nil
	case '-':
		l.advance()
		return Token{Type: TokenMinus, Value: "-", Line: startLine, Column: startCol}, nil
	case '.':
		if l.pos+1 < len(l.source) && l.source[l.pos+1] == '.' {
			l.advance()
			l.advance()
			return Token{Type: TokenDotDot, Value: "..", Line: startLine, Column: startCol}, nil
		}
		l.advance()
		return Token{}, &errors.PDLSyntaxError{Line: startLine, Column: startCol, Message: fmt.Sprintf("unexpected character %q", ch)}
	case '=':
		if l.pos+1 < len(l.source) && l.source[l.pos+1] == '=' {
			l.advance()
			l.advance()
			return Token{Type: TokenEqualEqual, Value: "==", Line: startLine, Column: startCol}, nil
		}
		l.advance()
		return Token{Type: TokenEquals, Value: "=", Line: startLine, Column: startCol}, nil
	case '>':
		if l.pos+1 < len(l.source) && l.source[l.pos+1] == '=' {
			l.advance()
			l.advance()
			return Token{Type: TokenGreaterEqual, Value: ">=", Line: startLine, Column: startCol}, nil
		}
		l.advance()
		return Token{Type: TokenGreater, Value: ">", Line: startLine, Column: startCol}, nil
	case '<':
		if l.pos+1 < len(l.source) && l.source[l.pos+1] == '=' {
			l.advance()
			l.advance()
			return Token{Type: TokenLessEqual, Value: "<=", Line: startLine, Column: startCol}, nil
		}
		l.advance()
		return Token{Type: TokenLess, Value: "<", Line: startLine, Column: startCol}, nil
	case '!':
		if l.pos+1 < len(l.source) && l.source[l.pos+1] == '=' {
			l.advance()
			l.advance()
			return Token{Type: TokenNotEqual, Value: "!=", Line: startLine, Column: startCol}, nil
		}
		l.advance()
		return Token{}, &errors.PDLSyntaxError{Line: startLine, Column: startCol, Message: fmt.Sprintf("unexpected character %q", ch)}
	}

	// Illegal character
	l.advance()
	return Token{}, &errors.PDLSyntaxError{Line: startLine, Column: startCol, Message: fmt.Sprintf("unexpected character %q", ch)}
}

// scanString scans a double-quoted string literal with basic escape sequences.
func (l *Lexer) scanString() (Token, error) {
	startLine := l.line
	startCol := l.col
	l.advance() // skip opening quote

	var sb strings.Builder
	for l.pos < len(l.source) {
		ch := l.source[l.pos]
		if ch == '"' {
			l.advance() // skip closing quote
			return Token{Type: TokenString, Value: sb.String(), Line: startLine, Column: startCol}, nil
		}
		if ch == '\\' {
			l.advance() // skip backslash
			if l.pos >= len(l.source) {
				return Token{}, &errors.PDLSyntaxError{Line: l.line, Column: l.col, Message: "unterminated string literal"}
			}
			esc := l.source[l.pos]
			switch esc {
			case 'n':
				sb.WriteByte('\n')
			case 't':
				sb.WriteByte('\t')
			case '\\':
				sb.WriteByte('\\')
			case '"':
				sb.WriteByte('"')
			default:
				sb.WriteByte('\\')
				sb.WriteByte(esc)
			}
			l.advance()
			continue
		}
		if ch == '\n' {
			return Token{}, &errors.PDLSyntaxError{Line: startLine, Column: startCol, Message: "unterminated string literal"}
		}
		sb.WriteByte(ch)
		l.advance()
	}
	return Token{}, &errors.PDLSyntaxError{Line: startLine, Column: startCol, Message: "unterminated string literal"}
}

// scanNumber scans a decimal or hexadecimal (0x...) integer literal.
func (l *Lexer) scanNumber(startLine, startCol int) Token {
	start := l.pos
	// Check for hex prefix 0x or 0X
	if l.source[l.pos] == '0' && l.pos+1 < len(l.source) && (l.source[l.pos+1] == 'x' || l.source[l.pos+1] == 'X') {
		l.advance() // '0'
		l.advance() // 'x'
		for l.pos < len(l.source) && isHexDigit(l.source[l.pos]) {
			l.advance()
		}
		return Token{Type: TokenInt, Value: l.source[start:l.pos], Line: startLine, Column: startCol}
	}
	// Decimal
	for l.pos < len(l.source) && l.source[l.pos] >= '0' && l.source[l.pos] <= '9' {
		l.advance()
	}
	return Token{Type: TokenInt, Value: l.source[start:l.pos], Line: startLine, Column: startCol}
}

// isHexDigit returns true if ch is a valid hexadecimal digit.
func isHexDigit(ch byte) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

// scanIdentOrKeyword scans an identifier or keyword, with special handling
// for "big-endian" and "little-endian".
func (l *Lexer) scanIdentOrKeyword(startLine, startCol int) Token {
	start := l.pos
	for l.pos < len(l.source) && isIdentChar(l.source[l.pos]) {
		l.advance()
	}
	word := l.source[start:l.pos]

	// Special case: "big" or "little" may be followed by "-endian"
	if (word == "big" || word == "little") && l.pos < len(l.source) && l.source[l.pos] == '-' {
		suffix := "-endian"
		if l.pos+len(suffix) <= len(l.source) && l.source[l.pos:l.pos+len(suffix)] == suffix {
			for i := 0; i < len(suffix); i++ {
				l.advance()
			}
			word = l.source[start:l.pos]
		}
	}

	tokType := LookupKeyword(word)
	return Token{Type: tokType, Value: word, Line: startLine, Column: startCol}
}

// isIdentStart returns true if ch can start an identifier (letter or underscore).
func isIdentStart(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

// isIdentChar returns true if ch can appear in an identifier (letter, digit, or underscore).
func isIdentChar(ch byte) bool {
	return isIdentStart(ch) || (ch >= '0' && ch <= '9')
}
