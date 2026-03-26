// Package grpcgen generates gRPC .proto files from message protocols.
package grpcgen

import (
	"fmt"
	"strings"

	"github.com/hsqbyte/protospec/src/core/schema"
)

// GenerateProto generates a gRPC .proto file from a message schema.
func GenerateProto(ms *schema.MessageSchema) string {
	var b strings.Builder
	b.WriteString("syntax = \"proto3\";\n\n")
	b.WriteString(fmt.Sprintf("package %s;\n\n", strings.ToLower(ms.Name)))

	for _, msg := range ms.Messages {
		b.WriteString(fmt.Sprintf("message %s {\n", msg.Name))
		for i, f := range msg.Fields {
			protoType := msgTypeToProto(f.Type)
			b.WriteString(fmt.Sprintf("  %s %s = %d;\n", protoType, f.Name, i+1))
		}
		b.WriteString("}\n\n")
	}

	// Generate service
	b.WriteString(fmt.Sprintf("service %sService {\n", ms.Name))
	for _, msg := range ms.Messages {
		if msg.Kind == "request" {
			respName := ms.Name + "Response"
			for _, m := range ms.Messages {
				if m.Kind == "response" {
					respName = m.Name
					break
				}
			}
			b.WriteString(fmt.Sprintf("  rpc %s (%s) returns (%s);\n", msg.Name, msg.Name, respName))
		}
	}
	b.WriteString("}\n")
	return b.String()
}

func msgTypeToProto(t schema.MessageFieldType) string {
	switch t {
	case schema.MsgString:
		return "string"
	case schema.MsgNumber:
		return "double"
	case schema.MsgBoolean:
		return "bool"
	default:
		return "string"
	}
}
