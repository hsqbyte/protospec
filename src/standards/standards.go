// Package standards provides international protocol standard integration.
package standards

import (
	"fmt"
	"strings"
)

// StandardOrg represents a standards organization.
type StandardOrg string

const (
	OrgIETF StandardOrg = "IETF"
	OrgIEEE StandardOrg = "IEEE"
	OrgITU  StandardOrg = "ITU-T"
	Org3GPP StandardOrg = "3GPP"
)

// Standard represents a protocol standard reference.
type Standard struct {
	ID       string      `json:"id"`
	Title    string      `json:"title"`
	Org      StandardOrg `json:"org"`
	Version  string      `json:"version"`
	Protocol string      `json:"protocol"`
	URL      string      `json:"url"`
}

// Registry holds known standards.
type Registry struct {
	Standards []Standard
}

// NewRegistry creates a new standards registry.
func NewRegistry() *Registry {
	return &Registry{
		Standards: []Standard{
			{ID: "RFC 791", Title: "Internet Protocol", Org: OrgIETF, Protocol: "IPv4", URL: "https://tools.ietf.org/html/rfc791"},
			{ID: "RFC 793", Title: "Transmission Control Protocol", Org: OrgIETF, Protocol: "TCP", URL: "https://tools.ietf.org/html/rfc793"},
			{ID: "RFC 768", Title: "User Datagram Protocol", Org: OrgIETF, Protocol: "UDP", URL: "https://tools.ietf.org/html/rfc768"},
			{ID: "IEEE 802.3", Title: "Ethernet", Org: OrgIEEE, Protocol: "Ethernet", URL: "https://standards.ieee.org/standard/802_3.html"},
			{ID: "IEEE 802.11", Title: "Wireless LAN", Org: OrgIEEE, Protocol: "IEEE80211", URL: "https://standards.ieee.org/standard/802_11.html"},
		},
	}
}

// FindByProtocol finds standards for a protocol.
func (r *Registry) FindByProtocol(protocol string) []Standard {
	var results []Standard
	for _, s := range r.Standards {
		if strings.EqualFold(s.Protocol, protocol) {
			results = append(results, s)
		}
	}
	return results
}

// FindByOrg finds standards by organization.
func (r *Registry) FindByOrg(org StandardOrg) []Standard {
	var results []Standard
	for _, s := range r.Standards {
		if s.Org == org {
			results = append(results, s)
		}
	}
	return results
}

// Describe returns a text listing.
func (r *Registry) Describe() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Standards Registry (%d entries):\n", len(r.Standards)))
	for _, s := range r.Standards {
		b.WriteString(fmt.Sprintf("  [%s] %s — %s (%s)\n", s.Org, s.ID, s.Title, s.Protocol))
	}
	return b.String()
}
