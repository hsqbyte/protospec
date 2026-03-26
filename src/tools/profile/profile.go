// Package profile provides protocol encoding/decoding performance profiling.
package profile

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/hsqbyte/protospec/src/protocol"
)

// FieldProfile holds profiling data for a single field.
type FieldProfile struct {
	Name     string        `json:"name"`
	Duration time.Duration `json:"duration"`
	Allocs   int64         `json:"allocs"`
	Bytes    int64         `json:"bytes"`
}

// ProfileResult holds the complete profiling result.
type ProfileResult struct {
	Protocol   string         `json:"protocol"`
	Operation  string         `json:"operation"`
	Total      time.Duration  `json:"total"`
	Fields     []FieldProfile `json:"fields"`
	TotalAlloc int64          `json:"total_alloc"`
	Bottleneck string         `json:"bottleneck"`
}

// Profiler profiles protocol operations.
type Profiler struct {
	lib *protocol.Library
}

// NewProfiler creates a new profiler.
func NewProfiler(lib *protocol.Library) *Profiler {
	return &Profiler{lib: lib}
}

// ProfileDecode profiles a decode operation.
func (p *Profiler) ProfileDecode(protoName string, data []byte, iterations int) (*ProfileResult, error) {
	if iterations <= 0 {
		iterations = 1000
	}

	runtime.GC()
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	start := time.Now()
	for i := 0; i < iterations; i++ {
		_, err := p.lib.Decode(protoName, data)
		if err != nil {
			return nil, err
		}
	}
	elapsed := time.Since(start)

	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	result := &ProfileResult{
		Protocol:   protoName,
		Operation:  "decode",
		Total:      elapsed / time.Duration(iterations),
		TotalAlloc: int64(memAfter.TotalAlloc-memBefore.TotalAlloc) / int64(iterations),
	}

	return result, nil
}

// FormatProfile formats profiling results.
func FormatProfile(r *ProfileResult) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Profile: %s %s\n", r.Protocol, r.Operation))
	b.WriteString(strings.Repeat("─", 50) + "\n")
	b.WriteString(fmt.Sprintf("  Total time:  %v per operation\n", r.Total))
	b.WriteString(fmt.Sprintf("  Alloc:       %d bytes per operation\n", r.TotalAlloc))
	if r.Bottleneck != "" {
		b.WriteString(fmt.Sprintf("  Bottleneck:  %s\n", r.Bottleneck))
	}
	return b.String()
}
