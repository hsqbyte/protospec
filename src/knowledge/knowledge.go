// Package knowledge provides protocol knowledge graph.
package knowledge

import (
	"fmt"
	"strings"
)

// RelationType represents a relationship type.
type RelationType string

const (
	RelEncapsulates RelationType = "encapsulates"
	RelDefinedBy    RelationType = "defined_by"
	RelUsedIn       RelationType = "used_in"
	RelSucceeds     RelationType = "succeeds"
)

// Node represents a knowledge graph node.
type Node struct {
	ID   string `json:"id"`
	Type string `json:"type"` // protocol, rfc, org, usecase
	Name string `json:"name"`
}

// Edge represents a knowledge graph edge.
type Edge struct {
	From     string       `json:"from"`
	To       string       `json:"to"`
	Relation RelationType `json:"relation"`
}

// Graph represents a protocol knowledge graph.
type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`
}

// NewGraph creates a new knowledge graph.
func NewGraph() *Graph {
	return &Graph{}
}

// AddNode adds a node.
func (g *Graph) AddNode(id, typ, name string) {
	g.Nodes = append(g.Nodes, Node{ID: id, Type: typ, Name: name})
}

// AddEdge adds an edge.
func (g *Graph) AddEdge(from, to string, rel RelationType) {
	g.Edges = append(g.Edges, Edge{From: from, To: to, Relation: rel})
}

// Query finds nodes related to a given node.
func (g *Graph) Query(nodeID string) []Edge {
	var results []Edge
	for _, e := range g.Edges {
		if e.From == nodeID || e.To == nodeID {
			results = append(results, e)
		}
	}
	return results
}

// ToMermaid generates a Mermaid graph.
func (g *Graph) ToMermaid() string {
	var b strings.Builder
	b.WriteString("graph LR\n")
	for _, e := range g.Edges {
		b.WriteString(fmt.Sprintf("  %s -->|%s| %s\n", e.From, e.Relation, e.To))
	}
	return b.String()
}

// Describe returns a text description.
func (g *Graph) Describe() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Knowledge Graph: %d nodes, %d edges\n", len(g.Nodes), len(g.Edges)))
	for _, n := range g.Nodes {
		b.WriteString(fmt.Sprintf("  [%s] %s (%s)\n", n.Type, n.Name, n.ID))
	}
	return b.String()
}
