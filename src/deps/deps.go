// Package deps provides protocol dependency analysis and impact assessment.
package deps

import (
	"fmt"
	"strings"

	"github.com/hsqbyte/protospec/src/protocol"
)

// DepNode represents a node in the dependency graph.
type DepNode struct {
	Name      string   `json:"name"`
	DependsOn []string `json:"depends_on"`
	UsedBy    []string `json:"used_by"`
}

// Graph represents a protocol dependency graph.
type Graph struct {
	Nodes map[string]*DepNode `json:"nodes"`
}

// BuildGraph builds a dependency graph from the protocol library.
func BuildGraph(lib *protocol.Library) *Graph {
	g := &Graph{Nodes: make(map[string]*DepNode)}

	for _, name := range lib.AllNames() {
		node := &DepNode{Name: name}
		meta := lib.Meta(name)
		if meta != nil && len(meta.DependsOn) > 0 {
			node.DependsOn = meta.DependsOn
		}
		g.Nodes[name] = node
	}

	// Build reverse dependencies
	for name, node := range g.Nodes {
		for _, dep := range node.DependsOn {
			if depNode, ok := g.Nodes[dep]; ok {
				depNode.UsedBy = append(depNode.UsedBy, name)
			}
		}
	}
	return g
}

// DetectCycles detects circular dependencies.
func (g *Graph) DetectCycles() [][]string {
	var cycles [][]string
	visited := make(map[string]bool)
	path := make(map[string]bool)

	var dfs func(name string, trail []string)
	dfs = func(name string, trail []string) {
		if path[name] {
			// Found cycle
			start := 0
			for i, n := range trail {
				if n == name {
					start = i
					break
				}
			}
			cycle := append(trail[start:], name)
			cycles = append(cycles, cycle)
			return
		}
		if visited[name] {
			return
		}
		visited[name] = true
		path[name] = true
		trail = append(trail, name)

		if node, ok := g.Nodes[name]; ok {
			for _, dep := range node.DependsOn {
				dfs(dep, trail)
			}
		}
		path[name] = false
	}

	for name := range g.Nodes {
		dfs(name, nil)
	}
	return cycles
}

// Impact returns all protocols affected by changing the given protocol.
func (g *Graph) Impact(name string) []string {
	affected := make(map[string]bool)
	var walk func(n string)
	walk = func(n string) {
		if node, ok := g.Nodes[n]; ok {
			for _, user := range node.UsedBy {
				if !affected[user] {
					affected[user] = true
					walk(user)
				}
			}
		}
	}
	walk(name)

	var result []string
	for n := range affected {
		result = append(result, n)
	}
	return result
}

// FormatTree formats dependency tree as ASCII art.
func FormatTree(g *Graph, root string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s\n", root))
	if node, ok := g.Nodes[root]; ok {
		for i, dep := range node.DependsOn {
			prefix := "├── "
			if i == len(node.DependsOn)-1 {
				prefix = "└── "
			}
			b.WriteString(fmt.Sprintf("%s%s\n", prefix, dep))
		}
	}
	return b.String()
}
