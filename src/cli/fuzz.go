package cli

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/hsqbyte/protospec/src/schema"
)

func runFuzz(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl fuzz <protocol> [-n count] [--malformed]")
	}

	name := args[0]
	count := 10
	malformed := false
	format := "hex"

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "-n":
			i++
			if i < len(args) {
				fmt.Sscanf(args[i], "%d", &count)
			}
		case "--malformed":
			malformed = true
		case "--format":
			i++
			if i < len(args) {
				format = args[i]
			}
		}
	}

	s, err := ctx.Lib.Registry().GetSchema(name)
	if err != nil {
		return err
	}

	for i := 0; i < count; i++ {
		data := generateFuzzData(s, malformed)
		switch format {
		case "hex":
			fmt.Printf("#%d %s\n", i+1, hex.EncodeToString(data))
		case "json":
			// Try to decode and show as JSON
			result, err := ctx.Lib.Decode(name, data)
			if err != nil {
				fmt.Printf("#%d [decode error: %v] %s\n", i+1, err, hex.EncodeToString(data))
			} else {
				j, _ := json.Marshal(result.Packet)
				fmt.Printf("#%d %s\n", i+1, string(j))
			}
		default:
			fmt.Printf("#%d %s\n", i+1, hex.EncodeToString(data))
		}
	}
	return nil
}

func generateFuzzData(s *schema.ProtocolSchema, malformed bool) []byte {
	var fields []schema.FieldDef
	for _, f := range s.Fields {
		if f.IsBitfieldGroup {
			fields = append(fields, f.BitfieldFields...)
		} else {
			fields = append(fields, f)
		}
	}

	// Calculate total bits
	totalBits := 0
	hasVariable := false
	for _, f := range fields {
		if f.BitWidth > 0 {
			totalBits += f.BitWidth
		} else {
			hasVariable = true
		}
	}

	totalBytes := (totalBits + 7) / 8
	if hasVariable {
		totalBytes += rand.Intn(64) // Add random payload
	}

	data := make([]byte, totalBytes)

	if malformed {
		// Completely random bytes
		rand.Read(data)
		return data
	}

	// Generate semi-valid data
	bitOffset := 0
	for _, f := range fields {
		if f.BitWidth == 0 {
			continue
		}

		var val uint64
		if f.DefaultValue != nil {
			switch v := f.DefaultValue.(type) {
			case int64:
				val = uint64(v)
			case int:
				val = uint64(v)
			}
		} else if f.RangeMin != nil && f.RangeMax != nil {
			min := *f.RangeMin
			max := *f.RangeMax
			val = uint64(min + rand.Int63n(max-min+1))
		} else if f.EnumMap != nil {
			// Pick a random valid enum value
			keys := make([]int, 0, len(f.EnumMap))
			for k := range f.EnumMap {
				keys = append(keys, k)
			}
			if len(keys) > 0 {
				val = uint64(keys[rand.Intn(len(keys))])
			}
		} else {
			// Random value within bit range
			maxVal := uint64(1) << f.BitWidth
			if f.BitWidth >= 64 {
				val = rand.Uint64()
			} else {
				val = rand.Uint64() % maxVal
			}
		}

		// Write value at bit offset
		writeBitsToBytes(data, bitOffset, f.BitWidth, val)
		bitOffset += f.BitWidth
	}

	// Fill remaining with random
	byteStart := (bitOffset + 7) / 8
	if byteStart < len(data) {
		rand.Read(data[byteStart:])
	}

	return data
}

func writeBitsToBytes(data []byte, bitOffset, bitWidth int, val uint64) {
	for i := 0; i < bitWidth; i++ {
		bitIdx := bitOffset + i
		byteIdx := bitIdx / 8
		bitPos := 7 - (bitIdx % 8)
		if byteIdx >= len(data) {
			break
		}
		if val&(1<<uint(bitWidth-1-i)) != 0 {
			data[byteIdx] |= 1 << uint(bitPos)
		}
	}
}
