// Package debug provides interactive protocol debugging with breakpoints.
package debug

import (
	"fmt"
	"strings"

	"github.com/hsqbyte/protospec/src/schema"
)

// Breakpoint defines a debug breakpoint.
type Breakpoint struct {
	Field     string `json:"field"`
	Condition string `json:"condition,omitempty"` // e.g. "== 4", "> 100"
}

// FieldState represents the state of a field during decoding.
type FieldState struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	BitWidth int    `json:"bit_width"`
	Offset   int    `json:"offset_bits"`
	Value    any    `json:"value,omitempty"`
	HexView  string `json:"hex_view"`
}

// Session is an interactive debug session.
type Session struct {
	schema      *schema.ProtocolSchema
	data        []byte
	breakpoints []Breakpoint
	states      []FieldState
	currentIdx  int
}

// NewSession creates a new debug session.
func NewSession(s *schema.ProtocolSchema, data []byte) *Session {
	return &Session{schema: s, data: data}
}

// AddBreakpoint adds a breakpoint.
func (s *Session) AddBreakpoint(bp Breakpoint) {
	s.breakpoints = append(s.breakpoints, bp)
}

// StepAll decodes all fields and returns their states.
func (s *Session) StepAll() []FieldState {
	var states []FieldState
	bitOffset := 0

	for _, f := range s.schema.Fields {
		if f.IsBitfieldGroup {
			for _, bf := range f.BitfieldFields {
				state := FieldState{
					Name:     bf.Name,
					Type:     bf.Type.String(),
					BitWidth: bf.BitWidth,
					Offset:   bitOffset,
					HexView:  hexSlice(s.data, bitOffset/8, (bitOffset+bf.BitWidth+7)/8),
				}
				states = append(states, state)
				bitOffset += bf.BitWidth
			}
		} else {
			byteLen := f.BitWidth / 8
			if f.BitWidth == 0 {
				byteLen = len(s.data) - bitOffset/8
			}
			state := FieldState{
				Name:     f.Name,
				Type:     f.Type.String(),
				BitWidth: f.BitWidth,
				Offset:   bitOffset,
				HexView:  hexSlice(s.data, bitOffset/8, bitOffset/8+byteLen),
			}
			states = append(states, state)
			bitOffset += f.BitWidth
		}
	}
	s.states = states
	return states
}

// FormatHexView formats data as annotated hex view.
func FormatHexView(data []byte, states []FieldState) string {
	var b strings.Builder
	b.WriteString("Offset  Hex                                      ASCII\n")
	b.WriteString(strings.Repeat("─", 70) + "\n")

	for i := 0; i < len(data); i += 16 {
		b.WriteString(fmt.Sprintf("%04x    ", i))
		end := i + 16
		if end > len(data) {
			end = len(data)
		}
		for j := i; j < end; j++ {
			b.WriteString(fmt.Sprintf("%02x ", data[j]))
		}
		for j := end; j < i+16; j++ {
			b.WriteString("   ")
		}
		b.WriteString("  ")
		for j := i; j < end; j++ {
			if data[j] >= 32 && data[j] < 127 {
				b.WriteByte(data[j])
			} else {
				b.WriteByte('.')
			}
		}
		b.WriteString("\n")
	}

	if len(states) > 0 {
		b.WriteString("\nFields:\n")
		for _, s := range states {
			b.WriteString(fmt.Sprintf("  [bit %d] %s: %s (%d bits) %s\n",
				s.Offset, s.Name, s.Type, s.BitWidth, s.HexView))
		}
	}
	return b.String()
}

func hexSlice(data []byte, start, end int) string {
	if start >= len(data) {
		return ""
	}
	if end > len(data) {
		end = len(data)
	}
	var parts []string
	for _, b := range data[start:end] {
		parts = append(parts, fmt.Sprintf("%02x", b))
	}
	return strings.Join(parts, " ")
}
