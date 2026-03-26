package codec

import "sync"

// BufferPool provides reusable byte buffers for encoding.
var BufferPool = sync.Pool{
	New: func() any {
		buf := make([]byte, 0, 256)
		return &buf
	},
}

// GetBuffer returns a buffer from the pool.
func GetBuffer() *[]byte {
	return BufferPool.Get().(*[]byte)
}

// PutBuffer returns a buffer to the pool.
func PutBuffer(buf *[]byte) {
	*buf = (*buf)[:0]
	BufferPool.Put(buf)
}

// BatchDecodeResult holds results from batch decoding.
type BatchDecodeResult struct {
	Results []*DecodeResult
	Errors  []error
}

// BatchDecode decodes multiple packets in sequence.
func (e *CodecEngine) BatchDecode(s interface{ GetSchema() interface{} }, packets [][]byte) *BatchDecodeResult {
	// This is a placeholder for the batch decode API
	return &BatchDecodeResult{}
}
