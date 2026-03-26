package pdl

import (
	"fmt"
	"strings"

	"github.com/hsqbyte/protospec/src/core/schema"
)

// ParseMessage parses a message PSL source and returns a MessageSchema.
func (p *PDLParser) ParseMessage(source string) (*schema.MessageSchema, error) {
	p.source = source
	p.lexer = NewLexer(source)
	p.dynamicKeywords = nil // reset per-session registry
	if err := p.advance(); err != nil {
		return nil, err
	}
	return p.parseMessageProtocol()
}

func (p *PDLParser) parseMessageProtocol() (*schema.MessageSchema, error) {
	if p.current.Type != TokenMessage {
		return nil, p.syntaxError("expected 'message'")
	}
	if err := p.advance(); err != nil {
		return nil, err
	}

	nameTok, err := p.expectIdent()
	if err != nil {
		return nil, err
	}

	// version "x.y"
	if p.current.Type != TokenVersion {
		return nil, p.syntaxError("expected 'version'")
	}
	if err := p.advance(); err != nil {
		return nil, err
	}
	if p.current.Type != TokenString {
		return nil, p.syntaxError("expected version string")
	}
	version := p.current.Value
	if err := p.advance(); err != nil {
		return nil, err
	}

	if _, err := p.expect(TokenLBrace); err != nil {
		return nil, err
	}

	ms := &schema.MessageSchema{
		Name:    nameTok.Value,
		Version: version,
	}

	for p.current.Type != TokenRBrace && p.current.Type != TokenEOF {
		switch p.current.Type {
		case TokenTransport:
			if err := p.advance(); err != nil {
				return nil, err
			}
			// Parse transport reference: could be "jsonrpc" or "jsonrpc@2.0"
			// The lexer will produce an ident token. If followed by no semicolon
			// we need to handle the @ syntax. But since @ is not a valid ident char,
			// the lexer will just produce the name part. We need to handle the
			// name@version syntax by reading the raw ident value.
			tok, err := p.expectIdent()
			if err != nil {
				return nil, err
			}
			ref := ParseTransportRef(tok.Value)
			ms.Transport = ref.Name
			ms.TransportVer = ref.Version

			if _, err := p.expect(TokenSemicolon); err != nil {
				return nil, err
			}

			// If transport loader is set, load the transport and build dynamic keywords
			if p.transportLoader != nil {
				td, loadErr := p.transportLoader.LoadTransport(ref.Name, ref.Version)
				if loadErr == nil {
					ms.TransportDef = td
					p.dynamicKeywords = NewDynamicKeywordRegistry(td)
				}
				// If transport not found, fall back to legacy mode silently
			}
		default:
			// Parse message body entries
			if p.dynamicKeywords != nil {
				// Dynamic mode: check if current token is a transport-defined message type
				if err := p.parseDynamicMessageBody(ms); err != nil {
					return nil, err
				}
			} else {
				// Legacy mode: hardcoded request/response/notification
				if err := p.parseLegacyMessageBody(ms); err != nil {
					return nil, err
				}
			}
		}
	}

	if _, err := p.expect(TokenRBrace); err != nil {
		return nil, err
	}

	return ms, nil
}

// parseLegacyMessageBody handles one iteration of the legacy hardcoded message body loop.
func (p *PDLParser) parseLegacyMessageBody(ms *schema.MessageSchema) error {
	switch p.current.Type {
	case TokenRequest:
		msg, err := p.parseMessageDef("request")
		if err != nil {
			return err
		}
		ms.Messages = append(ms.Messages, *msg)
	case TokenResponse:
		msg, err := p.parseMessageDef("response")
		if err != nil {
			return err
		}
		ms.Messages = append(ms.Messages, *msg)
	case TokenNotify:
		msg, err := p.parseMessageDef("notification")
		if err != nil {
			return err
		}
		ms.Messages = append(ms.Messages, *msg)
	default:
		return p.syntaxError(fmt.Sprintf("unexpected token %s in message body", p.current.Type))
	}
	return nil
}

// parseDynamicMessageBody handles one iteration of the dynamic message body loop.
func (p *PDLParser) parseDynamicMessageBody(ms *schema.MessageSchema) error {
	tokenValue := p.current.Value
	typeDef, ok := p.dynamicKeywords.GetMessageType(tokenValue)
	if !ok {
		validTypes := p.dynamicKeywords.MessageTypeNames()
		return p.syntaxError(fmt.Sprintf("unexpected token %q: not a valid message type for transport (valid types: %s)",
			tokenValue, strings.Join(validTypes, ", ")))
	}
	if err := p.advance(); err != nil { // consume the message type keyword
		return err
	}
	msg, err := p.parseDynamicMessageDef(typeDef)
	if err != nil {
		return err
	}
	ms.Messages = append(ms.Messages, *msg)
	return nil
}

func (p *PDLParser) parseMessageDef(kind string) (*schema.MessageDef, error) {
	if err := p.advance(); err != nil { // consume request/response/notification
		return nil, err
	}

	nameTok, err := p.expectIdent()
	if err != nil {
		return nil, err
	}

	msg := &schema.MessageDef{
		Name: nameTok.Value,
		Kind: kind,
	}

	// notification may have no body (just a semicolon)
	if p.current.Type == TokenSemicolon {
		if err := p.advance(); err != nil {
			return nil, err
		}
		return msg, nil
	}

	if _, err := p.expect(TokenLBrace); err != nil {
		return nil, err
	}

	for p.current.Type != TokenRBrace && p.current.Type != TokenEOF {
		if p.current.Type != TokenField {
			return nil, p.syntaxError(fmt.Sprintf("expected 'field' in %s %s, got %s", kind, msg.Name, p.current.Type))
		}
		fd, err := p.parseMessageField()
		if err != nil {
			return nil, err
		}
		msg.Fields = append(msg.Fields, *fd)
	}

	if _, err := p.expect(TokenRBrace); err != nil {
		return nil, err
	}

	return msg, nil
}

// parseDynamicMessageDef parses a message block using transport-defined rules.
// The lexer is positioned after the message type keyword (e.g., after "request").
// Expected syntax:
//
//	{
//	  method "initialize";
//	  params { field ...; }
//	  response { result { field ...; } }
//	}
func (p *PDLParser) parseDynamicMessageDef(typeDef *schema.MessageTypeDef) (*schema.MessageDef, error) {
	msg := &schema.MessageDef{
		Kind: typeDef.Name,
	}

	if _, err := p.expect(TokenLBrace); err != nil {
		return nil, err
	}

	// Track which transport fields have been provided
	providedFields := make(map[string]bool)

	for p.current.Type != TokenRBrace && p.current.Type != TokenEOF {
		tokenValue := p.current.Value

		// Check for "response" keyword
		if p.current.Type == TokenResponse || tokenValue == "response" {
			if typeDef.ResponseDef == nil {
				return nil, p.syntaxError(fmt.Sprintf("message type %q does not support response blocks", typeDef.Name))
			}
			if err := p.advance(); err != nil { // consume "response"
				return nil, err
			}
			responseDef, err := p.parseDynamicResponseBlock(typeDef.ResponseDef)
			if err != nil {
				return nil, err
			}
			msg.Response = responseDef
			continue
		}

		// Check if token is a transport-defined field keyword
		fieldDef := findTransportField(typeDef.Fields, tokenValue)
		if fieldDef == nil {
			// Build list of valid field names for error message
			var validNames []string
			for _, f := range typeDef.Fields {
				validNames = append(validNames, f.Name)
			}
			if typeDef.ResponseDef != nil {
				validNames = append(validNames, "response")
			}
			return nil, p.syntaxError(fmt.Sprintf("unexpected token %q in %s block (valid keywords: %s)",
				tokenValue, typeDef.Name, strings.Join(validNames, ", ")))
		}

		if err := p.advance(); err != nil { // consume the field keyword
			return nil, err
		}

		// Parse the field value based on its type
		switch fieldDef.Type {
		case schema.MsgString:
			// Expect a string literal value: method "initialize";
			if p.current.Type != TokenString {
				return nil, p.syntaxError(fmt.Sprintf("expected string value for %q, got %s", fieldDef.Name, p.current.Type))
			}
			// Store as a field with the string value as the name
			msg.Fields = append(msg.Fields, schema.MessageFieldDef{
				Name: p.current.Value,
				Type: schema.MsgString,
			})
			if err := p.advance(); err != nil {
				return nil, err
			}
			if _, err := p.expect(TokenSemicolon); err != nil {
				return nil, err
			}
		case schema.MsgNumber:
			// Expect a number literal value: opcode 0x0006;
			if p.current.Type != TokenInt {
				return nil, p.syntaxError(fmt.Sprintf("expected number value for %q, got %s", fieldDef.Name, p.current.Type))
			}
			msg.Fields = append(msg.Fields, schema.MessageFieldDef{
				Name: p.current.Value,
				Type: schema.MsgNumber,
			})
			if err := p.advance(); err != nil {
				return nil, err
			}
			if _, err := p.expect(TokenSemicolon); err != nil {
				return nil, err
			}
		case schema.MsgObject:
			// Expect a block with field declarations: params { field ...; }
			if p.current.Type != TokenLBrace {
				return nil, p.syntaxError(fmt.Sprintf("expected '{' for object field %q, got %s", fieldDef.Name, p.current.Type))
			}
			if err := p.advance(); err != nil { // consume '{'
				return nil, err
			}
			objField := schema.MessageFieldDef{
				Name: fieldDef.Name,
				Type: schema.MsgObject,
			}
			for p.current.Type != TokenRBrace && p.current.Type != TokenEOF {
				if p.current.Type != TokenField {
					return nil, p.syntaxError(fmt.Sprintf("expected 'field' in %s block, got %s", fieldDef.Name, p.current.Type))
				}
				fd, err := p.parseMessageField()
				if err != nil {
					return nil, err
				}
				objField.Fields = append(objField.Fields, *fd)
			}
			if _, err := p.expect(TokenRBrace); err != nil {
				return nil, err
			}
			msg.Fields = append(msg.Fields, objField)
		case schema.MsgBoolean:
			// Expect a boolean-like identifier: enabled true;
			msg.Fields = append(msg.Fields, schema.MessageFieldDef{
				Name: fieldDef.Name,
				Type: schema.MsgBoolean,
			})
			if isIdentLike(p.current.Type) {
				if err := p.advance(); err != nil {
					return nil, err
				}
			}
			if _, err := p.expect(TokenSemicolon); err != nil {
				return nil, err
			}
		case schema.MsgArray:
			msg.Fields = append(msg.Fields, schema.MessageFieldDef{
				Name: fieldDef.Name,
				Type: schema.MsgArray,
			})
			if _, err := p.expect(TokenSemicolon); err != nil {
				return nil, err
			}
		default:
			return nil, p.syntaxError(fmt.Sprintf("unsupported field type for %q", fieldDef.Name))
		}

		providedFields[fieldDef.Name] = true
	}

	if _, err := p.expect(TokenRBrace); err != nil {
		return nil, err
	}

	// Validate required fields: fields without default and not optional must be provided
	for _, f := range typeDef.Fields {
		if f.DefaultValue == nil && !f.AutoValue && !f.Optional {
			if !providedFields[f.Name] {
				return nil, p.syntaxError(fmt.Sprintf("required field %q missing in %s block", f.Name, typeDef.Name))
			}
		}
	}

	return msg, nil
}

// parseDynamicResponseBlock parses a nested response block inside a dynamic message.
// Expected syntax: { result { field ...; } }
func (p *PDLParser) parseDynamicResponseBlock(responseDef *schema.MessageTypeDef) (*schema.MessageDef, error) {
	resp := &schema.MessageDef{
		Kind: "response",
	}

	if _, err := p.expect(TokenLBrace); err != nil {
		return nil, err
	}

	providedFields := make(map[string]bool)

	for p.current.Type != TokenRBrace && p.current.Type != TokenEOF {
		tokenValue := p.current.Value

		fieldDef := findTransportField(responseDef.Fields, tokenValue)
		if fieldDef == nil {
			var validNames []string
			for _, f := range responseDef.Fields {
				validNames = append(validNames, f.Name)
			}
			return nil, p.syntaxError(fmt.Sprintf("unexpected token %q in response block (valid keywords: %s)",
				tokenValue, strings.Join(validNames, ", ")))
		}

		if err := p.advance(); err != nil { // consume the field keyword
			return nil, err
		}

		switch fieldDef.Type {
		case schema.MsgString:
			if p.current.Type != TokenString {
				return nil, p.syntaxError(fmt.Sprintf("expected string value for %q, got %s", fieldDef.Name, p.current.Type))
			}
			resp.Fields = append(resp.Fields, schema.MessageFieldDef{
				Name: p.current.Value,
				Type: schema.MsgString,
			})
			if err := p.advance(); err != nil {
				return nil, err
			}
			if _, err := p.expect(TokenSemicolon); err != nil {
				return nil, err
			}
		case schema.MsgNumber:
			if p.current.Type != TokenInt {
				return nil, p.syntaxError(fmt.Sprintf("expected number value for %q, got %s", fieldDef.Name, p.current.Type))
			}
			resp.Fields = append(resp.Fields, schema.MessageFieldDef{
				Name: p.current.Value,
				Type: schema.MsgNumber,
			})
			if err := p.advance(); err != nil {
				return nil, err
			}
			if _, err := p.expect(TokenSemicolon); err != nil {
				return nil, err
			}
		case schema.MsgObject:
			if p.current.Type != TokenLBrace {
				return nil, p.syntaxError(fmt.Sprintf("expected '{' for object field %q, got %s", fieldDef.Name, p.current.Type))
			}
			if err := p.advance(); err != nil {
				return nil, err
			}
			objField := schema.MessageFieldDef{
				Name: fieldDef.Name,
				Type: schema.MsgObject,
			}
			for p.current.Type != TokenRBrace && p.current.Type != TokenEOF {
				if p.current.Type != TokenField {
					return nil, p.syntaxError(fmt.Sprintf("expected 'field' in %s block, got %s", fieldDef.Name, p.current.Type))
				}
				fd, err := p.parseMessageField()
				if err != nil {
					return nil, err
				}
				objField.Fields = append(objField.Fields, *fd)
			}
			if _, err := p.expect(TokenRBrace); err != nil {
				return nil, err
			}
			resp.Fields = append(resp.Fields, objField)
		case schema.MsgBoolean:
			resp.Fields = append(resp.Fields, schema.MessageFieldDef{
				Name: fieldDef.Name,
				Type: schema.MsgBoolean,
			})
			if isIdentLike(p.current.Type) {
				if err := p.advance(); err != nil {
					return nil, err
				}
			}
			if _, err := p.expect(TokenSemicolon); err != nil {
				return nil, err
			}
		case schema.MsgArray:
			resp.Fields = append(resp.Fields, schema.MessageFieldDef{
				Name: fieldDef.Name,
				Type: schema.MsgArray,
			})
			if _, err := p.expect(TokenSemicolon); err != nil {
				return nil, err
			}
		default:
			return nil, p.syntaxError(fmt.Sprintf("unsupported field type for %q", fieldDef.Name))
		}

		providedFields[fieldDef.Name] = true
	}

	if _, err := p.expect(TokenRBrace); err != nil {
		return nil, err
	}

	// Validate required fields in response
	for _, f := range responseDef.Fields {
		if f.DefaultValue == nil && !f.AutoValue && !f.Optional {
			if !providedFields[f.Name] {
				return nil, p.syntaxError(fmt.Sprintf("required field %q missing in response block", f.Name))
			}
		}
	}

	return resp, nil
}

// findTransportField looks up a field by name in a slice of TransportFieldDef.
func findTransportField(fields []schema.TransportFieldDef, name string) *schema.TransportFieldDef {
	for i := range fields {
		if fields[i].Name == name {
			return &fields[i]
		}
	}
	return nil
}

func (p *PDLParser) parseMessageField() (*schema.MessageFieldDef, error) {
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

	fd := &schema.MessageFieldDef{
		Name: nameTok.Value,
	}

	// Parse type
	if err := p.parseMessageFieldType(fd); err != nil {
		return nil, err
	}

	// Check for 'optional'
	if p.current.Type == TokenOptional {
		fd.Optional = true
		if err := p.advance(); err != nil {
			return nil, err
		}
	}

	if _, err := p.expect(TokenSemicolon); err != nil {
		return nil, err
	}

	return fd, nil
}

func (p *PDLParser) parseMessageFieldType(fd *schema.MessageFieldDef) error {
	switch p.current.Value {
	case "string":
		fd.Type = schema.MsgString
		if err := p.advance(); err != nil {
			return err
		}
	case "number":
		fd.Type = schema.MsgNumber
		if err := p.advance(); err != nil {
			return err
		}
	case "boolean":
		fd.Type = schema.MsgBoolean
		if err := p.advance(); err != nil {
			return err
		}
	case "object":
		fd.Type = schema.MsgObject
		if err := p.advance(); err != nil {
			return err
		}
		// Object may have inline fields
		if p.current.Type == TokenLBrace {
			if err := p.advance(); err != nil {
				return err
			}
			for p.current.Type != TokenRBrace && p.current.Type != TokenEOF {
				if p.current.Type != TokenField {
					return p.syntaxError(fmt.Sprintf("expected 'field' in object, got %s", p.current.Type))
				}
				sub, err := p.parseMessageField()
				if err != nil {
					return err
				}
				fd.Fields = append(fd.Fields, *sub)
			}
			if _, err := p.expect(TokenRBrace); err != nil {
				return err
			}
		}
	case "array":
		fd.Type = schema.MsgArray
		if err := p.advance(); err != nil {
			return err
		}
	default:
		return p.syntaxError(fmt.Sprintf("unknown message field type %q", p.current.Value))
	}
	return nil
}
