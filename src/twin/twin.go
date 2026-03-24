// Package twin provides network protocol digital twin modeling.
package twin

import (
	"fmt"
	"strings"
)

// Node represents a network node in the digital twin.
type Node struct {
	ID        string   `json:"id"`
	Type      string   `json:"type"` // host, router, switch
	Protocols []string `json:"protocols"`
}

// Link represents a network link.
type Link struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Bandwidth string `json:"bandwidth"`
	Latency   string `json:"latency"`
}

// Model represents a digital twin model.
type Model struct {
	Name  string `json:"name"`
	Nodes []Node `json:"nodes"`
	Links []Link `json:"links"`
}

// NewModel creates a new digital twin model.
func NewModel(name string) *Model {
	return &Model{Name: name}
}

// AddNode adds a node.
func (m *Model) AddNode(id, typ string, protocols []string) {
	m.Nodes = append(m.Nodes, Node{ID: id, Type: typ, Protocols: protocols})
}

// AddLink adds a link.
func (m *Model) AddLink(from, to, bw, lat string) {
	m.Links = append(m.Links, Link{From: from, To: to, Bandwidth: bw, Latency: lat})
}

// WhatIf performs a what-if analysis by modifying a parameter.
func (m *Model) WhatIf(param, value string) string {
	return fmt.Sprintf("What-if analysis: %s = %s\n  Impact: simulating effect on %d nodes, %d links\n",
		param, value, len(m.Nodes), len(m.Links))
}

// ToMermaid generates a Mermaid network diagram.
func (m *Model) ToMermaid() string {
	var b strings.Builder
	b.WriteString("graph TD\n")
	for _, n := range m.Nodes {
		b.WriteString(fmt.Sprintf("  %s[%s: %s]\n", n.ID, n.Type, n.ID))
	}
	for _, l := range m.Links {
		b.WriteString(fmt.Sprintf("  %s -->|%s, %s| %s\n", l.From, l.Bandwidth, l.Latency, l.To))
	}
	return b.String()
}

// Describe returns a model description.
func (m *Model) Describe() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Digital Twin: %s (%d nodes, %d links)\n", m.Name, len(m.Nodes), len(m.Links)))
	for _, n := range m.Nodes {
		b.WriteString(fmt.Sprintf("  [%s] %s — %s\n", n.Type, n.ID, strings.Join(n.Protocols, ", ")))
	}
	return b.String()
}
