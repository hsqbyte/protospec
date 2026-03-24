// Package sandbox provides isolated execution environment for protocol decoding.
package sandbox

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/hsqbyte/protospec/src/protocol"
)

// Policy defines sandbox security constraints.
type Policy struct {
	MaxMemoryMB    int           `json:"max_memory_mb"`
	MaxCPUTime     time.Duration `json:"max_cpu_time"`
	MaxDecodeDepth int           `json:"max_decode_depth"`
	MaxOutputSize  int           `json:"max_output_size"`
	MaxNestLevel   int           `json:"max_nest_level"`
}

// DefaultPolicy returns a default sandbox policy.
func DefaultPolicy() *Policy {
	return &Policy{
		MaxMemoryMB:    256,
		MaxCPUTime:     5 * time.Second,
		MaxDecodeDepth: 20,
		MaxOutputSize:  10 * 1024 * 1024, // 10MB
		MaxNestLevel:   10,
	}
}

// Sandbox provides isolated protocol decoding.
type Sandbox struct {
	lib    *protocol.Library
	policy *Policy
}

// NewSandbox creates a new sandbox with the given policy.
func NewSandbox(lib *protocol.Library, policy *Policy) *Sandbox {
	if policy == nil {
		policy = DefaultPolicy()
	}
	return &Sandbox{lib: lib, policy: policy}
}

// Result holds the sandbox execution result.
type Result struct {
	Success   bool           `json:"success"`
	Data      map[string]any `json:"data,omitempty"`
	BytesRead int            `json:"bytes_read,omitempty"`
	Error     string         `json:"error,omitempty"`
	Duration  time.Duration  `json:"duration"`
	MemUsed   uint64         `json:"mem_used_bytes"`
}

// Decode decodes data in the sandbox with resource limits.
func (s *Sandbox) Decode(protoName string, data []byte) *Result {
	start := time.Now()

	// Check memory before
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), s.policy.MaxCPUTime)
	defer cancel()

	// Run decode in goroutine with timeout
	type decodeResult struct {
		packet    map[string]any
		bytesRead int
		err       error
	}
	ch := make(chan decodeResult, 1)

	go func() {
		result, err := s.lib.Decode(protoName, data)
		if err != nil {
			ch <- decodeResult{err: err}
			return
		}
		ch <- decodeResult{packet: result.Packet, bytesRead: result.BytesRead}
	}()

	select {
	case <-ctx.Done():
		return &Result{
			Error:    "timeout: decode exceeded CPU time limit",
			Duration: time.Since(start),
		}
	case r := <-ch:
		var memAfter runtime.MemStats
		runtime.ReadMemStats(&memAfter)

		if r.err != nil {
			return &Result{
				Error:    r.err.Error(),
				Duration: time.Since(start),
				MemUsed:  memAfter.Alloc - memBefore.Alloc,
			}
		}

		memUsed := memAfter.Alloc - memBefore.Alloc
		if memUsed > uint64(s.policy.MaxMemoryMB)*1024*1024 {
			return &Result{
				Error:    fmt.Sprintf("memory limit exceeded: %d MB used, limit %d MB", memUsed/1024/1024, s.policy.MaxMemoryMB),
				Duration: time.Since(start),
				MemUsed:  memUsed,
			}
		}

		return &Result{
			Success:   true,
			Data:      r.packet,
			BytesRead: r.bytesRead,
			Duration:  time.Since(start),
			MemUsed:   memUsed,
		}
	}
}
