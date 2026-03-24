// Package infer provides protocol reverse engineering — automatic structure inference from binary data.
package infer

import (
	"fmt"
	"math"
	"strings"
)

// InferredField represents a field discovered by inference.
type InferredField struct {
	Offset     int            `json:"offset"`
	Length     int            `json:"length"`
	Name       string         `json:"name"`
	Type       string         `json:"type"` // "uint", "bytes", "string", "length_field", "enum"
	Confidence float64        `json:"confidence"`
	Values     []uint64       `json:"sample_values,omitempty"`
	EnumVals   map[uint64]int `json:"enum_distribution,omitempty"`
}

// InferResult holds the inference result.
type InferResult struct {
	Fields      []InferredField `json:"fields"`
	MinLength   int             `json:"min_length"`
	MaxLength   int             `json:"max_length"`
	FixedFields []int           `json:"fixed_offsets"`
}

// Infer analyzes multiple packets and infers protocol structure.
func Infer(packets [][]byte) (*InferResult, error) {
	if len(packets) == 0 {
		return nil, fmt.Errorf("no packets to analyze")
	}

	result := &InferResult{
		MinLength: len(packets[0]),
		MaxLength: len(packets[0]),
	}

	for _, pkt := range packets {
		if len(pkt) < result.MinLength {
			result.MinLength = len(pkt)
		}
		if len(pkt) > result.MaxLength {
			result.MaxLength = len(pkt)
		}
	}

	// Analyze byte-by-byte variance
	if result.MinLength == 0 {
		return result, nil
	}

	variances := analyzeVariance(packets, result.MinLength)
	boundaries := detectBoundaries(variances, result.MinLength)
	fields := buildFields(boundaries, packets, result.MinLength)

	// Detect length fields
	detectLengthFields(fields, packets)

	// Detect enum fields
	detectEnumFields(fields, packets)

	result.Fields = fields
	return result, nil
}

func analyzeVariance(packets [][]byte, minLen int) []float64 {
	variances := make([]float64, minLen)
	for offset := 0; offset < minLen; offset++ {
		var sum, sumSq float64
		n := float64(len(packets))
		for _, pkt := range packets {
			v := float64(pkt[offset])
			sum += v
			sumSq += v * v
		}
		mean := sum / n
		variances[offset] = sumSq/n - mean*mean
	}
	return variances
}

func detectBoundaries(variances []float64, minLen int) []int {
	boundaries := []int{0}
	for i := 1; i < minLen; i++ {
		// Boundary when variance changes significantly
		diff := math.Abs(variances[i] - variances[i-1])
		if diff > 50 || (variances[i-1] == 0 && variances[i] > 0) || (variances[i-1] > 0 && variances[i] == 0) {
			boundaries = append(boundaries, i)
		}
	}
	// Align to common sizes
	aligned := alignBoundaries(boundaries, minLen)
	return aligned
}

func alignBoundaries(boundaries []int, maxLen int) []int {
	if len(boundaries) == 0 {
		return []int{0}
	}
	// Merge boundaries that are too close
	var merged []int
	merged = append(merged, boundaries[0])
	for i := 1; i < len(boundaries); i++ {
		if boundaries[i]-merged[len(merged)-1] >= 1 {
			merged = append(merged, boundaries[i])
		}
	}
	return merged
}

func buildFields(boundaries []int, packets [][]byte, minLen int) []InferredField {
	var fields []InferredField
	for i := 0; i < len(boundaries); i++ {
		start := boundaries[i]
		end := minLen
		if i+1 < len(boundaries) {
			end = boundaries[i+1]
		}
		length := end - start

		fieldType := "uint"
		if length > 4 {
			fieldType = "bytes"
			// Check if it looks like a string
			isString := true
			for _, pkt := range packets {
				for j := start; j < end && j < len(pkt); j++ {
					if pkt[j] < 0x20 || pkt[j] > 0x7E {
						if pkt[j] != 0 {
							isString = false
							break
						}
					}
				}
				if !isString {
					break
				}
			}
			if isString {
				fieldType = "string"
			}
		}

		// Collect sample values
		var values []uint64
		for _, pkt := range packets {
			if start+length <= len(pkt) {
				var v uint64
				for j := 0; j < length && j < 8; j++ {
					v = v<<8 | uint64(pkt[start+j])
				}
				values = append(values, v)
			}
		}

		// Check if constant
		confidence := 0.5
		if len(values) > 1 {
			allSame := true
			for _, v := range values[1:] {
				if v != values[0] {
					allSame = false
					break
				}
			}
			if allSame {
				confidence = 0.9
			}
		}

		name := fmt.Sprintf("field_%d", i+1)
		bitWidth := length * 8
		typeName := fieldType
		if fieldType == "uint" {
			typeName = fmt.Sprintf("uint%d", bitWidth)
		}

		fields = append(fields, InferredField{
			Offset:     start,
			Length:     length,
			Name:       name,
			Type:       typeName,
			Confidence: confidence,
			Values:     values,
		})
	}
	return fields
}

func detectLengthFields(fields []InferredField, packets [][]byte) {
	for i := range fields {
		if fields[i].Length > 4 || fields[i].Length == 0 {
			continue
		}
		// Check if this field's value correlates with packet length
		matches := 0
		for _, pkt := range packets {
			if fields[i].Offset+fields[i].Length > len(pkt) {
				continue
			}
			var v int
			for j := 0; j < fields[i].Length; j++ {
				v = v<<8 | int(pkt[fields[i].Offset+j])
			}
			// Check various length relationships
			pktLen := len(pkt)
			if v == pktLen || v == pktLen-fields[i].Offset-fields[i].Length || v == pktLen-4 {
				matches++
			}
		}
		if matches > len(packets)/2 {
			fields[i].Type = "length_field"
			fields[i].Confidence = float64(matches) / float64(len(packets))
		}
	}
}

func detectEnumFields(fields []InferredField, packets [][]byte) {
	for i := range fields {
		if fields[i].Length > 2 {
			continue
		}
		dist := make(map[uint64]int)
		for _, v := range fields[i].Values {
			dist[v]++
		}
		// If few distinct values relative to samples, likely an enum
		if len(dist) > 0 && len(dist) <= 16 && len(dist) < len(packets)/2 {
			fields[i].Type = "enum"
			fields[i].EnumVals = dist
			fields[i].Confidence = 0.7
		}
	}
}

// ToPSL generates a candidate PSL file from inference results.
func (r *InferResult) ToPSL(name string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("protocol %s version \"1.0\" {\n", name))
	b.WriteString("    byte_order big-endian;\n\n")

	for _, f := range r.Fields {
		switch f.Type {
		case "length_field":
			b.WriteString(fmt.Sprintf("    field %s: uint%d; // likely length field (confidence: %.0f%%)\n",
				f.Name, f.Length*8, f.Confidence*100))
		case "enum":
			b.WriteString(fmt.Sprintf("    field %s: uint%d enum {\n", f.Name, f.Length*8))
			for val, count := range f.EnumVals {
				b.WriteString(fmt.Sprintf("        %d = \"value_%d\" // seen %d times\n", val, val, count))
			}
			b.WriteString("    };\n")
		case "bytes":
			b.WriteString(fmt.Sprintf("    field %s: bytes[%d];\n", f.Name, f.Length))
		case "string":
			b.WriteString(fmt.Sprintf("    field %s: string; // likely string\n", f.Name))
		default:
			b.WriteString(fmt.Sprintf("    field %s: %s;\n", f.Name, f.Type))
		}
	}

	// Variable tail
	if r.MaxLength > r.MinLength {
		b.WriteString("    field payload: bytes; // variable length tail\n")
	}

	b.WriteString("}\n")
	return b.String()
}
