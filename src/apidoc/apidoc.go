// Package apidoc generates API documentation (OpenAPI, AsyncAPI) from message protocols.
package apidoc

import (
	"encoding/json"
	"fmt"

	"github.com/hsqbyte/protospec/src/schema"
)

// Generator generates API documentation from message schemas.
type Generator struct{}

// NewGenerator creates a new API doc generator.
func NewGenerator() *Generator {
	return &Generator{}
}

// GenerateOpenAPI generates an OpenAPI 3.0 spec from a message schema.
func (g *Generator) GenerateOpenAPI(ms *schema.MessageSchema) (string, error) {
	spec := map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":   ms.Name + " API",
			"version": ms.Version,
		},
		"paths": g.buildPaths(ms),
		"components": map[string]any{
			"schemas": g.buildSchemas(ms),
		},
	}
	data, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GenerateAsyncAPI generates an AsyncAPI 2.6 spec from a message schema.
func (g *Generator) GenerateAsyncAPI(ms *schema.MessageSchema) (string, error) {
	spec := map[string]any{
		"asyncapi": "2.6.0",
		"info": map[string]any{
			"title":   ms.Name,
			"version": ms.Version,
		},
		"channels": g.buildChannels(ms),
		"components": map[string]any{
			"schemas": g.buildSchemas(ms),
		},
	}
	data, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (g *Generator) buildPaths(ms *schema.MessageSchema) map[string]any {
	paths := make(map[string]any)
	for _, msg := range ms.Messages {
		if msg.Kind == "request" {
			path := fmt.Sprintf("/%s", msg.Name)
			paths[path] = map[string]any{
				"post": map[string]any{
					"summary":     msg.Name,
					"operationId": msg.Name,
					"requestBody": map[string]any{
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{"$ref": "#/components/schemas/" + msg.Name},
							},
						},
					},
					"responses": map[string]any{
						"200": map[string]any{"description": "Success"},
					},
				},
			}
		}
	}
	return paths
}

func (g *Generator) buildChannels(ms *schema.MessageSchema) map[string]any {
	channels := make(map[string]any)
	for _, msg := range ms.Messages {
		ch := fmt.Sprintf("%s/%s", ms.Name, msg.Name)
		op := "subscribe"
		if msg.Kind == "request" {
			op = "publish"
		}
		channels[ch] = map[string]any{
			op: map[string]any{
				"message": map[string]any{
					"$ref": "#/components/schemas/" + msg.Name,
				},
			},
		}
	}
	return channels
}

func (g *Generator) buildSchemas(ms *schema.MessageSchema) map[string]any {
	schemas := make(map[string]any)
	for _, msg := range ms.Messages {
		props := make(map[string]any)
		var required []string
		for _, f := range msg.Fields {
			props[f.Name] = map[string]any{"type": msgTypeToJSON(f.Type)}
			if !f.Optional {
				required = append(required, f.Name)
			}
		}
		s := map[string]any{
			"type":       "object",
			"properties": props,
		}
		if len(required) > 0 {
			s["required"] = required
		}
		schemas[msg.Name] = s
	}
	return schemas
}

func msgTypeToJSON(t schema.MessageFieldType) string {
	switch t {
	case schema.MsgString:
		return "string"
	case schema.MsgNumber:
		return "number"
	case schema.MsgBoolean:
		return "boolean"
	case schema.MsgObject:
		return "object"
	case schema.MsgArray:
		return "array"
	default:
		return "string"
	}
}
