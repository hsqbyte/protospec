// Package template provides protocol template system for rapid protocol creation.
package template

import (
	"fmt"
	"strings"
)

// Template represents a protocol template.
type Template struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Category    string     `json:"category"` // "tlv", "rpc", "stream", "request-response"
	Variables   []Variable `json:"variables"`
	Body        string     `json:"body"`
}

// Variable represents a template variable.
type Variable struct {
	Name    string `json:"name"`
	Default string `json:"default"`
	Desc    string `json:"description"`
}

// BuiltinTemplates returns all built-in protocol templates.
func BuiltinTemplates() []Template {
	return []Template{
		{
			Name: "tlv", Description: "Type-Length-Value protocol", Category: "tlv",
			Variables: []Variable{{Name: "name", Default: "MyTLV"}, {Name: "version", Default: "1.0"}},
			Body: `protocol {{name}} version "{{version}}" {
    byte_order big-endian;
    field type: uint8;
    field length: uint16;
    field value: bytes;
}`,
		},
		{
			Name: "request-response", Description: "Request-Response message protocol", Category: "rpc",
			Variables: []Variable{{Name: "name", Default: "MyAPI"}, {Name: "version", Default: "1.0"}},
			Body: `message {{name}} version "{{version}}" {
    transport rest;
    request Ping {
        field id: string;
        field timestamp: number;
    }
    response Pong {
        field id: string;
        field timestamp: number;
        field server_time: number;
    }
}`,
		},
		{
			Name: "header-payload", Description: "Fixed header with variable payload", Category: "stream",
			Variables: []Variable{{Name: "name", Default: "MyProto"}, {Name: "version", Default: "1.0"}},
			Body: `protocol {{name}} version "{{version}}" {
    byte_order big-endian;
    field magic: uint16;
    field version: uint8;
    field flags: uint8;
    field length: uint32;
    field payload: bytes;
}`,
		},
	}
}

// Render renders a template with the given variables.
func Render(tmpl *Template, vars map[string]string) string {
	result := tmpl.Body
	for _, v := range tmpl.Variables {
		val := v.Default
		if userVal, ok := vars[v.Name]; ok {
			val = userVal
		}
		result = strings.ReplaceAll(result, "{{"+v.Name+"}}", val)
	}
	return result
}

// FormatList formats template list for display.
func FormatList(templates []Template) string {
	var b strings.Builder
	b.WriteString("Available templates:\n")
	for _, t := range templates {
		b.WriteString(fmt.Sprintf("  %-20s %s [%s]\n", t.Name, t.Description, t.Category))
	}
	return b.String()
}
