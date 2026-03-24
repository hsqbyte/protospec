// Package gateway provides a protocol-aware API gateway.
package gateway

import (
	"encoding/json"
	"fmt"
	"strings"
)

// RouteType represents a protocol conversion route type.
type RouteType string

const (
	RouteBinaryToJSON RouteType = "binary_to_json"
	RouteJSONToBinary RouteType = "json_to_binary"
	RoutePassthrough  RouteType = "passthrough"
)

// Route defines a gateway route.
type Route struct {
	Path      string    `json:"path"`
	Protocol  string    `json:"protocol"`
	Type      RouteType `json:"type"`
	RateLimit int       `json:"rate_limit,omitempty"`
	AuthReq   bool      `json:"auth_required"`
}

// Config holds gateway configuration.
type Config struct {
	Listen string  `json:"listen"`
	Routes []Route `json:"routes"`
}

// Gateway represents a protocol API gateway.
type Gateway struct {
	Config *Config
}

// NewGateway creates a new gateway.
func NewGateway(cfg *Config) *Gateway {
	return &Gateway{Config: cfg}
}

// DefaultConfig returns a default gateway configuration.
func DefaultConfig() *Config {
	return &Config{
		Listen: ":8080",
		Routes: []Route{
			{Path: "/api/ipv4/decode", Protocol: "IPv4", Type: RouteBinaryToJSON},
			{Path: "/api/ipv4/encode", Protocol: "IPv4", Type: RouteJSONToBinary},
			{Path: "/api/dns/decode", Protocol: "DNS", Type: RouteBinaryToJSON},
		},
	}
}

// Describe returns a description of the gateway configuration.
func (g *Gateway) Describe() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Gateway listening on %s\n", g.Config.Listen))
	b.WriteString(fmt.Sprintf("Routes (%d):\n", len(g.Config.Routes)))
	for _, r := range g.Config.Routes {
		auth := ""
		if r.AuthReq {
			auth = " [auth]"
		}
		rl := ""
		if r.RateLimit > 0 {
			rl = fmt.Sprintf(" [%d req/s]", r.RateLimit)
		}
		b.WriteString(fmt.Sprintf("  %s → %s (%s)%s%s\n", r.Path, r.Protocol, r.Type, auth, rl))
	}
	return b.String()
}

// ExportConfig exports the gateway config as JSON.
func (g *Gateway) ExportConfig() string {
	data, _ := json.MarshalIndent(g.Config, "", "  ")
	return string(data)
}
