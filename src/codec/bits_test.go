package codec

import (
	"testing"

	"github.com/hsqbyte/protospec/src/schema"
)

func TestBitWriterReaderRoundtrip_SingleByte(t *testing.T) {
	w := NewBitWriter()
	// Write 4 bits (0xA = 1010) + 4 bits (0x5 = 0101) = 0xA5
	if err := w.WriteBits(0xA, 4, schema.BigEndian); err != nil {
		t.Fatal(err)
	}
	if err := w.WriteBits(0x5, 4, schema.BigEndian); err != nil {
		t.Fatal(err)
	}
	got := w.Bytes()
	if len(got) != 1 || got[0] != 0xA5 {
		t.Fatalf("expected [0xA5], got %x", got)
	}

	r := NewBitReader(got)
	v1, err := r.ReadBits(4, schema.BigEndian)
	if err != nil {
		t.Fatal(err)
	}
	if v1 != 0xA {
		t.Fatalf("expected 0xA, got 0x%X", v1)
	}
	v2, err := r.ReadBits(4, schema.BigEndian)
	if err != nil {
		t.Fatal(err)
	}
	if v2 != 0x5 {
		t.Fatalf("expected 0x5, got 0x%X", v2)
	}
}

func TestBitWriterReaderRoundtrip_16BitBigEndian(t *testing.T) {
	w := NewBitWriter()
	if err := w.WriteBits(0x1234, 16, schema.BigEndian); err != nil {
		t.Fatal(err)
	}
	got := w.Bytes()
	if len(got) != 2 || got[0] != 0x12 || got[1] != 0x34 {
		t.Fatalf("expected [0x12, 0x34], got %x", got)
	}

	r := NewBitReader(got)
	v, err := r.ReadBits(16, schema.BigEndian)
	if err != nil {
		t.Fatal(err)
	}
	if v != 0x1234 {
		t.Fatalf("expected 0x1234, got 0x%X", v)
	}
}

func TestBitWriterReaderRoundtrip_16BitLittleEndian(t *testing.T) {
	w := NewBitWriter()
	if err := w.WriteBits(0x1234, 16, schema.LittleEndian); err != nil {
		t.Fatal(err)
	}
	got := w.Bytes()
	// Little-endian: 0x34 first, then 0x12
	if len(got) != 2 || got[0] != 0x34 || got[1] != 0x12 {
		t.Fatalf("expected [0x34, 0x12], got %x", got)
	}

	r := NewBitReader(got)
	v, err := r.ReadBits(16, schema.LittleEndian)
	if err != nil {
		t.Fatal(err)
	}
	if v != 0x1234 {
		t.Fatalf("expected 0x1234, got 0x%X", v)
	}
}

func TestBitWriterReaderRoundtrip_32BitBigEndian(t *testing.T) {
	w := NewBitWriter()
	if err := w.WriteBits(0xDEADBEEF, 32, schema.BigEndian); err != nil {
		t.Fatal(err)
	}
	got := w.Bytes()
	if len(got) != 4 {
		t.Fatalf("expected 4 bytes, got %d", len(got))
	}

	r := NewBitReader(got)
	v, err := r.ReadBits(32, schema.BigEndian)
	if err != nil {
		t.Fatal(err)
	}
	if v != 0xDEADBEEF {
		t.Fatalf("expected 0xDEADBEEF, got 0x%X", v)
	}
}

func TestBitWriterReaderRoundtrip_Bitfield(t *testing.T) {
	// Simulate IPv4-like: version(4) + ihl(4) + tos(8) + totalLength(16)
	w := NewBitWriter()
	w.WriteBits(4, 4, schema.BigEndian)   // version
	w.WriteBits(5, 4, schema.BigEndian)   // ihl
	w.WriteBits(0, 8, schema.BigEndian)   // tos
	w.WriteBits(40, 16, schema.BigEndian) // totalLength

	got := w.Bytes()
	if len(got) != 4 {
		t.Fatalf("expected 4 bytes, got %d", len(got))
	}
	// version=4, ihl=5 → 0x45
	if got[0] != 0x45 {
		t.Fatalf("expected first byte 0x45, got 0x%02X", got[0])
	}

	r := NewBitReader(got)
	ver, _ := r.ReadBits(4, schema.BigEndian)
	ihl, _ := r.ReadBits(4, schema.BigEndian)
	tos, _ := r.ReadBits(8, schema.BigEndian)
	tl, _ := r.ReadBits(16, schema.BigEndian)

	if ver != 4 || ihl != 5 || tos != 0 || tl != 40 {
		t.Fatalf("got ver=%d ihl=%d tos=%d tl=%d", ver, ihl, tos, tl)
	}
}

func TestBitReader_InsufficientData(t *testing.T) {
	r := NewBitReader([]byte{0xFF})
	_, err := r.ReadBits(16, schema.BigEndian)
	if err == nil {
		t.Fatal("expected error for insufficient data")
	}
}

func TestBitWriter_WriteByteSlice(t *testing.T) {
	w := NewBitWriter()
	data := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	if err := w.WriteByteSlice(data); err != nil {
		t.Fatal(err)
	}
	got := w.Bytes()
	if len(got) != 4 {
		t.Fatalf("expected 4 bytes, got %d", len(got))
	}
	for i, b := range data {
		if got[i] != b {
			t.Fatalf("byte %d: expected 0x%02X, got 0x%02X", i, b, got[i])
		}
	}
}

func TestBitWriter_WriteByteSlice_NotAligned(t *testing.T) {
	w := NewBitWriter()
	w.WriteBits(1, 1, schema.BigEndian)
	err := w.WriteByteSlice([]byte{0xFF})
	if err == nil {
		t.Fatal("expected error for non-aligned WriteByteSlice")
	}
}

func TestBitWriter_SetBytesAt(t *testing.T) {
	w := NewBitWriter()
	w.WriteBits(0, 16, schema.BigEndian)      // placeholder
	w.WriteBits(0xBEEF, 16, schema.BigEndian) // data after

	// Backfill the first 2 bytes
	w.SetBytesAt(0, []byte{0xDE, 0xAD})
	got := w.Bytes()
	if got[0] != 0xDE || got[1] != 0xAD {
		t.Fatalf("expected [0xDE, 0xAD, ...], got %x", got[:2])
	}
}

func TestBitWriter_Len_BitOffset(t *testing.T) {
	w := NewBitWriter()
	if w.Len() != 0 || w.BitOffset() != 0 {
		t.Fatal("initial state wrong")
	}
	w.WriteBits(0xFF, 8, schema.BigEndian)
	if w.Len() != 1 || w.BitOffset() != 0 {
		t.Fatalf("after 8 bits: Len=%d BitOffset=%d", w.Len(), w.BitOffset())
	}
	w.WriteBits(1, 3, schema.BigEndian)
	if w.Len() != 1 || w.BitOffset() != 3 {
		t.Fatalf("after 11 bits: Len=%d BitOffset=%d", w.Len(), w.BitOffset())
	}
}

func TestBitReader_ByteOffset_BitOffset_RemainingBits(t *testing.T) {
	r := NewBitReader([]byte{0xFF, 0x00})
	if r.ByteOffset() != 0 || r.BitOffset() != 0 || r.RemainingBits() != 16 {
		t.Fatal("initial state wrong")
	}
	r.ReadBits(4, schema.BigEndian)
	if r.ByteOffset() != 0 || r.BitOffset() != 4 || r.RemainingBits() != 12 {
		t.Fatalf("after 4 bits: byte=%d bit=%d remaining=%d", r.ByteOffset(), r.BitOffset(), r.RemainingBits())
	}
}

func TestBitReader_ReadBytes(t *testing.T) {
	r := NewBitReader([]byte{0xDE, 0xAD, 0xBE, 0xEF})
	got, err := r.ReadBytes(2)
	if err != nil {
		t.Fatal(err)
	}
	if got[0] != 0xDE || got[1] != 0xAD {
		t.Fatalf("expected [0xDE, 0xAD], got %x", got)
	}
	if r.ByteOffset() != 2 {
		t.Fatalf("expected byte offset 2, got %d", r.ByteOffset())
	}
}

func TestBitReader_ReadRemainingBytes(t *testing.T) {
	r := NewBitReader([]byte{0x01, 0x02, 0x03})
	r.ReadBytes(1)
	remaining := r.ReadRemainingBytes()
	if len(remaining) != 2 || remaining[0] != 0x02 || remaining[1] != 0x03 {
		t.Fatalf("expected [0x02, 0x03], got %x", remaining)
	}
}

func TestBitReader_ReadBytes_NotAligned(t *testing.T) {
	r := NewBitReader([]byte{0xFF, 0x00})
	r.ReadBits(1, schema.BigEndian)
	_, err := r.ReadBytes(1)
	if err == nil {
		t.Fatal("expected error for non-aligned ReadBytes")
	}
}

func TestBitReader_ReadBytes_InsufficientData(t *testing.T) {
	r := NewBitReader([]byte{0xFF})
	_, err := r.ReadBytes(2)
	if err == nil {
		t.Fatal("expected error for insufficient data")
	}
}

func TestWriteBits_InvalidNumBits(t *testing.T) {
	w := NewBitWriter()
	if err := w.WriteBits(0, 0, schema.BigEndian); err == nil {
		t.Fatal("expected error for numBits=0")
	}
	if err := w.WriteBits(0, 65, schema.BigEndian); err == nil {
		t.Fatal("expected error for numBits=65")
	}
}

func TestReadBits_InvalidNumBits(t *testing.T) {
	r := NewBitReader([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
	if _, err := r.ReadBits(0, schema.BigEndian); err == nil {
		t.Fatal("expected error for numBits=0")
	}
	if _, err := r.ReadBits(65, schema.BigEndian); err == nil {
		t.Fatal("expected error for numBits=65")
	}
}
