package lsp

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hsqbyte/protospec/psl"
	"github.com/hsqbyte/protospec/src/core/pdl"
)

type textDocumentItem struct {
	URI  string `json:"uri"`
	Text string `json:"text"`
}

type didOpenParams struct {
	TextDocument textDocumentItem `json:"textDocument"`
}

type didChangeParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
	ContentChanges []struct {
		Text string `json:"text"`
	} `json:"contentChanges"`
}

type didCloseParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
}

type position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type textDocumentPosition struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
	Position position `json:"position"`
}

func (s *Server) handleDidOpen(msg *jsonrpcMessage) {
	var params didOpenParams
	json.Unmarshal(msg.Params, &params)
	s.docs[params.TextDocument.URI] = params.TextDocument.Text
	s.publishDiagnostics(params.TextDocument.URI)
}

func (s *Server) handleDidChange(msg *jsonrpcMessage) {
	var params didChangeParams
	json.Unmarshal(msg.Params, &params)
	if len(params.ContentChanges) > 0 {
		s.docs[params.TextDocument.URI] = params.ContentChanges[0].Text
	}
	s.publishDiagnostics(params.TextDocument.URI)
}

func (s *Server) handleDidClose(msg *jsonrpcMessage) {
	var params didCloseParams
	json.Unmarshal(msg.Params, &params)
	delete(s.docs, params.TextDocument.URI)
	// Clear diagnostics
	s.sendNotification("textDocument/publishDiagnostics", map[string]any{
		"uri":         params.TextDocument.URI,
		"diagnostics": []any{},
	})
}

func (s *Server) publishDiagnostics(uri string) {
	content, ok := s.docs[uri]
	if !ok {
		return
	}

	var diagnostics []any

	// Try parsing to find errors
	parser := pdl.NewPDLParser(nil, nil)

	// Set up transport loader with embedded PSL files for dynamic transport support
	loader := pdl.NewTransportLoader(psl.FS, nil)
	parser.SetTransportLoader(loader)

	// Detect if it's a message protocol
	if strings.Contains(content, "message ") && strings.Contains(content, "transport ") {
		_, err := parser.ParseMessage(content)
		if err != nil {
			diag := errorToDiagnostic(err)
			if diag != nil {
				diagnostics = append(diagnostics, diag)
			}
		}
	} else {
		_, err := parser.Parse(content)
		if err != nil {
			diag := errorToDiagnostic(err)
			if diag != nil {
				diagnostics = append(diagnostics, diag)
			}
		}
	}

	s.sendNotification("textDocument/publishDiagnostics", map[string]any{
		"uri":         uri,
		"diagnostics": diagnostics,
	})
}

func errorToDiagnostic(err error) map[string]any {
	msg := err.Error()
	line := 0
	col := 0

	// Try to extract line/col from error message
	// PDLSyntaxError has Line and Column fields
	if se, ok := err.(interface{ GetLine() int }); ok {
		line = se.GetLine() - 1
	}
	if se, ok := err.(interface{ GetColumn() int }); ok {
		col = se.GetColumn() - 1
	}

	return map[string]any{
		"range": map[string]any{
			"start": map[string]int{"line": line, "character": col},
			"end":   map[string]int{"line": line, "character": col + 10},
		},
		"severity": 1, // Error
		"source":   "psl",
		"message":  msg,
	}
}

func (s *Server) handleCompletion(msg *jsonrpcMessage) {
	var params textDocumentPosition
	json.Unmarshal(msg.Params, &params)

	content := s.docs[params.TextDocument.URI]
	line := getLine(content, params.Position.Line)

	var items []map[string]any

	// Context-aware completion
	trimmed := strings.TrimSpace(line)

	// Check if we're inside a message protocol with a transport declaration
	// and offer dynamic completions based on the transport definition.
	if transportName := extractTransportName(content); transportName != "" {
		loader := pdl.NewTransportLoader(psl.FS, nil)
		td, err := loader.LoadTransport(transportName, nil)
		if err == nil {
			registry := pdl.NewDynamicKeywordRegistry(td)

			// Determine context: are we inside a message type block or at the message body level?
			if msgType := findEnclosingMessageType(content, params.Position.Line, registry); msgType != "" {
				// Inside a message type block — offer transport-defined field names
				if mtd, ok := registry.GetMessageType(msgType); ok {
					for _, f := range mtd.Fields {
						items = append(items, map[string]any{
							"label":  f.Name,
							"kind":   6, // Variable
							"detail": "transport field: " + f.Type.String(),
						})
					}
					// If this message type supports response, offer "response" keyword
					if mtd.ResponseDef != nil {
						items = append(items, map[string]any{
							"label":  "response",
							"kind":   14, // Keyword
							"detail": "Nested response block",
						})
					}
				}
			} else {
				// At message body level — offer transport-defined message type names
				for _, name := range registry.MessageTypeNames() {
					items = append(items, map[string]any{
						"label":  name,
						"kind":   14, // Keyword
						"detail": "transport message type",
					})
				}
			}

			// Add default and auto keywords
			items = append(items, map[string]any{
				"label": "default", "kind": 14, "detail": "Default value annotation",
			})
			items = append(items, map[string]any{
				"label": "auto", "kind": 14, "detail": "Auto-generated value",
			})

			s.sendResponse(msg.ID, map[string]any{"items": items})
			return
		}
	}

	if trimmed == "" || strings.HasPrefix(trimmed, "//") {
		// Top-level keywords
		for _, kw := range []string{"protocol", "message"} {
			items = append(items, map[string]any{
				"label": kw, "kind": 14, // Keyword
			})
		}
	} else if strings.Contains(trimmed, ":") && !strings.Contains(trimmed, ";") {
		// After colon — type completion
		for _, t := range []string{"uint8", "uint16", "uint32", "uint64", "int8", "int16", "int32", "bytes", "string", "bool"} {
			items = append(items, map[string]any{
				"label": t, "kind": 6, // Variable
			})
		}
	} else {
		// Body keywords
		for _, kw := range []string{"field", "bitfield", "const", "byte_order", "checksum", "length_ref", "when", "display", "enum", "range", "default", "auto"} {
			items = append(items, map[string]any{
				"label": kw, "kind": 14,
			})
		}
	}

	s.sendResponse(msg.ID, map[string]any{"items": items})
}

// extractTransportName extracts the transport name from a PSL document content.
// Returns empty string if no transport declaration is found.
func extractTransportName(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "transport ") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				name := strings.TrimSuffix(parts[1], ";")
				// Strip version pin for loading
				if idx := strings.Index(name, "@"); idx > 0 {
					name = name[:idx]
				}
				return name
			}
		}
	}
	return ""
}

// findEnclosingMessageType determines if the cursor is inside a message type block.
// Returns the message type name (e.g., "request", "notification") or empty string.
func findEnclosingMessageType(content string, cursorLine int, registry *pdl.DynamicKeywordRegistry) string {
	lines := strings.Split(content, "\n")
	depth := 0
	currentType := ""

	for i, line := range lines {
		if i >= cursorLine {
			break
		}
		trimmed := strings.TrimSpace(line)

		// Track brace depth to determine scope
		for _, ch := range trimmed {
			if ch == '{' {
				depth++
			} else if ch == '}' {
				depth--
				if depth <= 1 {
					currentType = ""
				}
			}
		}

		// Check if this line starts a message type block
		parts := strings.Fields(trimmed)
		if len(parts) >= 1 && registry.IsMessageType(parts[0]) {
			if strings.Contains(trimmed, "{") {
				currentType = parts[0]
			}
		}
	}

	return currentType
}

func (s *Server) handleHover(msg *jsonrpcMessage) {
	var params textDocumentPosition
	json.Unmarshal(msg.Params, &params)

	content := s.docs[params.TextDocument.URI]
	word := getWordAt(content, params.Position.Line, params.Position.Character)

	info := ""
	switch word {
	case "protocol":
		info = "Defines a binary protocol with fields and byte order."
	case "message":
		info = "Defines a message-based protocol (JSON-RPC, REST, etc.)."
	case "field":
		info = "Declares a protocol field with name, type, and optional modifiers."
	case "bitfield":
		info = "Groups multiple sub-byte fields that share byte boundaries."
	case "checksum":
		info = "Specifies a checksum algorithm and covered fields."
	case "length_ref":
		info = "References another field for variable-length data."
	case "when":
		info = "Conditional field — only present when condition is true."
	case "display":
		info = "Display format hint (e.g., ipv4, mac, hex)."
	case "enum":
		info = "Maps integer values to human-readable names."
	case "const":
		info = "Defines a named constant value."
	case "range":
		info = "Constrains field value to [min..max]."
	case "default":
		info = "Specifies a default value for a transport field. Use `default auto` for engine-generated values."
	case "auto":
		info = "Used with `default auto` to indicate an engine-generated value (e.g., auto-incrementing ID)."
	}

	// If no static info found, check if the word is a transport-defined keyword
	if info == "" {
		if transportName := extractTransportName(content); transportName != "" {
			loader := pdl.NewTransportLoader(psl.FS, nil)
			td, err := loader.LoadTransport(transportName, nil)
			if err == nil {
				registry := pdl.NewDynamicKeywordRegistry(td)
				if mtd, ok := registry.GetMessageType(word); ok {
					fieldNames := make([]string, 0, len(mtd.Fields))
					for _, f := range mtd.Fields {
						fieldNames = append(fieldNames, f.Name)
					}
					info = "Transport-defined message type `" + word + "`. Fields: " + strings.Join(fieldNames, ", ")
					if mtd.ResponseDef != nil {
						info += ". Supports nested response blocks."
					}
				} else {
					// Check if it's a transport field name
					for _, mt := range td.MessageTypes {
						for _, f := range mt.Fields {
							if f.Name == word {
								info = "Transport field `" + word + "` (type: " + f.Type.String() + ") in message type `" + mt.Name + "`"
								if f.Optional {
									info += " (optional)"
								}
								if f.AutoValue {
									info += " (default auto)"
								} else if f.DefaultValue != nil {
									info += fmt.Sprintf(" (default: %v)", f.DefaultValue)
								}
								break
							}
						}
						if info != "" {
							break
						}
					}
				}
			}
		}
	}

	if info != "" {
		s.sendResponse(msg.ID, map[string]any{
			"contents": map[string]string{
				"kind":  "markdown",
				"value": "**" + word + "** — " + info,
			},
		})
	} else {
		s.sendResponse(msg.ID, nil)
	}
}

func (s *Server) handleFormatting(msg *jsonrpcMessage) {
	var params struct {
		TextDocument struct {
			URI string `json:"uri"`
		} `json:"textDocument"`
	}
	json.Unmarshal(msg.Params, &params)

	content, ok := s.docs[params.TextDocument.URI]
	if !ok {
		s.sendResponse(msg.ID, nil)
		return
	}

	// Try to parse and re-print for formatting
	parser := pdl.NewPDLParser(nil, nil)
	loader := pdl.NewTransportLoader(psl.FS, nil)
	parser.SetTransportLoader(loader)
	printer := &pdl.PDLPrinter{}

	var formatted string
	if strings.Contains(content, "message ") && strings.Contains(content, "transport ") {
		ms, err := parser.ParseMessage(content)
		if err != nil {
			s.sendResponse(msg.ID, nil)
			return
		}
		formatted = printer.PrintMessage(ms)
	} else {
		ps, err := parser.Parse(content)
		if err != nil {
			s.sendResponse(msg.ID, nil)
			return
		}
		formatted = printer.Print(ps)
	}

	lines := strings.Count(content, "\n")
	edits := []map[string]any{{
		"range": map[string]any{
			"start": map[string]int{"line": 0, "character": 0},
			"end":   map[string]int{"line": lines + 1, "character": 0},
		},
		"newText": formatted,
	}}
	s.sendResponse(msg.ID, edits)
}

func (s *Server) handleDefinition(msg *jsonrpcMessage) {
	var params textDocumentPosition
	json.Unmarshal(msg.Params, &params)

	content := s.docs[params.TextDocument.URI]
	word := getWordAt(content, params.Position.Line, params.Position.Character)

	// Search for field definition
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "field "+word+":") || strings.HasPrefix(trimmed, "field "+word+" :") {
			s.sendResponse(msg.ID, map[string]any{
				"uri": params.TextDocument.URI,
				"range": map[string]any{
					"start": map[string]int{"line": i, "character": strings.Index(line, word)},
					"end":   map[string]int{"line": i, "character": strings.Index(line, word) + len(word)},
				},
			})
			return
		}
	}
	s.sendResponse(msg.ID, nil)
}

func getLine(content string, line int) string {
	lines := strings.Split(content, "\n")
	if line < len(lines) {
		return lines[line]
	}
	return ""
}

func getWordAt(content string, line, col int) string {
	l := getLine(content, line)
	if col >= len(l) {
		return ""
	}
	// Find word boundaries
	start := col
	for start > 0 && isWordChar(l[start-1]) {
		start--
	}
	end := col
	for end < len(l) && isWordChar(l[end]) {
		end++
	}
	return l[start:end]
}

func isWordChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}
