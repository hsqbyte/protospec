package pdl

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hsqbyte/protospec/src/core/errors"
	"github.com/hsqbyte/protospec/src/core/schema"
)

// TransportLoader loads and caches transport PSL definitions.
type TransportLoader struct {
	cache       map[string]*schema.TransportDef
	embedFS     fs.FS
	searchPaths []string
}

// NewTransportLoader creates a new TransportLoader with the given embedded FS and search paths.
func NewTransportLoader(embedFS fs.FS, searchPaths []string) *TransportLoader {
	return &TransportLoader{
		cache:       make(map[string]*schema.TransportDef),
		embedFS:     embedFS,
		searchPaths: searchPaths,
	}
}

// cacheKey returns the cache key for a transport name and optional version.
func cacheKey(name string, version *string) string {
	if version != nil {
		return name + "@" + *version
	}
	return name
}

// LoadTransport loads a transport by name and optional version.
// Returns cached result if already loaded.
func (tl *TransportLoader) LoadTransport(name string, version *string) (*schema.TransportDef, error) {
	key := cacheKey(name, version)
	if td, ok := tl.cache[key]; ok {
		return td, nil
	}

	source, err := tl.resolveAndRead(name, version)
	if err != nil {
		return nil, err
	}

	td, err := ParseTransportPSL(source)
	if err != nil {
		return nil, err
	}

	tl.cache[key] = td
	return td, nil
}

// ResolveTransportPath finds the PSL file path for a transport reference.
// Returns an error if the path contains directory traversal patterns.
func (tl *TransportLoader) ResolveTransportPath(name string, version *string) (string, error) {
	if strings.Contains(name, "..") {
		return "", fmt.Errorf("invalid transport name %q: directory traversal not allowed", name)
	}

	// Try embedded FS first.
	if tl.embedFS != nil {
		path, err := tl.resolveInFS(tl.embedFS, name, version)
		if err == nil {
			return path, nil
		}
	}

	// Try filesystem search paths.
	for _, base := range tl.searchPaths {
		dir := filepath.Join(base, name)
		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			continue
		}

		if version != nil {
			// Look for versioned path: <name>/<version>/<name>.psl
			versionedDir := filepath.Join(dir, *version)
			pslPath := filepath.Join(versionedDir, name+".psl")
			if _, err := os.Stat(pslPath); err == nil {
				return pslPath, nil
			}
		} else {
			// Try version resolution.
			resolved, err := ResolveVersion(dir, nil)
			if err == nil && resolved != dir {
				// Versioned directory found.
				pslPath := filepath.Join(resolved, name+".psl")
				if _, err := os.Stat(pslPath); err == nil {
					return pslPath, nil
				}
			}
		}

		// Flat fallback: <name>/<name>.psl
		pslPath := filepath.Join(dir, name+".psl")
		if _, err := os.Stat(pslPath); err == nil {
			return pslPath, nil
		}
	}

	return "", fmt.Errorf("transport %q not found", name)
}

// resolveInFS resolves a transport path within an fs.FS.
func (tl *TransportLoader) resolveInFS(fsys fs.FS, name string, version *string) (string, error) {
	if version != nil {
		// Versioned: <name>/<version>/<name>.psl
		path := name + "/" + *version + "/" + name + ".psl"
		if _, err := fs.Stat(fsys, path); err == nil {
			return path, nil
		}
	} else {
		// Try to find version subdirectories.
		entries, err := fs.ReadDir(fsys, name)
		if err == nil {
			var versions []string
			for _, e := range entries {
				if e.IsDir() && isVersionDir(e.Name()) {
					versions = append(versions, e.Name())
				}
			}
			if len(versions) > 0 {
				// Sort and pick latest.
				sortVersions(versions)
				latest := versions[len(versions)-1]
				path := name + "/" + latest + "/" + name + ".psl"
				if _, err := fs.Stat(fsys, path); err == nil {
					return path, nil
				}
			}
		}
	}

	// Flat fallback: <name>/<name>.psl
	path := name + "/" + name + ".psl"
	if _, err := fs.Stat(fsys, path); err == nil {
		return path, nil
	}

	return "", fmt.Errorf("transport %q not found in embedded FS", name)
}

// resolveAndRead resolves a transport and reads its PSL source.
func (tl *TransportLoader) resolveAndRead(name string, version *string) (string, error) {
	if strings.Contains(name, "..") {
		return "", fmt.Errorf("invalid transport name %q: directory traversal not allowed", name)
	}

	// Try embedded FS first.
	if tl.embedFS != nil {
		path, err := tl.resolveInFS(tl.embedFS, name, version)
		if err == nil {
			data, readErr := fs.ReadFile(tl.embedFS, path)
			if readErr == nil {
				return string(data), nil
			}
		}
	}

	// Try filesystem search paths.
	for _, base := range tl.searchPaths {
		dir := filepath.Join(base, name)
		info, err := os.Stat(dir)
		if err != nil || !info.IsDir() {
			continue
		}

		if version != nil {
			versionedDir := filepath.Join(dir, *version)
			pslPath := filepath.Join(versionedDir, name+".psl")
			data, err := os.ReadFile(pslPath)
			if err == nil {
				return string(data), nil
			}
		} else {
			resolved, err := ResolveVersion(dir, nil)
			if err == nil && resolved != dir {
				pslPath := filepath.Join(resolved, name+".psl")
				data, err := os.ReadFile(pslPath)
				if err == nil {
					return string(data), nil
				}
			}
		}

		// Flat fallback.
		pslPath := filepath.Join(dir, name+".psl")
		data, err := os.ReadFile(pslPath)
		if err == nil {
			return string(data), nil
		}
	}

	return "", fmt.Errorf("transport %q not found", name)
}

// sortVersions sorts version strings by semver ordering.
func sortVersions(versions []string) {
	for i := 1; i < len(versions); i++ {
		for j := i; j > 0 && compareSemver(versions[j-1], versions[j]) > 0; j-- {
			versions[j-1], versions[j] = versions[j], versions[j-1]
		}
	}
}

// ParseTransportPSL parses a transport PSL source into a TransportDef.
// The expected format is:
//
//	message <Name> version "<ver>" {
//	  <messageType> {
//	    field <name>: <type> [optional] [default <value>|auto];
//	    [response { ... }]
//	  }
//	}
func ParseTransportPSL(source string) (*schema.TransportDef, error) {
	p := &transportParser{
		lexer:  NewLexer(source),
		source: source,
	}
	if err := p.advance(); err != nil {
		return nil, err
	}
	return p.parse()
}

// transportParser is a recursive-descent parser for transport PSL files.
type transportParser struct {
	lexer   *Lexer
	current Token
	source  string
}

func (p *transportParser) advance() error {
	tok, err := p.lexer.NextToken()
	if err != nil {
		return err
	}
	p.current = tok
	return nil
}

func (p *transportParser) expect(tt TokenType) (Token, error) {
	if p.current.Type != tt {
		return Token{}, p.syntaxError(fmt.Sprintf("expected %s, got %s (%q)", tt, p.current.Type, p.current.Value))
	}
	tok := p.current
	if err := p.advance(); err != nil {
		return Token{}, err
	}
	return tok, nil
}

func (p *transportParser) syntaxError(msg string) *errors.PDLSyntaxError {
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

// parse parses the top-level transport PSL: message <Name> version "<ver>" { ... }
func (p *transportParser) parse() (*schema.TransportDef, error) {
	// Expect "message"
	if p.current.Type != TokenMessage {
		return nil, p.syntaxError("expected 'message'")
	}
	if err := p.advance(); err != nil {
		return nil, err
	}

	// Parse transport name (identifier-like token).
	nameTok, err := p.expectIdent()
	if err != nil {
		return nil, err
	}

	// Expect "version"
	if p.current.Type != TokenVersion {
		return nil, p.syntaxError("expected 'version'")
	}
	if err := p.advance(); err != nil {
		return nil, err
	}

	// Expect version string.
	if p.current.Type != TokenString {
		return nil, p.syntaxError("expected version string")
	}
	version := p.current.Value
	if err := p.advance(); err != nil {
		return nil, err
	}

	// Expect "{"
	if _, err := p.expect(TokenLBrace); err != nil {
		return nil, err
	}

	td := &schema.TransportDef{
		Name:    nameTok.Value,
		Version: version,
	}

	// Parse message types until "}"
	for p.current.Type != TokenRBrace && p.current.Type != TokenEOF {
		mtd, err := p.parseMessageType()
		if err != nil {
			return nil, err
		}
		td.MessageTypes = append(td.MessageTypes, *mtd)
	}

	if _, err := p.expect(TokenRBrace); err != nil {
		return nil, err
	}

	return td, nil
}

// expectIdent expects an identifier-like token (including keywords used as names).
func (p *transportParser) expectIdent() (Token, error) {
	if !isIdentLike(p.current.Type) {
		return Token{}, p.syntaxError(fmt.Sprintf("expected identifier, got %s (%q)", p.current.Type, p.current.Value))
	}
	tok := p.current
	if err := p.advance(); err != nil {
		return Token{}, err
	}
	return tok, nil
}

// parseMessageType parses a message type block: <typeName> { field ...; [response { ... }] }
func (p *transportParser) parseMessageType() (*schema.MessageTypeDef, error) {
	// The current token should be an identifier representing the message type name
	// (e.g., "request", "notification", "command", "event", "data").
	typeTok, err := p.expectIdent()
	if err != nil {
		return nil, err
	}

	mtd := &schema.MessageTypeDef{
		Name: typeTok.Value,
	}

	if _, err := p.expect(TokenLBrace); err != nil {
		return nil, err
	}

	// Parse fields and nested response blocks until "}"
	for p.current.Type != TokenRBrace && p.current.Type != TokenEOF {
		if p.current.Type == TokenField {
			field, err := p.parseTransportField()
			if err != nil {
				return nil, err
			}
			mtd.Fields = append(mtd.Fields, *field)
		} else if p.current.Type == TokenResponse {
			// Nested response block.
			if err := p.advance(); err != nil {
				return nil, err
			}
			responseDef, err := p.parseResponseBlock()
			if err != nil {
				return nil, err
			}
			mtd.ResponseDef = responseDef
		} else {
			return nil, p.syntaxError(fmt.Sprintf("expected 'field' or 'response' in message type %q, got %s (%q)", mtd.Name, p.current.Type, p.current.Value))
		}
	}

	if _, err := p.expect(TokenRBrace); err != nil {
		return nil, err
	}

	return mtd, nil
}

// parseResponseBlock parses a nested response block: { field ...; }
func (p *transportParser) parseResponseBlock() (*schema.MessageTypeDef, error) {
	if _, err := p.expect(TokenLBrace); err != nil {
		return nil, err
	}

	rd := &schema.MessageTypeDef{
		Name: "response",
	}

	for p.current.Type != TokenRBrace && p.current.Type != TokenEOF {
		if p.current.Type != TokenField {
			return nil, p.syntaxError(fmt.Sprintf("expected 'field' in response block, got %s (%q)", p.current.Type, p.current.Value))
		}
		field, err := p.parseTransportField()
		if err != nil {
			return nil, err
		}
		rd.Fields = append(rd.Fields, *field)
	}

	if _, err := p.expect(TokenRBrace); err != nil {
		return nil, err
	}

	return rd, nil
}

// parseTransportField parses: field <name>: <type> [{ ... }] [optional] [default <value>|auto];
func (p *transportParser) parseTransportField() (*schema.TransportFieldDef, error) {
	// Consume "field"
	if err := p.advance(); err != nil {
		return nil, err
	}

	nameTok, err := p.expectIdent()
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(TokenColon); err != nil {
		return nil, err
	}

	fd := &schema.TransportFieldDef{
		Name: nameTok.Value,
	}

	// Parse type.
	if err := p.parseTransportFieldType(fd); err != nil {
		return nil, err
	}

	// Parse optional modifiers: "optional", "default <value>|auto"
	// These can appear in any order before the semicolon.
	for p.current.Type != TokenSemicolon && p.current.Type != TokenEOF {
		if p.current.Type == TokenOptional {
			fd.Optional = true
			if err := p.advance(); err != nil {
				return nil, err
			}
		} else if p.current.Type == TokenIdent && p.current.Value == "default" {
			if err := p.advance(); err != nil {
				return nil, err
			}
			// "default auto" or "default <literal>"
			if p.current.Type == TokenIdent && p.current.Value == "auto" {
				fd.AutoValue = true
				if err := p.advance(); err != nil {
					return nil, err
				}
			} else if p.current.Type == TokenString {
				fd.DefaultValue = p.current.Value
				if err := p.advance(); err != nil {
					return nil, err
				}
			} else if p.current.Type == TokenInt {
				val, parseErr := strconv.ParseFloat(p.current.Value, 64)
				if parseErr != nil {
					return nil, p.syntaxError(fmt.Sprintf("invalid default value %q", p.current.Value))
				}
				fd.DefaultValue = val
				if err := p.advance(); err != nil {
					return nil, err
				}
			} else {
				return nil, p.syntaxError(fmt.Sprintf("expected default value or 'auto', got %s (%q)", p.current.Type, p.current.Value))
			}
		} else if p.current.Type == TokenDefault {
			// TokenDefault is "=" in the grammar, but "default" as an identifier
			// is handled above. This branch handles the case where the lexer
			// might produce TokenDefault for "=".
			return nil, p.syntaxError(fmt.Sprintf("unexpected token %s in field declaration", p.current.Type))
		} else {
			return nil, p.syntaxError(fmt.Sprintf("unexpected token %s (%q) in field declaration", p.current.Type, p.current.Value))
		}
	}

	if _, err := p.expect(TokenSemicolon); err != nil {
		return nil, err
	}

	return fd, nil
}

// parseTransportFieldType parses the type portion of a transport field.
func (p *transportParser) parseTransportFieldType(fd *schema.TransportFieldDef) error {
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
		// Object may have inline nested fields.
		if p.current.Type == TokenLBrace {
			if err := p.advance(); err != nil {
				return err
			}
			for p.current.Type != TokenRBrace && p.current.Type != TokenEOF {
				if p.current.Type != TokenField {
					return p.syntaxError(fmt.Sprintf("expected 'field' in object, got %s (%q)", p.current.Type, p.current.Value))
				}
				sub, err := p.parseTransportField()
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
		return p.syntaxError(fmt.Sprintf("unknown transport field type %q", p.current.Value))
	}
	return nil
}
