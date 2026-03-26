// Package mesh provides Service Mesh protocol integration.
package mesh

import (
	"fmt"
	"strings"
)

// EnvoyFilter represents an Envoy filter configuration.
type EnvoyFilter struct {
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Action   string `json:"action"` // inspect, transform, reject
}

// TrafficPolicy represents an Istio traffic policy.
type TrafficPolicy struct {
	Protocol    string `json:"protocol"`
	RateLimit   int    `json:"rate_limit"`
	RetryPolicy int    `json:"retry_count"`
	Timeout     string `json:"timeout"`
}

// AccessRule represents a protocol-level access control rule.
type AccessRule struct {
	Protocol string   `json:"protocol"`
	Allow    []string `json:"allow"`
	Deny     []string `json:"deny"`
}

// Config holds service mesh configuration.
type Config struct {
	Filters  []EnvoyFilter   `json:"filters"`
	Policies []TrafficPolicy `json:"policies"`
	Rules    []AccessRule    `json:"rules"`
}

// NewConfig creates a default service mesh config.
func NewConfig(protocol string) *Config {
	return &Config{
		Filters: []EnvoyFilter{
			{Name: protocol + "-inspector", Protocol: protocol, Action: "inspect"},
		},
		Policies: []TrafficPolicy{
			{Protocol: protocol, RateLimit: 1000, RetryPolicy: 3, Timeout: "30s"},
		},
		Rules: []AccessRule{
			{Protocol: protocol, Allow: []string{"service-a", "service-b"}, Deny: []string{"*"}},
		},
	}
}

// GenerateEnvoyConfig generates Envoy filter YAML.
func (c *Config) GenerateEnvoyConfig() string {
	var b strings.Builder
	b.WriteString("# Envoy Protocol-Aware Filters\n")
	for _, f := range c.Filters {
		b.WriteString(fmt.Sprintf("- name: %s\n  protocol: %s\n  action: %s\n", f.Name, f.Protocol, f.Action))
	}
	return b.String()
}

// Describe returns a config description.
func (c *Config) Describe() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Service Mesh Config (%d filters, %d policies, %d rules)\n",
		len(c.Filters), len(c.Policies), len(c.Rules)))
	for _, f := range c.Filters {
		b.WriteString(fmt.Sprintf("  filter: %s [%s]\n", f.Name, f.Action))
	}
	for _, p := range c.Policies {
		b.WriteString(fmt.Sprintf("  policy: %s (rate=%d, retries=%d, timeout=%s)\n", p.Protocol, p.RateLimit, p.RetryPolicy, p.Timeout))
	}
	return b.String()
}
