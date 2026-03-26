// Package benchsuite provides standardized protocol performance benchmarking.
package benchsuite

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/hsqbyte/protospec/src/protocol"
)

// BenchResult holds the result of a single benchmark.
type BenchResult struct {
	Protocol    string        `json:"protocol"`
	Operation   string        `json:"operation"` // "encode", "decode"
	Iterations  int           `json:"iterations"`
	TotalTime   time.Duration `json:"total_time"`
	OpsPerSec   float64       `json:"ops_per_sec"`
	NsPerOp     int64         `json:"ns_per_op"`
	BytesPerOp  int64         `json:"bytes_per_op"`
	AllocsPerOp int64         `json:"allocs_per_op"`
}

// Suite runs protocol benchmarks.
type Suite struct {
	lib        *protocol.Library
	iterations int
}

// NewSuite creates a new benchmark suite.
func NewSuite(lib *protocol.Library, iterations int) *Suite {
	if iterations <= 0 {
		iterations = 10000
	}
	return &Suite{lib: lib, iterations: iterations}
}

// RunEncode benchmarks encoding for a protocol.
func (s *Suite) RunEncode(protoName string, packet map[string]any) (*BenchResult, error) {
	// Warmup
	for i := 0; i < 100; i++ {
		_, err := s.lib.Encode(protoName, packet)
		if err != nil {
			return nil, fmt.Errorf("encode warmup: %w", err)
		}
	}

	runtime.GC()
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	start := time.Now()
	for i := 0; i < s.iterations; i++ {
		s.lib.Encode(protoName, packet)
	}
	elapsed := time.Since(start)

	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	opsPerSec := float64(s.iterations) / elapsed.Seconds()
	nsPerOp := elapsed.Nanoseconds() / int64(s.iterations)
	bytesPerOp := int64(memAfter.TotalAlloc-memBefore.TotalAlloc) / int64(s.iterations)

	return &BenchResult{
		Protocol:   protoName,
		Operation:  "encode",
		Iterations: s.iterations,
		TotalTime:  elapsed,
		OpsPerSec:  opsPerSec,
		NsPerOp:    nsPerOp,
		BytesPerOp: bytesPerOp,
	}, nil
}

// RunDecode benchmarks decoding for a protocol.
func (s *Suite) RunDecode(protoName string, data []byte) (*BenchResult, error) {
	// Warmup
	for i := 0; i < 100; i++ {
		_, err := s.lib.Decode(protoName, data)
		if err != nil {
			return nil, fmt.Errorf("decode warmup: %w", err)
		}
	}

	runtime.GC()
	start := time.Now()
	for i := 0; i < s.iterations; i++ {
		s.lib.Decode(protoName, data)
	}
	elapsed := time.Since(start)

	opsPerSec := float64(s.iterations) / elapsed.Seconds()
	nsPerOp := elapsed.Nanoseconds() / int64(s.iterations)

	return &BenchResult{
		Protocol:   protoName,
		Operation:  "decode",
		Iterations: s.iterations,
		TotalTime:  elapsed,
		OpsPerSec:  opsPerSec,
		NsPerOp:    nsPerOp,
	}, nil
}

// FormatResults formats benchmark results as a table.
func FormatResults(results []*BenchResult) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%-15s %-8s %12s %12s %12s\n", "Protocol", "Op", "ops/sec", "ns/op", "B/op"))
	b.WriteString(strings.Repeat("-", 65) + "\n")
	for _, r := range results {
		b.WriteString(fmt.Sprintf("%-15s %-8s %12.0f %12d %12d\n",
			r.Protocol, r.Operation, r.OpsPerSec, r.NsPerOp, r.BytesPerOp))
	}
	return b.String()
}
