package pdl

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hsqbyte/protospec/src/core/checksum"
	"github.com/hsqbyte/protospec/src/core/errors"
	"github.com/hsqbyte/protospec/src/core/format"
	"github.com/hsqbyte/protospec/src/core/schema"
)

// PDLParser parses PDL source text into a ProtocolSchema using recursive descent.
type PDLParser struct {
	lexer            *Lexer
	current          Token
	source           string
	checksumRegistry *checksum.ChecksumRegistry
	formatRegistry   *format.FormatRegistry
	transportLoader  *TransportLoader
	dynamicKeywords  *DynamicKeywordRegistry
}

// SetTransportLoader sets the transport loader for dynamic keyword support.
func (p *PDLParser) SetTransportLoader(loader *TransportLoader) {
	p.transportLoader = loader
}

// NewPDLParser creates a new PDLParser with optional registries for semantic validation.
func NewPDLParser(checksumReg *checksum.ChecksumRegistry, formatReg *format.FormatRegistry) *PDLParser {
	return &PDLParser{
		checksumRegistry: checksumReg,
		formatRegistry:   formatReg,
	}
}

// Parse parses PDL source text and returns a ProtocolSchema.
func (p *PDLParser) Parse(source string) (*schema.ProtocolSchema, error) {
	p.source = source
	p.lexer = NewLexer(source)
	if err := p.advance(); err != nil {
		return nil, err
	}

	// Parse top-level imports before protocol
	var imports []string
	for p.current.Type == TokenImport {
		imp, err := p.parseImport()
		if err != nil {
			return nil, err
		}
		imports = append(imports, imp)
	}

	// Parse type aliases before protocol
	typeAliases := make(map[string]*schema.TypeAlias)
	for p.current.Type == TokenTypeAlias {
		alias, err := p.parseTypeAlias()
		if err != nil {
			return nil, err
		}
		typeAliases[alias.Name] = alias
	}

	ps, err := p.parseProtocol()
	if err != nil {
		return nil, err
	}
	ps.Imports = imports
	ps.TypeAliases = typeAliases
	return ps, nil
}

// advance consumes the next token from the lexer.
func (p *PDLParser) advance() error {
	tok, err := p.lexer.NextToken()
	if err != nil {
		return err
	}
	p.current = tok
	return nil
}

// expect checks that the current token matches the expected type, then advances.
// Returns the matched token or a PDLSyntaxError.
func (p *PDLParser) expect(tt TokenType) (Token, error) {
	if p.current.Type != tt {
		return Token{}, p.syntaxError(fmt.Sprintf("expected %s, got %s (%q)", tt, p.current.Type, p.current.Value))
	}
	tok := p.current
	if err := p.advance(); err != nil {
		return Token{}, err
	}
	return tok, nil
}

// isIdentLike returns true if the token can be used as an identifier.
// Keywords can appear as field names, algorithm names, etc.
func isIdentLike(tt TokenType) bool {
	switch tt {
	case TokenIdent, TokenProtocol, TokenVersion, TokenByteOrder, TokenField,
		TokenBitfield, TokenChecksum, TokenCovers, TokenWhen, TokenDisplay,
		TokenEnum, TokenLengthRef, TokenScale, TokenOffset,
		TokenMessage, TokenRequest, TokenResponse, TokenNotify, TokenTransport,
		TokenOptional, TokenObject, TokenArray,
		TokenImport, TokenEmbed, TokenExtends, TokenTypeAlias:
		return true
	}
	return false
}

// expectIdent expects an identifier-like token (including keywords used as names).
func (p *PDLParser) expectIdent() (Token, error) {
	if !isIdentLike(p.current.Type) {
		return Token{}, p.syntaxError(fmt.Sprintf("expected identifier, got %s (%q)", p.current.Type, p.current.Value))
	}
	tok := p.current
	if err := p.advance(); err != nil {
		return Token{}, err
	}
	return tok, nil
}

// syntaxError creates a PDLSyntaxError at the current token position.
func (p *PDLParser) syntaxError(msg string) *errors.PDLSyntaxError {
	srcLine := ""
	if p.source != "" && p.current.Line > 0 {
		lines := strings.Split(p.source, "\n")
		if p.current.Line <= len(lines) {
			srcLine = lines[p.current.Line-1]
		}
	}
	return &errors.PDLSyntaxError{
		Line:    p.current.Line,
		Column:  p.current.Column,
		Message: msg,
		Source:  srcLine,
	}
}

// parseProtocol parses: protocol <name> [extends <parent>] version "<ver>" { ... }
func (p *PDLParser) parseProtocol() (*schema.ProtocolSchema, error) {
	if _, err := p.expect(TokenProtocol); err != nil {
		return nil, err
	}

	nameTok, err := p.expectIdent()
	if err != nil {
		return nil, err
	}

	ps := &schema.ProtocolSchema{
		Name:             nameTok.Value,
		DefaultByteOrder: schema.BigEndian,
	}

	// Optional extends
	if p.current.Type == TokenExtends {
		if err := p.advance(); err != nil {
			return nil, err
		}
		parentTok, err := p.expectIdent()
		if err != nil {
			return nil, err
		}
		ps.Extends = parentTok.Value
	}

	if _, err := p.expect(TokenVersion); err != nil {
		return nil, err
	}

	versionTok, err := p.expect(TokenString)
	if err != nil {
		return nil, err
	}
	ps.Version = versionTok.Value

	if _, err := p.expect(TokenLBrace); err != nil {
		return nil, err
	}

	if err := p.parseProtocolBody(ps); err != nil {
		return nil, err
	}

	if _, err := p.expect(TokenRBrace); err != nil {
		return nil, err
	}

	p.markPayload(ps)

	if err := p.validate(ps); err != nil {
		return nil, err
	}

	return ps, nil
}

// parseProtocolBody parses the body inside protocol { ... }.
func (p *PDLParser) parseProtocolBody(ps *schema.ProtocolSchema) error {
	for p.current.Type != TokenRBrace && p.current.Type != TokenEOF {
		switch p.current.Type {
		case TokenByteOrder:
			if err := p.parseByteOrder(ps); err != nil {
				return err
			}
		case TokenConst:
			if err := p.parseConst(ps); err != nil {
				return err
			}
		case TokenField:
			fd, err := p.parseField()
			if err != nil {
				return err
			}
			ps.Fields = append(ps.Fields, *fd)
		case TokenBitfield:
			fd, err := p.parseBitfield()
			if err != nil {
				return err
			}
			ps.Fields = append(ps.Fields, *fd)
		case TokenEmbed:
			// embed <protocol> [when <cond>] ;
			if err := p.advance(); err != nil {
				return err
			}
			nameTok, err := p.expectIdent()
			if err != nil {
				return err
			}
			embed := schema.FieldDef{
				Name:     "__embed_" + nameTok.Value,
				Type:     schema.Bytes,
				BitWidth: 0,
			}
			// Optional condition
			if p.current.Type == TokenWhen {
				cond, err := p.parseWhen()
				if err != nil {
					return err
				}
				embed.Condition = cond
			}
			if _, err := p.expect(TokenSemicolon); err != nil {
				return err
			}
			ps.Fields = append(ps.Fields, embed)
		default:
			return p.syntaxError(fmt.Sprintf("unexpected token %s (%q) in protocol body", p.current.Type, p.current.Value))
		}
	}
	return nil
}

// parseImport parses: import "<path>";
func (p *PDLParser) parseImport() (string, error) {
	if err := p.advance(); err != nil { // consume 'import'
		return "", err
	}
	pathTok, err := p.expect(TokenString)
	if err != nil {
		return "", err
	}
	if _, err := p.expect(TokenSemicolon); err != nil {
		return "", err
	}
	return pathTok.Value, nil
}

// parseTypeAlias parses: type <name> = <type>[N] [display <fmt>] ;
func (p *PDLParser) parseTypeAlias() (*schema.TypeAlias, error) {
	if err := p.advance(); err != nil { // consume 'type'
		return nil, err
	}
	nameTok, err := p.expectIdent()
	if err != nil {
		return nil, err
	}
	if _, err := p.expect(TokenEquals); err != nil {
		return nil, err
	}

	alias := &schema.TypeAlias{Name: nameTok.Value}

	// Parse base type
	typeTok, err := p.expectIdent()
	if err != nil {
		return nil, err
	}

	fd := &schema.FieldDef{}
	if err := p.parseFieldType(fd, typeTok.Value); err != nil {
		return nil, err
	}
	alias.BaseType = fd.Type
	alias.BitWidth = fd.BitWidth
	alias.FixedLength = fd.FixedLength

	// Optional display
	if p.current.Type == TokenDisplay {
		df, err := p.parseDisplay()
		if err != nil {
			return nil, err
		}
		alias.DisplayFormat = df
	}

	if _, err := p.expect(TokenSemicolon); err != nil {
		return nil, err
	}
	return alias, nil
}

// parseByteOrder parses: byte_order big-endian|little-endian ;
func (p *PDLParser) parseByteOrder(ps *schema.ProtocolSchema) error {
	if err := p.advance(); err != nil { // consume byte_order
		return err
	}

	switch p.current.Type {
	case TokenBigEndian:
		ps.DefaultByteOrder = schema.BigEndian
	case TokenLittleEndian:
		ps.DefaultByteOrder = schema.LittleEndian
	default:
		return p.syntaxError(fmt.Sprintf("expected big-endian or little-endian, got %q", p.current.Value))
	}

	if err := p.advance(); err != nil {
		return err
	}

	if _, err := p.expect(TokenSemicolon); err != nil {
		return err
	}
	return nil
}

// parseConst parses: const NAME = <int>;
func (p *PDLParser) parseConst(ps *schema.ProtocolSchema) error {
	if err := p.advance(); err != nil { // consume 'const'
		return err
	}

	nameTok, err := p.expectIdent()
	if err != nil {
		return err
	}

	if _, err := p.expect(TokenEquals); err != nil {
		return err
	}

	if p.current.Type != TokenInt {
		return p.syntaxError("expected integer value for const")
	}
	val, parseErr := ParseIntLiteral(p.current.Value)
	if parseErr != nil {
		return p.syntaxError(fmt.Sprintf("invalid const value %q", p.current.Value))
	}
	if err := p.advance(); err != nil {
		return err
	}

	if _, err := p.expect(TokenSemicolon); err != nil {
		return err
	}

	if ps.Constants == nil {
		ps.Constants = make(map[string]int64)
	}
	ps.Constants[nameTok.Value] = val
	return nil
}

// parseField parses: field <name>: <type><bits> [modifiers...] ;
func (p *PDLParser) parseField() (*schema.FieldDef, error) {
	if err := p.advance(); err != nil { // consume 'field'
		return nil, err
	}

	nameTok, err := p.expectIdent()
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(TokenColon); err != nil {
		return nil, err
	}

	fd := &schema.FieldDef{
		Name: nameTok.Value,
	}

	// Parse type+bits (e.g. "uint16", "bytes", "bool")
	typeTok, err := p.expectIdent()
	if err != nil {
		return nil, err
	}
	if err := p.parseFieldType(fd, typeTok.Value); err != nil {
		return nil, err
	}

	// Parse optional modifiers until semicolon
	if err := p.parseFieldModifiers(fd); err != nil {
		return nil, err
	}

	if _, err := p.expect(TokenSemicolon); err != nil {
		return nil, err
	}

	return fd, nil
}

// parseFieldType parses a type+bits string like "uint16", "int8", "bytes", "string", "bool".
func (p *PDLParser) parseFieldType(fd *schema.FieldDef, typeStr string) error {
	switch {
	case typeStr == "bytes":
		fd.Type = schema.Bytes
		fd.BitWidth = 0
		// Check for fixed-length syntax: bytes[N]
		if p.current.Type == TokenLBracket {
			if err := p.advance(); err != nil { // consume '['
				return err
			}
			if p.current.Type != TokenInt {
				return p.syntaxError("expected integer in bytes[N]")
			}
			n, err := ParseIntLiteral(p.current.Value)
			if err != nil || n <= 0 {
				return p.syntaxError(fmt.Sprintf("invalid fixed length %q", p.current.Value))
			}
			fd.FixedLength = int(n)
			if err := p.advance(); err != nil { // consume number
				return err
			}
			if _, err := p.expect(TokenRBracket); err != nil {
				return err
			}
		}
	case typeStr == "string":
		fd.Type = schema.String
		fd.BitWidth = 0
	case typeStr == "bool":
		fd.Type = schema.Bool
		fd.BitWidth = 1
	case strings.HasPrefix(typeStr, "uint"):
		fd.Type = schema.Uint
		bits, err := strconv.Atoi(typeStr[4:])
		if err != nil || bits <= 0 {
			return p.syntaxError(fmt.Sprintf("invalid type %q: cannot parse bit width", typeStr))
		}
		fd.BitWidth = bits
	case strings.HasPrefix(typeStr, "int"):
		fd.Type = schema.Int
		bits, err := strconv.Atoi(typeStr[3:])
		if err != nil || bits <= 0 {
			return p.syntaxError(fmt.Sprintf("invalid type %q: cannot parse bit width", typeStr))
		}
		fd.BitWidth = bits
	default:
		return p.syntaxError(fmt.Sprintf("unknown type %q", typeStr))
	}
	return nil
}

// parseFieldModifiers parses optional modifiers: checksum, length_ref, enum, when, display.
// These can appear in any order, terminated by the upcoming semicolon.
func (p *PDLParser) parseFieldModifiers(fd *schema.FieldDef) error {
	for {
		switch p.current.Type {
		case TokenChecksum:
			cfg, err := p.parseChecksum()
			if err != nil {
				return err
			}
			fd.Checksum = cfg
		case TokenLengthRef:
			lr, err := p.parseLengthRef()
			if err != nil {
				return err
			}
			fd.LengthRef = lr
		case TokenEnum:
			em, err := p.parseEnum()
			if err != nil {
				return err
			}
			fd.EnumMap = em
		case TokenWhen:
			cond, err := p.parseWhen()
			if err != nil {
				return err
			}
			fd.Condition = cond
		case TokenDisplay:
			df, err := p.parseDisplay()
			if err != nil {
				return err
			}
			fd.DisplayFormat = df
		case TokenEquals:
			// Default value: = <int>
			if err := p.advance(); err != nil {
				return err
			}
			if p.current.Type != TokenInt {
				return p.syntaxError("expected integer for default value")
			}
			val, err := ParseIntLiteral(p.current.Value)
			if err != nil {
				return p.syntaxError(fmt.Sprintf("invalid default value %q", p.current.Value))
			}
			fd.DefaultValue = val
			if err := p.advance(); err != nil {
				return err
			}
		case TokenRange:
			// range [min..max]
			if err := p.advance(); err != nil {
				return err
			}
			if _, err := p.expect(TokenLBracket); err != nil {
				return err
			}
			if p.current.Type != TokenInt {
				return p.syntaxError("expected integer for range min")
			}
			minVal, err := ParseIntLiteral(p.current.Value)
			if err != nil {
				return p.syntaxError(fmt.Sprintf("invalid range min %q", p.current.Value))
			}
			fd.RangeMin = &minVal
			if err := p.advance(); err != nil {
				return err
			}
			if _, err := p.expect(TokenDotDot); err != nil {
				return err
			}
			if p.current.Type != TokenInt {
				return p.syntaxError("expected integer for range max")
			}
			maxVal, err := ParseIntLiteral(p.current.Value)
			if err != nil {
				return p.syntaxError(fmt.Sprintf("invalid range max %q", p.current.Value))
			}
			fd.RangeMax = &maxVal
			if err := p.advance(); err != nil {
				return err
			}
			if _, err := p.expect(TokenRBracket); err != nil {
				return err
			}
		default:
			// No more modifiers
			return nil
		}
	}
}

// parseChecksum parses: checksum <algo> covers [field1, field2..field3, field4]
// Algorithm names can contain hyphens (e.g. "internet-checksum").
func (p *PDLParser) parseChecksum() (*schema.ChecksumConfig, error) {
	if err := p.advance(); err != nil { // consume 'checksum'
		return nil, err
	}

	algoName, err := p.parseHyphenatedIdent()
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(TokenCovers); err != nil {
		return nil, err
	}

	if _, err := p.expect(TokenLBracket); err != nil {
		return nil, err
	}

	var coverFields []string
	for p.current.Type != TokenRBracket {
		fieldTok, err := p.expectIdent()
		if err != nil {
			return nil, err
		}

		// Check for range notation: field1..field2
		if p.current.Type == TokenDotDot {
			if err := p.advance(); err != nil { // consume '..'
				return nil, err
			}
			endTok, err := p.expectIdent()
			if err != nil {
				return nil, err
			}
			coverFields = append(coverFields, fieldTok.Value+".."+endTok.Value)
		} else {
			coverFields = append(coverFields, fieldTok.Value)
		}

		if p.current.Type == TokenComma {
			if err := p.advance(); err != nil { // consume ','
				return nil, err
			}
		}
	}

	if _, err := p.expect(TokenRBracket); err != nil {
		return nil, err
	}

	return &schema.ChecksumConfig{
		Algorithm:   algoName,
		CoverFields: coverFields,
	}, nil
}

// parseHyphenatedIdent parses an identifier that may contain hyphens,
// e.g. "internet-checksum". The lexer produces separate tokens for each part.
func (p *PDLParser) parseHyphenatedIdent() (string, error) {
	tok, err := p.expectIdent()
	if err != nil {
		return "", err
	}
	name := tok.Value

	// Consume hyphen-separated continuations
	for p.current.Type == TokenMinus {
		if err := p.advance(); err != nil { // consume '-'
			return "", err
		}
		next, err := p.expectIdent()
		if err != nil {
			return "", err
		}
		name += "-" + next.Value
	}

	return name, nil
}

// parseLengthRef parses: length_ref <field> [scale N] [offset N]
// offset can be negative (TokenMinus + TokenInt).
func (p *PDLParser) parseLengthRef() (*schema.LengthRef, error) {
	if err := p.advance(); err != nil { // consume 'length_ref'
		return nil, err
	}

	fieldTok, err := p.expectIdent()
	if err != nil {
		return nil, err
	}

	lr := &schema.LengthRef{
		FieldName: fieldTok.Value,
		Scale:     1,
		Offset:    0,
	}

	// Optional scale
	if p.current.Type == TokenScale {
		if err := p.advance(); err != nil { // consume 'scale'
			return nil, err
		}
		valTok, err := p.expect(TokenInt)
		if err != nil {
			return nil, err
		}
		v, err := ParseIntLiteral(valTok.Value)
		if err != nil {
			return nil, p.syntaxError(fmt.Sprintf("invalid scale value %q", valTok.Value))
		}
		lr.Scale = int(v)
	}

	// Optional offset (can be negative)
	if p.current.Type == TokenOffset {
		if err := p.advance(); err != nil { // consume 'offset'
			return nil, err
		}
		negative := false
		if p.current.Type == TokenMinus {
			negative = true
			if err := p.advance(); err != nil { // consume '-'
				return nil, err
			}
		}
		valTok, err := p.expect(TokenInt)
		if err != nil {
			return nil, err
		}
		v, err := ParseIntLiteral(valTok.Value)
		if err != nil {
			return nil, p.syntaxError(fmt.Sprintf("invalid offset value %q", valTok.Value))
		}
		if negative {
			v = -v
		}
		lr.Offset = int(v)
	}

	return lr, nil
}

// parseEnum parses: enum { <int> = <string>, ... }
func (p *PDLParser) parseEnum() (map[int]string, error) {
	if err := p.advance(); err != nil { // consume 'enum'
		return nil, err
	}

	if _, err := p.expect(TokenLBrace); err != nil {
		return nil, err
	}

	em := make(map[int]string)
	for p.current.Type != TokenRBrace {
		keyTok, err := p.expect(TokenInt)
		if err != nil {
			return nil, err
		}
		key, err := ParseIntLiteral(keyTok.Value)
		if err != nil {
			return nil, p.syntaxError(fmt.Sprintf("invalid enum key %q", keyTok.Value))
		}

		if _, err := p.expect(TokenEquals); err != nil {
			return nil, err
		}

		valTok, err := p.expect(TokenString)
		if err != nil {
			return nil, err
		}

		em[int(key)] = valTok.Value

		// Optional trailing comma
		if p.current.Type == TokenComma {
			if err := p.advance(); err != nil {
				return nil, err
			}
		}
	}

	if _, err := p.expect(TokenRBrace); err != nil {
		return nil, err
	}

	return em, nil
}

// parseWhen parses: when <ident> <op> <int>
// op is one of: >, <, >=, <=, ==, !=
func (p *PDLParser) parseWhen() (*schema.ConditionExpr, error) {
	if err := p.advance(); err != nil { // consume 'when'
		return nil, err
	}

	fieldTok, err := p.expectIdent()
	if err != nil {
		return nil, err
	}

	// Parse operator
	var op string
	switch p.current.Type {
	case TokenGreater:
		op = ">"
	case TokenLess:
		op = "<"
	case TokenGreaterEqual:
		op = ">="
	case TokenLessEqual:
		op = "<="
	case TokenEqualEqual:
		op = "=="
	case TokenNotEqual:
		op = "!="
	default:
		return nil, p.syntaxError(fmt.Sprintf("expected comparison operator, got %s (%q)", p.current.Type, p.current.Value))
	}
	if err := p.advance(); err != nil { // consume operator
		return nil, err
	}

	valTok, err := p.expect(TokenInt)
	if err != nil {
		return nil, err
	}
	val, err := ParseIntLiteral(valTok.Value)
	if err != nil {
		return nil, p.syntaxError(fmt.Sprintf("invalid condition value %q", valTok.Value))
	}

	return &schema.ConditionExpr{
		FieldName: fieldTok.Value,
		Operator:  op,
		Value:     int(val),
	}, nil
}

// parseDisplay parses: display <ident>
func (p *PDLParser) parseDisplay() (string, error) {
	if err := p.advance(); err != nil { // consume 'display'
		return "", err
	}

	fmtTok, err := p.expectIdent()
	if err != nil {
		return "", err
	}

	return fmtTok.Value, nil
}

// parseBitfield parses: bitfield { field ...; field ...; }
func (p *PDLParser) parseBitfield() (*schema.FieldDef, error) {
	if err := p.advance(); err != nil { // consume 'bitfield'
		return nil, err
	}

	if _, err := p.expect(TokenLBrace); err != nil {
		return nil, err
	}

	group := &schema.FieldDef{
		IsBitfieldGroup: true,
	}

	for p.current.Type != TokenRBrace && p.current.Type != TokenEOF {
		if p.current.Type != TokenField {
			return nil, p.syntaxError(fmt.Sprintf("expected field inside bitfield, got %s (%q)", p.current.Type, p.current.Value))
		}
		fd, err := p.parseField()
		if err != nil {
			return nil, err
		}
		group.BitfieldFields = append(group.BitfieldFields, *fd)
	}

	if _, err := p.expect(TokenRBrace); err != nil {
		return nil, err
	}

	return group, nil
}

// markPayload marks the last bytes field without a length_ref as payload.
// Payload fields have BitWidth=0 and no LengthRef, meaning they consume remaining bytes.
func (p *PDLParser) markPayload(ps *schema.ProtocolSchema) {
	// The last bytes field without length_ref is already naturally a payload field:
	// it has Type=Bytes, BitWidth=0, and LengthRef=nil.
	// No additional marking is needed since the codec will treat it as consuming remaining bytes.
	// This method exists as a hook for any future payload-specific logic.
}

// validate performs semantic validation on a parsed ProtocolSchema.
// It checks field references, bit widths, checksum algorithms, and display formats.
func (p *PDLParser) validate(ps *schema.ProtocolSchema) error {
	// Build a set of all defined field names (including bitfield sub-fields).
	defined := make(map[string]bool)
	for _, f := range ps.Fields {
		if f.IsBitfieldGroup {
			for _, bf := range f.BitfieldFields {
				defined[bf.Name] = true
			}
		} else {
			defined[f.Name] = true
		}
	}

	// Validate each field (and bitfield sub-fields).
	for _, f := range ps.Fields {
		if f.IsBitfieldGroup {
			for _, bf := range f.BitfieldFields {
				if err := p.validateField(&bf, defined); err != nil {
					return err
				}
			}
		} else {
			if err := p.validateField(&f, defined); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateField validates a single field definition.
func (p *PDLParser) validateField(fd *schema.FieldDef, defined map[string]bool) error {
	// 1. Bit width validation: must be in range 1-64 if explicitly set.
	if fd.BitWidth > 64 {
		return &errors.PDLSemanticError{
			FieldName: fd.Name,
			Message:   fmt.Sprintf("bit width %d out of valid range 1-64", fd.BitWidth),
		}
	}

	// 2. LengthRef field reference validation.
	if fd.LengthRef != nil {
		if !defined[fd.LengthRef.FieldName] {
			return &errors.PDLSemanticError{
				FieldName: fd.Name,
				Message:   fmt.Sprintf("length_ref references undefined field %q", fd.LengthRef.FieldName),
			}
		}
	}

	// 3. Checksum covers field reference validation.
	if fd.Checksum != nil {
		for _, cover := range fd.Checksum.CoverFields {
			if strings.Contains(cover, "..") {
				// Range notation: "field1..field2"
				parts := strings.SplitN(cover, "..", 2)
				if !defined[parts[0]] {
					return &errors.PDLSemanticError{
						FieldName: fd.Name,
						Message:   fmt.Sprintf("checksum covers references undefined field %q", parts[0]),
					}
				}
				if !defined[parts[1]] {
					return &errors.PDLSemanticError{
						FieldName: fd.Name,
						Message:   fmt.Sprintf("checksum covers references undefined field %q", parts[1]),
					}
				}
			} else {
				if !defined[cover] {
					return &errors.PDLSemanticError{
						FieldName: fd.Name,
						Message:   fmt.Sprintf("checksum covers references undefined field %q", cover),
					}
				}
			}
		}

		// Checksum algorithm validation.
		if p.checksumRegistry != nil {
			if !p.checksumRegistry.Has(fd.Checksum.Algorithm) {
				return &errors.PDLSemanticError{
					FieldName: fd.Name,
					Message:   fmt.Sprintf("unknown checksum algorithm: %s", fd.Checksum.Algorithm),
				}
			}
		}
	}

	// 4. When condition field reference validation.
	if fd.Condition != nil {
		if !defined[fd.Condition.FieldName] {
			return &errors.PDLSemanticError{
				FieldName: fd.Name,
				Message:   fmt.Sprintf("when condition references undefined field %q", fd.Condition.FieldName),
			}
		}
	}

	// 5. Display format validation.
	if fd.DisplayFormat != "" && p.formatRegistry != nil {
		if !p.formatRegistry.Has(fd.DisplayFormat) {
			return &errors.PDLSemanticError{
				FieldName: fd.Name,
				Message:   fmt.Sprintf("unknown display format: %s", fd.DisplayFormat),
			}
		}
	}

	return nil
}
