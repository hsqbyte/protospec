// Package sequence generates protocol sequence diagrams from packet captures.
package sequence

import (
	"fmt"
	"strings"
)

// Event represents a protocol event in a sequence.
type Event struct {
	Timestamp float64 `json:"timestamp"`
	Source    string  `json:"source"`
	Dest      string  `json:"dest"`
	Protocol  string  `json:"protocol"`
	Info      string  `json:"info"`
	Latency   float64 `json:"latency_ms,omitempty"`
}

// Diagram generates sequence diagrams from events.
type Diagram struct {
	Events []Event
	Title  string
}

// NewDiagram creates a new sequence diagram.
func NewDiagram(title string) *Diagram {
	return &Diagram{Title: title}
}

// AddEvent adds an event to the diagram.
func (d *Diagram) AddEvent(e Event) {
	d.Events = append(d.Events, e)
}

// ToMermaid generates a Mermaid sequence diagram.
func (d *Diagram) ToMermaid() string {
	var b strings.Builder
	b.WriteString("sequenceDiagram\n")
	if d.Title != "" {
		b.WriteString(fmt.Sprintf("    title %s\n", d.Title))
	}

	// Collect participants
	seen := make(map[string]bool)
	var participants []string
	for _, e := range d.Events {
		if !seen[e.Source] {
			seen[e.Source] = true
			participants = append(participants, e.Source)
		}
		if !seen[e.Dest] {
			seen[e.Dest] = true
			participants = append(participants, e.Dest)
		}
	}
	for _, p := range participants {
		b.WriteString(fmt.Sprintf("    participant %s\n", p))
	}

	for _, e := range d.Events {
		label := e.Protocol
		if e.Info != "" {
			label += ": " + e.Info
		}
		if e.Latency > 0 {
			label += fmt.Sprintf(" (%.1fms)", e.Latency)
		}
		b.WriteString(fmt.Sprintf("    %s->>%s: %s\n", e.Source, e.Dest, label))
	}
	return b.String()
}

// ToPlantUML generates a PlantUML sequence diagram.
func (d *Diagram) ToPlantUML() string {
	var b strings.Builder
	b.WriteString("@startuml\n")
	if d.Title != "" {
		b.WriteString(fmt.Sprintf("title %s\n", d.Title))
	}

	for _, e := range d.Events {
		label := e.Protocol
		if e.Info != "" {
			label += ": " + e.Info
		}
		b.WriteString(fmt.Sprintf("%s -> %s : %s\n", e.Source, e.Dest, label))
	}
	b.WriteString("@enduml\n")
	return b.String()
}

// LatencyStats computes latency statistics from events.
type LatencyStats struct {
	Min   float64 `json:"min_ms"`
	Max   float64 `json:"max_ms"`
	Avg   float64 `json:"avg_ms"`
	Count int     `json:"count"`
}

// ComputeLatency computes request-response latency statistics.
func ComputeLatency(events []Event) *LatencyStats {
	if len(events) == 0 {
		return &LatencyStats{}
	}

	stats := &LatencyStats{Min: 1e9}
	for _, e := range events {
		if e.Latency > 0 {
			stats.Count++
			stats.Avg += e.Latency
			if e.Latency < stats.Min {
				stats.Min = e.Latency
			}
			if e.Latency > stats.Max {
				stats.Max = e.Latency
			}
		}
	}
	if stats.Count > 0 {
		stats.Avg /= float64(stats.Count)
	} else {
		stats.Min = 0
	}
	return stats
}
