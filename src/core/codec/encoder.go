package codec

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/core/checksum"
	"github.com/hsqbyte/protospec/src/core/errors"
	"github.com/hsqbyte/protospec/src/core/format"
	"github.com/hsqbyte/protospec/src/core/schema"
)

// Encoder encodes a Packet (map[string]any) into raw bytes
// according to a ProtocolSchema.
type Encoder struct {
	checksumRegistry *checksum.ChecksumRegistry
	formatRegistry   *format.FormatRegistry
}

// NewEncoder creates a new Encoder with the given registries.
func NewEncoder(cr *checksum.ChecksumRegistry, fr *format.FormatRegistry) *Encoder {
	return &Encoder{
		checksumRegistry: cr,
		formatRegistry:   fr,
	}
}

// checksumDeferred records a checksum field that needs backfill after all
// fields have been encoded.
type checksumDeferred struct {
	field      schema.FieldDef
	byteOffset int
	byteOrder  schema.ByteOrder
}

// Encode serialises packet into []byte according to the schema.
func (enc *Encoder) Encode(s *schema.ProtocolSchema, packet map[string]any) ([]byte, error) {
	w := NewBitWriter()

	// Track byte ranges for each field (for checksum coverage).
	// fieldStart[name] = byte offset where the field starts.
	// fieldEnd[name]   = byte offset just past the field's last byte.
	type fieldRange struct {
		start int
		end   int
	}
	fieldRanges := make(map[string]fieldRange)

	var deferred []checksumDeferred

	// Flat list of all field names in order (for expanding ranges).
	allFieldNames := flatFieldNames(s.Fields)

	for _, f := range s.Fields {
		order := effectiveByteOrder(f, s.DefaultByteOrder)

		if f.IsBitfieldGroup {
			for _, sub := range f.BitfieldFields {
				subOrder := effectiveByteOrder(sub, s.DefaultByteOrder)
				startOff := w.Len()
				if err := enc.encodeField(w, sub, packet, subOrder); err != nil {
					return nil, err
				}
				endOff := w.Len()
				if w.BitOffset() != 0 {
					endOff++ // partial byte counts
				}
				fieldRanges[sub.Name] = fieldRange{start: startOff, end: endOff}
			}
			continue
		}

		// Conditional field
		if f.Condition != nil {
			ok, err := evaluateCondition(f.Condition, packet)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}
		}

		// Checksum field — write placeholder, defer computation
		if f.Checksum != nil {
			startOff := w.Len()
			deferred = append(deferred, checksumDeferred{
				field:      f,
				byteOffset: startOff,
				byteOrder:  order,
			})
			// Write zeros as placeholder
			if err := w.WriteBits(0, f.BitWidth, order); err != nil {
				return nil, err
			}
			endOff := w.Len()
			if w.BitOffset() != 0 {
				endOff++
			}
			fieldRanges[f.Name] = fieldRange{start: startOff, end: endOff}
			continue
		}

		startOff := w.Len()
		if err := enc.encodeField(w, f, packet, order); err != nil {
			return nil, err
		}
		endOff := w.Len()
		if w.BitOffset() != 0 {
			endOff++
		}
		fieldRanges[f.Name] = fieldRange{start: startOff, end: endOff}
	}

	// Backfill checksums
	result := w.Bytes()
	for _, d := range deferred {
		covered, err := resolveCoverFields(d.field.Checksum.CoverFields, allFieldNames)
		if err != nil {
			return nil, err
		}
		// Determine the byte range covered
		var minStart, maxEnd int
		first := true
		for _, name := range covered {
			r, ok := fieldRanges[name]
			if !ok {
				return nil, fmt.Errorf("checksum cover field %q not found in encoded fields", name)
			}
			if first || r.start < minStart {
				minStart = r.start
			}
			if first || r.end > maxEnd {
				maxEnd = r.end
			}
			first = false
		}
		if first {
			return nil, fmt.Errorf("checksum for field %q covers no fields", d.field.Name)
		}

		coveredBytes := result[minStart:maxEnd]

		fn, err := enc.checksumRegistry.Get(d.field.Checksum.Algorithm)
		if err != nil {
			return nil, err
		}
		csVal := fn(coveredBytes)

		// Backfill: convert checksum value to bytes
		numBytes := d.field.BitWidth / 8
		csBytes := make([]byte, numBytes)
		if d.byteOrder == schema.BigEndian {
			for i := numBytes - 1; i >= 0; i-- {
				csBytes[i] = byte(csVal & 0xFF)
				csVal >>= 8
			}
		} else {
			for i := 0; i < numBytes; i++ {
				csBytes[i] = byte(csVal & 0xFF)
				csVal >>= 8
			}
		}
		w.SetBytesAt(d.byteOffset, csBytes)
	}

	return w.Bytes(), nil
}

// encodeField encodes a single field value into the BitWriter.
func (enc *Encoder) encodeField(w *BitWriter, f schema.FieldDef, packet map[string]any, order schema.ByteOrder) error {
	switch {
	case f.Type == schema.Bool:
		v, err := getUint64(packet, f.Name)
		if err != nil {
			return err
		}
		if v > 1 {
			return &errors.InvalidFieldError{
				FieldName:   f.Name,
				ValidRange:  "0..1",
				ActualValue: v,
			}
		}
		return w.WriteBits(v, 1, order)

	case (f.Type == schema.Bytes || f.Type == schema.String) && f.FixedLength > 0:
		// Fixed-length field (e.g. bytes[16])
		data, err := getBytes(packet, f.Name)
		if err != nil {
			return err
		}
		toWrite := data
		if len(toWrite) > f.FixedLength {
			toWrite = toWrite[:f.FixedLength]
		}
		if err := w.WriteByteSlice(toWrite); err != nil {
			return err
		}
		if len(toWrite) < f.FixedLength {
			pad := make([]byte, f.FixedLength-len(toWrite))
			return w.WriteByteSlice(pad)
		}
		return nil

	case (f.Type == schema.Bytes || f.Type == schema.String) && f.LengthRef != nil:
		// Variable-length field
		refVal, err := getUint64(packet, f.LengthRef.FieldName)
		if err != nil {
			return err
		}
		length := int(refVal)*f.LengthRef.Scale + f.LengthRef.Offset
		if length < 0 {
			length = 0
		}
		data, err := getBytes(packet, f.Name)
		if err != nil {
			return err
		}
		// Write exactly 'length' bytes
		toWrite := data
		if len(toWrite) > length {
			toWrite = toWrite[:length]
		}
		if err := w.WriteByteSlice(toWrite); err != nil {
			return err
		}
		// Pad with zeros if data is shorter than length
		if len(toWrite) < length {
			pad := make([]byte, length-len(toWrite))
			return w.WriteByteSlice(pad)
		}
		return nil

	case (f.Type == schema.Bytes || f.Type == schema.String) && f.LengthRef == nil:
		// Payload field — write all bytes from packet value
		data, err := getBytes(packet, f.Name)
		if err != nil {
			return err
		}
		return w.WriteByteSlice(data)

	case f.Type == schema.Uint || f.Type == schema.Int:
		var val uint64
		if f.DisplayFormat != "" {
			// The packet value may be a human-readable string
			val2, err := enc.resolveDisplayValue(f, packet)
			if err != nil {
				return err
			}
			val = val2
		} else {
			v, err := getUint64(packet, f.Name)
			if err != nil {
				// Use default value if available
				if f.DefaultValue != nil {
					val = uint64(f.DefaultValue.(int64))
				} else {
					return err
				}
			} else {
				val = v
			}
		}

		// Validate user-defined range constraint
		if f.RangeMin != nil || f.RangeMax != nil {
			sval := int64(val)
			if f.RangeMin != nil && sval < *f.RangeMin {
				return &errors.InvalidFieldError{
					FieldName:   f.Name,
					ValidRange:  fmt.Sprintf("%d..%d", *f.RangeMin, *f.RangeMax),
					ActualValue: val,
				}
			}
			if f.RangeMax != nil && sval > *f.RangeMax {
				return &errors.InvalidFieldError{
					FieldName:   f.Name,
					ValidRange:  fmt.Sprintf("%d..%d", *f.RangeMin, *f.RangeMax),
					ActualValue: val,
				}
			}
		}

		// Validate range for unsigned
		if f.Type == schema.Uint {
			maxVal := uint64(0)
			if f.BitWidth < 64 {
				maxVal = (1 << uint(f.BitWidth)) - 1
			} else {
				maxVal = ^uint64(0)
			}
			if val > maxVal {
				return &errors.InvalidFieldError{
					FieldName:   f.Name,
					ValidRange:  fmt.Sprintf("0..%d", maxVal),
					ActualValue: val,
				}
			}
		}

		return w.WriteBits(val, f.BitWidth, order)

	default:
		return fmt.Errorf("unsupported field type %v for field %q", f.Type, f.Name)
	}
}

// resolveDisplayValue handles fields with a display format. If the packet
// value is a string, it uses the format's Decode method to convert to a raw
// numeric value. If the value is already numeric, it returns it directly.
func (enc *Encoder) resolveDisplayValue(f schema.FieldDef, packet map[string]any) (uint64, error) {
	raw, ok := packet[f.Name]
	if !ok {
		return 0, fmt.Errorf("missing field %q in packet", f.Name)
	}

	switch v := raw.(type) {
	case string:
		formatter, err := enc.formatRegistry.Get(f.DisplayFormat)
		if err != nil {
			return 0, err
		}
		decoded, err := formatter.Decode(v)
		if err != nil {
			return 0, fmt.Errorf("display decode for field %q: %w", f.Name, err)
		}
		return toUint64Value(decoded)
	default:
		return toUint64Value(raw)
	}
}

// effectiveByteOrder returns the field-level byte order if set, otherwise
// the protocol default.
func effectiveByteOrder(f schema.FieldDef, defaultOrder schema.ByteOrder) schema.ByteOrder {
	if f.ByteOrder != nil {
		return *f.ByteOrder
	}
	return defaultOrder
}

// getUint64 extracts a field value from the packet as uint64.
func getUint64(packet map[string]any, fieldName string) (uint64, error) {
	raw, ok := packet[fieldName]
	if !ok {
		return 0, fmt.Errorf("missing field %q in packet", fieldName)
	}
	return toUint64Value(raw)
}

// toUint64Value converts various numeric types to uint64.
func toUint64Value(v any) (uint64, error) {
	switch n := v.(type) {
	case uint64:
		return n, nil
	case uint32:
		return uint64(n), nil
	case uint16:
		return uint64(n), nil
	case uint8:
		return uint64(n), nil
	case uint:
		return uint64(n), nil
	case int:
		if n < 0 {
			return 0, fmt.Errorf("negative value %d cannot be encoded as unsigned", n)
		}
		return uint64(n), nil
	case int64:
		if n < 0 {
			return 0, fmt.Errorf("negative value %d cannot be encoded as unsigned", n)
		}
		return uint64(n), nil
	case int32:
		if n < 0 {
			return 0, fmt.Errorf("negative value %d cannot be encoded as unsigned", n)
		}
		return uint64(n), nil
	case int16:
		if n < 0 {
			return 0, fmt.Errorf("negative value %d cannot be encoded as unsigned", n)
		}
		return uint64(n), nil
	case int8:
		if n < 0 {
			return 0, fmt.Errorf("negative value %d cannot be encoded as unsigned", n)
		}
		return uint64(n), nil
	case float64:
		if n < 0 {
			return 0, fmt.Errorf("negative value %v cannot be encoded as unsigned", n)
		}
		return uint64(n), nil
	case float32:
		if n < 0 {
			return 0, fmt.Errorf("negative value %v cannot be encoded as unsigned", n)
		}
		return uint64(n), nil
	case bool:
		if n {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("unsupported type %T for uint64 conversion", v)
	}
}

// getBytes extracts a field value from the packet as []byte.
func getBytes(packet map[string]any, fieldName string) ([]byte, error) {
	raw, ok := packet[fieldName]
	if !ok {
		return nil, fmt.Errorf("missing field %q in packet", fieldName)
	}
	switch v := raw.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return nil, fmt.Errorf("field %q: expected []byte or string, got %T", fieldName, raw)
	}
}

// evaluateCondition evaluates a ConditionExpr against packet values.
// Returns true if the condition is satisfied.
func evaluateCondition(cond *schema.ConditionExpr, packet map[string]any) (bool, error) {
	lhs, err := getUint64(packet, cond.FieldName)
	if err != nil {
		return false, err
	}

	rhs, err := toUint64Value(cond.Value)
	if err != nil {
		return false, fmt.Errorf("condition value for field %q: %w", cond.FieldName, err)
	}

	switch cond.Operator {
	case ">":
		return lhs > rhs, nil
	case "<":
		return lhs < rhs, nil
	case ">=":
		return lhs >= rhs, nil
	case "<=":
		return lhs <= rhs, nil
	case "==":
		return lhs == rhs, nil
	case "!=":
		return lhs != rhs, nil
	default:
		return false, fmt.Errorf("unsupported condition operator %q", cond.Operator)
	}
}

// flatFieldNames returns a flat list of all field names in order, expanding
// bitfield groups into their sub-fields.
func flatFieldNames(fields []schema.FieldDef) []string {
	var names []string
	for _, f := range fields {
		if f.IsBitfieldGroup {
			for _, sub := range f.BitfieldFields {
				names = append(names, sub.Name)
			}
		} else {
			names = append(names, f.Name)
		}
	}
	return names
}

// resolveCoverFields expands a list of cover field specifiers (which may
// include range notation like "field1..field2") into a flat list of field
// names.
func resolveCoverFields(coverFields []string, allFieldNames []string) ([]string, error) {
	var result []string
	for _, spec := range coverFields {
		// Check for range notation "field1..field2"
		if idx := findRangeSeparator(spec); idx >= 0 {
			start := spec[:idx]
			end := spec[idx+2:]
			expanded, err := expandRange(start, end, allFieldNames)
			if err != nil {
				return nil, err
			}
			result = append(result, expanded...)
		} else {
			result = append(result, spec)
		}
	}
	return result, nil
}

// findRangeSeparator returns the index of ".." in s, or -1 if not found.
func findRangeSeparator(s string) int {
	for i := 0; i < len(s)-1; i++ {
		if s[i] == '.' && s[i+1] == '.' {
			return i
		}
	}
	return -1
}

// expandRange expands "start..end" into all field names between start and end
// (inclusive) in the allFieldNames list.
func expandRange(start, end string, allFieldNames []string) ([]string, error) {
	startIdx := -1
	endIdx := -1
	for i, name := range allFieldNames {
		if name == start {
			startIdx = i
		}
		if name == end {
			endIdx = i
		}
	}
	if startIdx < 0 {
		return nil, fmt.Errorf("range start field %q not found", start)
	}
	if endIdx < 0 {
		return nil, fmt.Errorf("range end field %q not found", end)
	}
	if startIdx > endIdx {
		return nil, fmt.Errorf("range start %q comes after end %q", start, end)
	}
	return allFieldNames[startIdx : endIdx+1], nil
}
