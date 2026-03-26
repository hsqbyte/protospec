package codec

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/core/errors"
	"github.com/hsqbyte/protospec/src/core/schema"
)

// BitWriter accumulates bits into a byte buffer.
// Bits are written MSB-first within each byte.
type BitWriter struct {
	buf     []byte // completed bytes
	current byte   // current byte being accumulated
	bitPos  int    // number of bits written into current (0-7)
}

// NewBitWriter creates a new BitWriter.
func NewBitWriter() *BitWriter {
	return &BitWriter{}
}

// WriteBits writes numBits bits (1-64) of value in the specified byte order.
// For big-endian: most significant bits are written first.
// For little-endian multi-byte values: least significant byte is written first.
func (w *BitWriter) WriteBits(value uint64, numBits int, order schema.ByteOrder) error {
	if numBits < 1 || numBits > 64 {
		return fmt.Errorf("numBits must be 1-64, got %d", numBits)
	}

	// Mask value to numBits
	if numBits < 64 {
		value &= (1 << uint(numBits)) - 1
	}

	// If the write is byte-aligned (bitPos == 0) and numBits is a multiple of 8
	// and numBits > 8, we can write whole bytes directly in the specified order.
	if w.bitPos == 0 && numBits > 8 && numBits%8 == 0 {
		numBytes := numBits / 8
		if order == schema.BigEndian {
			for i := numBytes - 1; i >= 0; i-- {
				w.buf = append(w.buf, byte(value>>(uint(i)*8)))
			}
		} else {
			// Little-endian: least significant byte first
			for i := 0; i < numBytes; i++ {
				w.buf = append(w.buf, byte(value>>(uint(i)*8)))
			}
		}
		return nil
	}

	// For non-byte-aligned writes or small writes (<=8 bits),
	// write bits one at a time from MSB to LSB of the value.
	// This handles big-endian bit order within each byte.
	for i := numBits - 1; i >= 0; i-- {
		bit := (value >> uint(i)) & 1
		w.current |= byte(bit << uint(7-w.bitPos))
		w.bitPos++
		if w.bitPos == 8 {
			w.buf = append(w.buf, w.current)
			w.current = 0
			w.bitPos = 0
		}
	}

	return nil
}

// Bytes returns the written bytes, flushing any partial byte with zero padding.
func (w *BitWriter) Bytes() []byte {
	if w.bitPos == 0 {
		result := make([]byte, len(w.buf))
		copy(result, w.buf)
		return result
	}
	result := make([]byte, len(w.buf)+1)
	copy(result, w.buf)
	result[len(w.buf)] = w.current
	return result
}

// Len returns the number of complete bytes written so far.
func (w *BitWriter) Len() int {
	return len(w.buf)
}

// BitOffset returns the current bit offset within the current byte (0-7).
func (w *BitWriter) BitOffset() int {
	return w.bitPos
}

// WriteByteSlice writes raw bytes. The writer must be byte-aligned.
func (w *BitWriter) WriteByteSlice(data []byte) error {
	if w.bitPos != 0 {
		return fmt.Errorf("WriteByteSlice requires byte-aligned position, current bit offset is %d", w.bitPos)
	}
	w.buf = append(w.buf, data...)
	return nil
}

// SetBytesAt overwrites bytes at a specific offset in the buffer.
// Used for checksum backfill after encoding.
func (w *BitWriter) SetBytesAt(offset int, data []byte) {
	copy(w.buf[offset:offset+len(data)], data)
}

// BitReader reads bits from a byte slice.
// Bits are read MSB-first within each byte.
type BitReader struct {
	data    []byte
	byteOff int // current byte offset
	bitOff  int // current bit offset within current byte (0-7)
}

// NewBitReader creates a new BitReader from a byte slice.
func NewBitReader(data []byte) *BitReader {
	return &BitReader{data: data}
}

// ReadBits reads numBits bits (1-64) in the specified byte order and returns the value.
func (r *BitReader) ReadBits(numBits int, order schema.ByteOrder) (uint64, error) {
	if numBits < 1 || numBits > 64 {
		return 0, fmt.Errorf("numBits must be 1-64, got %d", numBits)
	}

	if r.RemainingBits() < numBits {
		return 0, &errors.InsufficientDataError{
			ExpectedMin: (numBits + 7) / 8,
			ActualLen:   len(r.data) - r.byteOff,
		}
	}

	// If byte-aligned and numBits is a multiple of 8 and > 8,
	// read whole bytes in the specified order.
	if r.bitOff == 0 && numBits > 8 && numBits%8 == 0 {
		numBytes := numBits / 8
		var value uint64
		if order == schema.BigEndian {
			for i := 0; i < numBytes; i++ {
				value = (value << 8) | uint64(r.data[r.byteOff])
				r.byteOff++
			}
		} else {
			// Little-endian: first byte is least significant
			for i := 0; i < numBytes; i++ {
				value |= uint64(r.data[r.byteOff]) << (uint(i) * 8)
				r.byteOff++
			}
		}
		return value, nil
	}

	// For non-byte-aligned reads or small reads (<=8 bits),
	// read bits one at a time, MSB first.
	var value uint64
	for i := 0; i < numBits; i++ {
		bit := (r.data[r.byteOff] >> uint(7-r.bitOff)) & 1
		value = (value << 1) | uint64(bit)
		r.bitOff++
		if r.bitOff == 8 {
			r.bitOff = 0
			r.byteOff++
		}
	}

	return value, nil
}

// ReadBytes reads n bytes. The reader must be byte-aligned.
func (r *BitReader) ReadBytes(n int) ([]byte, error) {
	if r.bitOff != 0 {
		return nil, fmt.Errorf("ReadBytes requires byte-aligned position, current bit offset is %d", r.bitOff)
	}
	if r.byteOff+n > len(r.data) {
		return nil, &errors.InsufficientDataError{
			ExpectedMin: n,
			ActualLen:   len(r.data) - r.byteOff,
		}
	}
	result := make([]byte, n)
	copy(result, r.data[r.byteOff:r.byteOff+n])
	r.byteOff += n
	return result, nil
}

// ReadRemainingBytes reads all remaining bytes. Must be byte-aligned.
func (r *BitReader) ReadRemainingBytes() []byte {
	if r.bitOff != 0 {
		// Skip remaining bits in current byte
		r.bitOff = 0
		r.byteOff++
	}
	if r.byteOff >= len(r.data) {
		return nil
	}
	result := make([]byte, len(r.data)-r.byteOff)
	copy(result, r.data[r.byteOff:])
	r.byteOff = len(r.data)
	return result
}

// ByteOffset returns the current byte offset.
func (r *BitReader) ByteOffset() int {
	return r.byteOff
}

// BitOffset returns the current bit offset within the current byte (0-7).
func (r *BitReader) BitOffset() int {
	return r.bitOff
}

// RemainingBits returns the number of remaining bits.
func (r *BitReader) RemainingBits() int {
	return (len(r.data)-r.byteOff)*8 - r.bitOff
}
