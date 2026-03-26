// Package abtest provides protocol A/B testing framework.
package abtest

import (
	"fmt"
	"strings"
)

// Variant represents a protocol version variant.
type Variant struct {
	Name    string  `json:"name"`
	Version string  `json:"version"`
	Weight  float64 `json:"weight"` // 0.0-1.0
	Metrics Metrics `json:"metrics"`
}

// Metrics holds performance metrics for a variant.
type Metrics struct {
	Latency    float64 `json:"latency_ms"`
	ErrorRate  float64 `json:"error_rate"`
	Throughput float64 `json:"throughput_rps"`
}

// Experiment represents an A/B test experiment.
type Experiment struct {
	Name     string    `json:"name"`
	Protocol string    `json:"protocol"`
	Variants []Variant `json:"variants"`
}

// NewExperiment creates a new A/B test experiment.
func NewExperiment(name, protocol string) *Experiment {
	return &Experiment{Name: name, Protocol: protocol}
}

// AddVariant adds a variant.
func (e *Experiment) AddVariant(name, version string, weight float64) {
	e.Variants = append(e.Variants, Variant{Name: name, Version: version, Weight: weight})
}

// Analyze analyzes experiment results and returns a recommendation.
func (e *Experiment) Analyze() string {
	if len(e.Variants) < 2 {
		return "need at least 2 variants"
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Experiment: %s (%s)\n\n", e.Name, e.Protocol))
	best := e.Variants[0]
	for _, v := range e.Variants {
		b.WriteString(fmt.Sprintf("  %s (v%s): weight=%.0f%%, latency=%.1fms, errors=%.2f%%, throughput=%.0f rps\n",
			v.Name, v.Version, v.Weight*100, v.Metrics.Latency, v.Metrics.ErrorRate*100, v.Metrics.Throughput))
		if v.Metrics.Latency < best.Metrics.Latency && v.Metrics.ErrorRate <= best.Metrics.ErrorRate {
			best = v
		}
	}
	b.WriteString(fmt.Sprintf("\nRecommendation: %s (v%s)\n", best.Name, best.Version))
	return b.String()
}
