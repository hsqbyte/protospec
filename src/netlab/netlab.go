// Package netlab provides virtual network simulation for protocol testing.
package netlab

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// NetworkConfig defines a virtual network topology.
type NetworkConfig struct {
	Name  string       `json:"name"`
	Nodes []NodeConfig `json:"nodes"`
	Links []LinkConfig `json:"links"`
}

// NodeConfig defines a virtual network node.
type NodeConfig struct {
	ID       string `json:"id"`
	Type     string `json:"type"` // "host", "router", "switch"
	IP       string `json:"ip"`
	Protocol string `json:"protocol,omitempty"`
}

// LinkConfig defines a virtual network link.
type LinkConfig struct {
	From      string        `json:"from"`
	To        string        `json:"to"`
	Latency   time.Duration `json:"latency"`
	LossRate  float64       `json:"loss_rate"` // 0.0 to 1.0
	Bandwidth int64         `json:"bandwidth"` // bytes per second
}

// Network is a running virtual network.
type Network struct {
	config *NetworkConfig
	nodes  map[string]*VirtualNode
	rng    *rand.Rand
}

// VirtualNode is a simulated network node.
type VirtualNode struct {
	Config   NodeConfig
	Sent     int64
	Received int64
	Dropped  int64
}

// TrafficPattern defines how traffic is generated.
type TrafficPattern struct {
	Type     string        `json:"type"` // "constant", "burst", "poisson"
	Rate     int           `json:"rate"` // packets per second
	Duration time.Duration `json:"duration"`
	Size     int           `json:"size"` // packet size in bytes
}

// NewNetwork creates a virtual network from config.
func NewNetwork(config *NetworkConfig) *Network {
	n := &Network{
		config: config,
		nodes:  make(map[string]*VirtualNode),
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	for _, nc := range config.Nodes {
		n.nodes[nc.ID] = &VirtualNode{Config: nc}
	}
	return n
}

// SimulateTraffic simulates traffic between two nodes.
func (n *Network) SimulateTraffic(from, to string, pattern TrafficPattern) *SimResult {
	srcNode := n.nodes[from]
	dstNode := n.nodes[to]
	if srcNode == nil || dstNode == nil {
		return &SimResult{Error: "node not found"}
	}

	// Find link
	var link *LinkConfig
	for _, l := range n.config.Links {
		if (l.From == from && l.To == to) || (l.From == to && l.To == from) {
			link = &l
			break
		}
	}

	totalPackets := pattern.Rate * int(pattern.Duration.Seconds())
	if totalPackets == 0 {
		totalPackets = pattern.Rate
	}

	result := &SimResult{
		From:         from,
		To:           to,
		TotalPackets: totalPackets,
	}

	for i := 0; i < totalPackets; i++ {
		srcNode.Sent++
		if link != nil && n.rng.Float64() < link.LossRate {
			srcNode.Dropped++
			result.Dropped++
			continue
		}
		dstNode.Received++
		result.Delivered++
		if link != nil {
			result.AvgLatency += link.Latency
		}
	}

	if result.Delivered > 0 {
		result.AvgLatency /= time.Duration(result.Delivered)
	}
	result.DeliveryRate = float64(result.Delivered) / float64(result.TotalPackets)
	return result
}

// SimResult holds simulation results.
type SimResult struct {
	From         string        `json:"from"`
	To           string        `json:"to"`
	TotalPackets int           `json:"total_packets"`
	Delivered    int           `json:"delivered"`
	Dropped      int           `json:"dropped"`
	DeliveryRate float64       `json:"delivery_rate"`
	AvgLatency   time.Duration `json:"avg_latency"`
	Error        string        `json:"error,omitempty"`
}

// FormatResult formats simulation results.
func FormatResult(r *SimResult) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Simulation: %s → %s\n", r.From, r.To))
	b.WriteString(fmt.Sprintf("  Packets: %d sent, %d delivered, %d dropped\n", r.TotalPackets, r.Delivered, r.Dropped))
	b.WriteString(fmt.Sprintf("  Delivery rate: %.1f%%\n", r.DeliveryRate*100))
	b.WriteString(fmt.Sprintf("  Avg latency: %v\n", r.AvgLatency))
	return b.String()
}
