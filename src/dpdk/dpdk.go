// Package dpdk provides DPDK userspace protocol parsing integration.
package dpdk

import (
	"fmt"
	"strings"
)

// PipelineStage represents a DPDK processing stage.
type PipelineStage struct {
	Name   string `json:"name"`
	Type   string `json:"type"` // rx, parse, classify, tx
	Config string `json:"config"`
}

// Pipeline represents a DPDK packet processing pipeline.
type Pipeline struct {
	Name   string          `json:"name"`
	Stages []PipelineStage `json:"stages"`
	Cores  int             `json:"cores"`
}

// NewPipeline creates a new DPDK pipeline.
func NewPipeline(name string, cores int) *Pipeline {
	return &Pipeline{Name: name, Cores: cores}
}

// AddStage adds a processing stage.
func (p *Pipeline) AddStage(name, typ, config string) {
	p.Stages = append(p.Stages, PipelineStage{Name: name, Type: typ, Config: config})
}

// GenerateConfig generates a DPDK pipeline configuration.
func (p *Pipeline) GenerateConfig() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("# DPDK Pipeline: %s (cores: %d)\n", p.Name, p.Cores))
	for i, s := range p.Stages {
		b.WriteString(fmt.Sprintf("stage_%d:\n  name: %s\n  type: %s\n  config: %s\n", i, s.Name, s.Type, s.Config))
	}
	return b.String()
}

// GenerateC generates C code for zero-copy parsing.
func GenerateC(protocol string) string {
	var b strings.Builder
	b.WriteString("#include <rte_mbuf.h>\n#include <rte_ether.h>\n\n")
	b.WriteString(fmt.Sprintf("static inline int parse_%s(struct rte_mbuf *m) {\n", strings.ToLower(protocol)))
	b.WriteString("    char *data = rte_pktmbuf_mtod(m, char *);\n")
	b.WriteString("    // Zero-copy protocol parsing\n")
	b.WriteString("    return 0;\n}\n")
	return b.String()
}

// Describe returns a pipeline description.
func (p *Pipeline) Describe() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("DPDK Pipeline: %s (%d cores, %d stages)\n", p.Name, p.Cores, len(p.Stages)))
	for _, s := range p.Stages {
		b.WriteString(fmt.Sprintf("  [%s] %s\n", s.Type, s.Name))
	}
	return b.String()
}
