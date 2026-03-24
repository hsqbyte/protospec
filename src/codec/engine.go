package codec

import (
	"github.com/hsqbyte/protospec/src/checksum"
	"github.com/hsqbyte/protospec/src/format"
	"github.com/hsqbyte/protospec/src/schema"
)

// CodecEngine provides a unified encode/decode interface backed by
// Encoder and Decoder with shared registries.
type CodecEngine struct {
	encoder *Encoder
	decoder *Decoder
}

// NewCodecEngine creates a new CodecEngine with the given registries.
func NewCodecEngine(cr *checksum.ChecksumRegistry, fr *format.FormatRegistry) *CodecEngine {
	return &CodecEngine{
		encoder: NewEncoder(cr, fr),
		decoder: NewDecoder(cr, fr),
	}
}

// Encode serialises packet into []byte according to the schema.
func (e *CodecEngine) Encode(s *schema.ProtocolSchema, packet map[string]any) ([]byte, error) {
	return e.encoder.Encode(s, packet)
}

// Decode deserialises data into a DecodeResult according to the schema.
func (e *CodecEngine) Decode(s *schema.ProtocolSchema, data []byte) (*DecodeResult, error) {
	return e.decoder.Decode(s, data)
}
