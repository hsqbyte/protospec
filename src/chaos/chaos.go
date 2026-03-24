// Package chaos provides protocol chaos engineering tools.
package chaos

import (
	"fmt"
	"math/rand"
	"strings"
)

// FaultType represents a chaos fault type.
type FaultType string

const (
	FaultCorrupt FaultType = "corrupt"
	FaultDelay   FaultType = "delay"
	FaultDrop    FaultType = "drop"
	FaultReorder FaultType = "reorder"
	FaultDegrade FaultType = "degrade"
)

// Fault represents a chaos fault injection.
type Fault struct {
	Type        FaultType `json:"type"`
	Probability float64   `json:"probability"` // 0.0-1.0
	Config      string    `json:"config"`
}

// Scenario represents a chaos test scenario.
type Scenario struct {
	Name     string  `json:"name"`
	Protocol string  `json:"protocol"`
	Target   string  `json:"target"`
	Faults   []Fault `json:"faults"`
}

// NewScenario creates a new chaos scenario.
func NewScenario(name, protocol, target string) *Scenario {
	return &Scenario{Name: name, Protocol: protocol, Target: target}
}

// AddFault adds a fault injection.
func (s *Scenario) AddFault(ft FaultType, prob float64, config string) {
	s.Faults = append(s.Faults, Fault{Type: ft, Probability: prob, Config: config})
}

// CorruptField randomly corrupts a byte in the data.
func CorruptField(data []byte) []byte {
	if len(data) == 0 {
		return data
	}
	result := make([]byte, len(data))
	copy(result, data)
	idx := rand.Intn(len(result))
	result[idx] ^= byte(rand.Intn(255) + 1)
	return result
}

// Describe returns a scenario description.
func (s *Scenario) Describe() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Chaos Scenario: %s\n", s.Name))
	b.WriteString(fmt.Sprintf("  Protocol: %s → %s\n", s.Protocol, s.Target))
	for _, f := range s.Faults {
		b.WriteString(fmt.Sprintf("  [%s] probability=%.0f%% %s\n", f.Type, f.Probability*100, f.Config))
	}
	return b.String()
}
