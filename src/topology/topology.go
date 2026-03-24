// Package topology provides network topology discovery from packet captures.
package topology

import (
	"fmt"
	"sort"
	"strings"
)

// Node represents a network node.
type Node struct {
	ID       string            `json:"id"`
	Type     string            `json:"type"` // "host", "router", "switch", "server"
	MAC      string            `json:"mac,omitempty"`
	IP       string            `json:"ip,omitempty"`
	Ports    []int             `json:"ports,omitempty"`
	Services []string          `json:"services,omitempty"`
	Meta     map[string]string `json:"meta,omitempty"`
}

// Link represents a connection between two nodes.
type Link struct {
	Source   string `json:"source"`
	Target   string `json:"target"`
	Protocol string `json:"protocol"`
	Packets  int    `json:"packets"`
	Bytes    int64  `json:"bytes"`
}

// Topology represents a discovered network topology.
type Topology struct {
	Nodes []Node `json:"nodes"`
	Links []Link `json:"links"`
}

// Builder builds network topology from observed traffic.
type Builder struct {
	nodes map[string]*Node
	links map[string]*Link
}

// NewBuilder creates a new topology builder.
func NewBuilder() *Builder {
	return &Builder{
		nodes: make(map[string]*Node),
		links: make(map[string]*Link),
	}
}

// AddTraffic adds observed traffic to the topology.
func (b *Builder) AddTraffic(srcIP, dstIP, protocol string, srcPort, dstPort int, bytes int64) {
	// Add/update source node
	if _, ok := b.nodes[srcIP]; !ok {
		b.nodes[srcIP] = &Node{ID: srcIP, IP: srcIP, Type: "host"}
	}
	// Add/update dest node
	if _, ok := b.nodes[dstIP]; !ok {
		b.nodes[dstIP] = &Node{ID: dstIP, IP: dstIP, Type: "host"}
	}
	if dstPort > 0 {
		node := b.nodes[dstIP]
		found := false
		for _, p := range node.Ports {
			if p == dstPort {
				found = true
				break
			}
		}
		if !found {
			node.Ports = append(node.Ports, dstPort)
		}
	}

	// Add/update link
	key := fmt.Sprintf("%s→%s", srcIP, dstIP)
	if link, ok := b.links[key]; ok {
		link.Packets++
		link.Bytes += bytes
	} else {
		b.links[key] = &Link{Source: srcIP, Target: dstIP, Protocol: protocol, Packets: 1, Bytes: bytes}
	}
}

// Build returns the completed topology.
func (b *Builder) Build() *Topology {
	t := &Topology{}
	for _, n := range b.nodes {
		sort.Ints(n.Ports)
		t.Nodes = append(t.Nodes, *n)
	}
	for _, l := range b.links {
		t.Links = append(t.Links, *l)
	}
	sort.Slice(t.Nodes, func(i, j int) bool { return t.Nodes[i].ID < t.Nodes[j].ID })
	sort.Slice(t.Links, func(i, j int) bool { return t.Links[i].Source < t.Links[j].Source })
	return t
}

// ToMermaid generates a Mermaid graph from the topology.
func (t *Topology) ToMermaid() string {
	var b strings.Builder
	b.WriteString("graph LR\n")
	for _, n := range t.Nodes {
		label := n.IP
		if len(n.Ports) > 0 {
			label += fmt.Sprintf(" [%d ports]", len(n.Ports))
		}
		b.WriteString(fmt.Sprintf("    %s[\"%s\"]\n", strings.ReplaceAll(n.ID, ".", "_"), label))
	}
	for _, l := range t.Links {
		src := strings.ReplaceAll(l.Source, ".", "_")
		dst := strings.ReplaceAll(l.Target, ".", "_")
		b.WriteString(fmt.Sprintf("    %s -->|%s %d pkts| %s\n", src, l.Protocol, l.Packets, dst))
	}
	return b.String()
}
