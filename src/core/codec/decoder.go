package codec

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/core/checksum"
	"github.com/hsqbyte/protospec/src/core/errors"
	"github.com/hsqbyte/protospec/src/core/format"
	"github.com/hsqbyte/protospec/src/core/schema"
)

// DecodeResult holds the decoded packet and the number of bytes consumed.
type DecodeResult struct {
	Packet    map[string]any
	BytesRead int
}

// Decoder decodes raw bytes into a Packet (map[string]any)
// according to a ProtocolSchema.
type Decoder struct {
	checksumRegistry *checksum.ChecksumRegistry
	formatRegistry   *format.FormatRegistry
}

// NewDecoder creates a new Decoder with the given registries.
func NewDecoder(cr *checksum.ChecksumRegistry, fr *format.FormatRegistry) *Decoder {
	return &Decoder{
		checksumRegistry: cr,
		formatRegistry:   fr,
	}
}

// checksumRecord records a checksum field that needs verification after all
// fields have been decoded.
type checksumRecord struct {
	field     schema.FieldDef
	byteOrder schema.ByteOrder
	storedVal uint64
}

// Decode deserialises data into a packet according to the schema.
func (dec *Decoder) Decode(s *schema.ProtocolSchema, data []byte) (*DecodeResult, error) {
	r := NewBitReader(data)
	packet := make(map[string]any)

	// Track byte ranges for each field (for checksum coverage).
	type fieldRange struct {
		start int
		end   int
	}
	fieldRanges := make(map[string]fieldRange)

	var checksums []checksumRecord

	// Flat list of all field names in order (for expanding ranges).
	allFieldNames := flatFieldNames(s.Fields)

	for i, f := range s.Fields {
		order := effectiveByteOrder(f, s.DefaultByteOrder)

		// --- Bitfield groups ---
		if f.IsBitfieldGroup {
			for _, sub := range f.BitfieldFields {
				subOrder := effectiveByteOrder(sub, s.DefaultByteOrder)
				startOff := r.ByteOffset()
				if err := dec.decodeField(r, sub, packet, subOrder); err != nil {
					return nil, err
				}
				endOff := r.ByteOffset()
				if r.BitOffset() != 0 {
					endOff++ // partial byte counts
				}
				fieldRanges[sub.Name] = fieldRange{start: startOff, end: endOff}
			}
			continue
		}

		// --- Conditional fields ---
		if f.Condition != nil {
			ok, err := evaluateCondition(f.Condition, packet)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}
		}

		// --- Checksum fields ---
		if f.Checksum != nil {
			startOff := r.ByteOffset()
			val, err := r.ReadBits(f.BitWidth, order)
			if err != nil {
				return nil, err
			}
			endOff := r.ByteOffset()
			if r.BitOffset() != 0 {
				endOff++
			}
			packet[f.Name] = val
			fieldRanges[f.Name] = fieldRange{start: startOff, end: endOff}
			checksums = append(checksums, checksumRecord{
				field:     f,
				byteOrder: order,
				storedVal: val,
			})
			continue
		}

		startOff := r.ByteOffset()

		// --- Display format fields ---
		if f.DisplayFormat != "" && (f.Type == schema.Uint || f.Type == schema.Int) {
			val, err := r.ReadBits(f.BitWidth, order)
			if err != nil {
				return nil, err
			}
			formatter, err := dec.formatRegistry.Get(f.DisplayFormat)
			if err != nil {
				return nil, err
			}
			display, err := formatter.Encode(val)
			if err != nil {
				return nil, fmt.Errorf("display encode for field %q: %w", f.Name, err)
			}
			packet[f.Name] = display
			endOff := r.ByteOffset()
			if r.BitOffset() != 0 {
				endOff++
			}
			fieldRanges[f.Name] = fieldRange{start: startOff, end: endOff}
			continue
		}

		// --- Fixed-length bytes fields ---
		if f.FixedLength > 0 && (f.Type == schema.Bytes || f.Type == schema.String) {
			bs, err := r.ReadBytes(f.FixedLength)
			if err != nil {
				return nil, err
			}
			if f.Type == schema.String {
				packet[f.Name] = string(bs)
			} else {
				packet[f.Name] = bs
			}
			endOff := r.ByteOffset()
			fieldRanges[f.Name] = fieldRange{start: startOff, end: endOff}
			continue
		}

		// --- Variable-length fields ---
		if f.LengthRef != nil && (f.Type == schema.Bytes || f.Type == schema.String) {
			refVal, err := getUint64(packet, f.LengthRef.FieldName)
			if err != nil {
				return nil, err
			}
			length := int(refVal)*f.LengthRef.Scale + f.LengthRef.Offset
			if length < 0 {
				length = 0
			}
			bs, err := r.ReadBytes(length)
			if err != nil {
				return nil, err
			}
			if f.Type == schema.String {
				packet[f.Name] = string(bs)
			} else {
				packet[f.Name] = bs
			}
			endOff := r.ByteOffset()
			fieldRanges[f.Name] = fieldRange{start: startOff, end: endOff}
			continue
		}

		// --- Payload fields (last bytes field with no LengthRef) ---
		if (f.Type == schema.Bytes || f.Type == schema.String) && f.LengthRef == nil {
			// Check if this is the last bytes/string field
			isLast := true
			for j := i + 1; j < len(s.Fields); j++ {
				candidate := s.Fields[j]
				if candidate.Type == schema.Bytes || candidate.Type == schema.String {
					if candidate.LengthRef == nil {
						isLast = false
						break
					}
				}
			}
			if isLast {
				bs := r.ReadRemainingBytes()
				if f.Type == schema.String {
					packet[f.Name] = string(bs)
				} else {
					packet[f.Name] = bs
				}
				endOff := r.ByteOffset()
				fieldRanges[f.Name] = fieldRange{start: startOff, end: endOff}
				continue
			}
		}

		// --- Fixed-width integer fields ---
		if f.Type == schema.Uint || f.Type == schema.Int {
			val, err := r.ReadBits(f.BitWidth, order)
			if err != nil {
				return nil, err
			}
			packet[f.Name] = val
			endOff := r.ByteOffset()
			if r.BitOffset() != 0 {
				endOff++
			}
			fieldRanges[f.Name] = fieldRange{start: startOff, end: endOff}
			continue
		}

		// --- Bool fields ---
		if f.Type == schema.Bool {
			val, err := r.ReadBits(1, order)
			if err != nil {
				return nil, err
			}
			packet[f.Name] = val != 0
			endOff := r.ByteOffset()
			if r.BitOffset() != 0 {
				endOff++
			}
			fieldRanges[f.Name] = fieldRange{start: startOff, end: endOff}
			continue
		}

		return nil, fmt.Errorf("unsupported field type %v for field %q", f.Type, f.Name)
	}

	// --- Verify checksums ---
	for _, cs := range checksums {
		covered, err := resolveCoverFields(cs.field.Checksum.CoverFields, allFieldNames)
		if err != nil {
			return nil, err
		}

		// Determine the byte range covered
		var minStart, maxEnd int
		first := true
		for _, name := range covered {
			rng, ok := fieldRanges[name]
			if !ok {
				return nil, fmt.Errorf("checksum cover field %q not found in decoded fields", name)
			}
			if first || rng.start < minStart {
				minStart = rng.start
			}
			if first || rng.end > maxEnd {
				maxEnd = rng.end
			}
			first = false
		}
		if first {
			return nil, fmt.Errorf("checksum for field %q covers no fields", cs.field.Name)
		}

		// Make a copy of the covered bytes
		coveredBytes := make([]byte, maxEnd-minStart)
		copy(coveredBytes, data[minStart:maxEnd])

		// Zero out the checksum field's bytes in the copy
		csRange, ok := fieldRanges[cs.field.Name]
		if ok && csRange.start >= minStart && csRange.end <= maxEnd {
			for j := csRange.start - minStart; j < csRange.end-minStart; j++ {
				coveredBytes[j] = 0
			}
		}

		fn, err := dec.checksumRegistry.Get(cs.field.Checksum.Algorithm)
		if err != nil {
			return nil, err
		}
		expected := fn(coveredBytes)

		if expected != cs.storedVal {
			return nil, &errors.ChecksumError{
				FieldName: cs.field.Name,
				Expected:  expected,
				Actual:    cs.storedVal,
			}
		}
	}

	bytesRead := r.ByteOffset()
	if r.BitOffset() != 0 {
		bytesRead++
	}

	return &DecodeResult{
		Packet:    packet,
		BytesRead: bytesRead,
	}, nil
}

// decodeField decodes a single field from the BitReader into the packet.
func (dec *Decoder) decodeField(r *BitReader, f schema.FieldDef, packet map[string]any, order schema.ByteOrder) error {
	switch {
	case f.Type == schema.Bool:
		val, err := r.ReadBits(1, order)
		if err != nil {
			return err
		}
		packet[f.Name] = val != 0
		return nil

	case f.Type == schema.Uint || f.Type == schema.Int:
		if f.DisplayFormat != "" {
			val, err := r.ReadBits(f.BitWidth, order)
			if err != nil {
				return err
			}
			formatter, fmtErr := dec.formatRegistry.Get(f.DisplayFormat)
			if fmtErr != nil {
				return fmtErr
			}
			display, encErr := formatter.Encode(val)
			if encErr != nil {
				return fmt.Errorf("display encode for field %q: %w", f.Name, encErr)
			}
			packet[f.Name] = display
			return nil
		}
		val, err := r.ReadBits(f.BitWidth, order)
		if err != nil {
			return err
		}
		packet[f.Name] = val
		return nil

	default:
		return fmt.Errorf("unsupported field type %v for field %q in bitfield group", f.Type, f.Name)
	}
}
