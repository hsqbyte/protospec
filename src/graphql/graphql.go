// Package graphql provides GraphQL schema generation from message protocols.
package graphql

import (
	"fmt"
	"strings"

	"github.com/hsqbyte/protospec/src/schema"
)

// GenerateSchema generates a GraphQL schema from a message schema.
func GenerateSchema(ms *schema.MessageSchema) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("# Generated from %s v%s\n\n", ms.Name, ms.Version))

	// Generate types for each message
	for _, msg := range ms.Messages {
		b.WriteString(fmt.Sprintf("type %s {\n", msg.Name))
		for _, f := range msg.Fields {
			gqlType := msgTypeToGraphQL(f.Type, f.Optional)
			b.WriteString(fmt.Sprintf("  %s: %s\n", f.Name, gqlType))
		}
		b.WriteString("}\n\n")
	}

	// Generate Query type
	b.WriteString("type Query {\n")
	for _, msg := range ms.Messages {
		if msg.Kind == "response" || msg.Kind == "notification" {
			b.WriteString(fmt.Sprintf("  %s: %s\n", lcFirst(msg.Name), msg.Name))
		}
	}
	b.WriteString("}\n\n")

	// Generate Mutation type for requests
	b.WriteString("type Mutation {\n")
	for _, msg := range ms.Messages {
		if msg.Kind == "request" {
			inputName := msg.Name + "Input"
			b.WriteString(fmt.Sprintf("  %s(input: %s!): Boolean\n", lcFirst(msg.Name), inputName))
		}
	}
	b.WriteString("}\n")

	return b.String()
}

func msgTypeToGraphQL(t schema.MessageFieldType, optional bool) string {
	base := ""
	switch t {
	case schema.MsgString:
		base = "String"
	case schema.MsgNumber:
		base = "Float"
	case schema.MsgBoolean:
		base = "Boolean"
	case schema.MsgObject:
		base = "JSON"
	case schema.MsgArray:
		base = "[JSON]"
	default:
		base = "String"
	}
	if !optional {
		base += "!"
	}
	return base
}

func lcFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}
