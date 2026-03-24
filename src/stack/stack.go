// Package stack provides multi-layer protocol stack decoding engine.
package stack

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/protocol"
)

// Rule defines a protocol dispatch rule.
type Rule struct {
	Field string `json:"field"`
	Value any    `json:"value"`
	Next  string `json:"next_protocol"`
}

// Layer represents a decoded protocol layer.
type Layer struct {
	Protocol string         `json:"protocol"`
	Fields   map[string]any `json:"fields"`
	Bytes    int            `json:"bytes_read"`
}

// Engine is the multi-layer protocol stack decoder.
type Engine struct {
	lib   *protocol.Library
	rules []Rule
}

// NewEngine creates a new stack engine with default rules.
func NewEngine(lib *protocol.Library) *Engine {
	e := &Engine{lib: lib}
	e.registerDefaults()
	return e
}

// AddRule adds a dispatch rule.
func (e *Engine) AddRule(r Rule) {
	e.rules = append(e.rules, r)
}

// Decode decodes multiple protocol layers from raw data.
func (e *Engine) Decode(startProtocol string, data []byte) ([]Layer, error) {
	var layers []Layer
	current := startProtocol
	offset := 0
	maxDepth := 10

	for i := 0; i < maxDepth && current != "" && offset < len(data); i++ {
		result, err := e.lib.Decode(current, data[offset:])
		if err != nil {
			break
		}

		layer := Layer{
			Protocol: current,
			Fields:   result.Packet,
			Bytes:    result.BytesRead,
		}
		layers = append(layers, layer)
		offset += result.BytesRead

		// Find next protocol
		current = e.findNext(result.Packet)
	}

	if len(layers) == 0 {
		return nil, fmt.Errorf("failed to decode %s", startProtocol)
	}
	return layers, nil
}

func (e *Engine) findNext(fields map[string]any) string {
	for _, rule := range e.rules {
		if v, ok := fields[rule.Field]; ok {
			if fmt.Sprintf("%v", v) == fmt.Sprintf("%v", rule.Value) {
				return rule.Next
			}
		}
	}
	return ""
}

func (e *Engine) registerDefaults() {
	e.rules = []Rule{
		{Field: "ether_type", Value: "2048", Next: "IPv4"},
		{Field: "ether_type", Value: "2054", Next: "ARP"},
		{Field: "ether_type", Value: "34525", Next: "IPv6"},
		{Field: "protocol", Value: "6", Next: "TCP"},
		{Field: "protocol", Value: "17", Next: "UDP"},
		{Field: "protocol", Value: "1", Next: "ICMP"},
	}
}

// Encapsulate wraps data in multiple protocol layers.
func Encapsulate(lib *protocol.Library, layers []struct {
	Protocol string
	Fields   map[string]any
}) ([]byte, error) {
	var result []byte
	for i := len(layers) - 1; i >= 0; i-- {
		l := layers[i]
		encoded, err := lib.Encode(l.Protocol, l.Fields)
		if err != nil {
			return nil, fmt.Errorf("encode %s: %w", l.Protocol, err)
		}
		result = append(encoded, result...)
	}
	return result, nil
}
