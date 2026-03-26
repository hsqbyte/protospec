// Package datagen generates test data for protocol schemas.
package datagen

import (
	"math"
	"math/rand"

	"github.com/hsqbyte/protospec/src/core/schema"
)

// Generator generates protocol test data.
type Generator struct {
	rng *rand.Rand
}

// NewGenerator creates a new data generator.
func NewGenerator(seed int64) *Generator {
	return &Generator{rng: rand.New(rand.NewSource(seed))}
}

// GenerateValid generates valid test data for a protocol.
func (g *Generator) GenerateValid(s *schema.ProtocolSchema, count int) []map[string]any {
	var results []map[string]any
	for i := 0; i < count; i++ {
		packet := make(map[string]any)
		for _, f := range s.Fields {
			if f.IsBitfieldGroup {
				for _, bf := range f.BitfieldFields {
					packet[bf.Name] = g.randomValue(bf)
				}
			} else {
				packet[f.Name] = g.randomValue(f)
			}
		}
		results = append(results, packet)
	}
	return results
}

// GenerateBoundary generates boundary value test data.
func (g *Generator) GenerateBoundary(s *schema.ProtocolSchema) []map[string]any {
	var results []map[string]any
	var fields []schema.FieldDef
	for _, f := range s.Fields {
		if f.IsBitfieldGroup {
			fields = append(fields, f.BitfieldFields...)
		} else {
			fields = append(fields, f)
		}
	}

	// Min values
	minPacket := make(map[string]any)
	for _, f := range fields {
		minPacket[f.Name] = g.minValue(f)
	}
	results = append(results, minPacket)

	// Max values
	maxPacket := make(map[string]any)
	for _, f := range fields {
		maxPacket[f.Name] = g.maxValue(f)
	}
	results = append(results, maxPacket)

	return results
}

// GenerateFuzz generates malformed/fuzz test data.
func (g *Generator) GenerateFuzz(s *schema.ProtocolSchema, count int) []map[string]any {
	var results []map[string]any
	for i := 0; i < count; i++ {
		packet := make(map[string]any)
		for _, f := range s.Fields {
			if f.IsBitfieldGroup {
				for _, bf := range f.BitfieldFields {
					packet[bf.Name] = g.fuzzValue(bf)
				}
			} else {
				packet[f.Name] = g.fuzzValue(f)
			}
		}
		results = append(results, packet)
	}
	return results
}

func (g *Generator) randomValue(f schema.FieldDef) any {
	switch f.Type {
	case schema.Uint, schema.Int:
		if f.BitWidth > 0 {
			max := int64(math.Pow(2, float64(f.BitWidth))) - 1
			if len(f.EnumMap) > 0 {
				keys := make([]int, 0, len(f.EnumMap))
				for k := range f.EnumMap {
					keys = append(keys, k)
				}
				return keys[g.rng.Intn(len(keys))]
			}
			return g.rng.Int63n(max + 1)
		}
		return 0
	case schema.Bytes:
		n := 4 + g.rng.Intn(16)
		b := make([]byte, n)
		g.rng.Read(b)
		return b
	case schema.String:
		return "test"
	case schema.Bool:
		return g.rng.Intn(2) == 1
	}
	return nil
}

func (g *Generator) minValue(f schema.FieldDef) any {
	switch f.Type {
	case schema.Uint:
		return int64(0)
	case schema.Int:
		return int64(0)
	case schema.Bytes:
		return []byte{}
	case schema.String:
		return ""
	case schema.Bool:
		return false
	}
	return nil
}

func (g *Generator) maxValue(f schema.FieldDef) any {
	switch f.Type {
	case schema.Uint:
		if f.BitWidth > 0 {
			return int64(math.Pow(2, float64(f.BitWidth))) - 1
		}
		return int64(0)
	case schema.Int:
		if f.BitWidth > 0 {
			return int64(math.Pow(2, float64(f.BitWidth-1))) - 1
		}
		return int64(0)
	case schema.Bytes:
		return make([]byte, 1500)
	case schema.String:
		return "max_test_string"
	case schema.Bool:
		return true
	}
	return nil
}

func (g *Generator) fuzzValue(f schema.FieldDef) any {
	switch f.Type {
	case schema.Uint, schema.Int:
		// Overflow values
		if f.BitWidth > 0 {
			return int64(math.Pow(2, float64(f.BitWidth)))
		}
		return int64(-1)
	case schema.Bytes:
		return []byte{0xFF, 0xFF, 0xFF, 0xFF}
	case schema.String:
		return "\x00\xff\xfe"
	case schema.Bool:
		return true
	}
	return nil
}
