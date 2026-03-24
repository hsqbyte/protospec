// Package evolution tracks protocol evolution and deprecation.
package evolution

import (
	"fmt"
	"strings"
)

// ProtocolFamily represents a family of related protocols.
type ProtocolFamily struct {
	Name     string            `json:"name"`
	Versions []ProtocolVersion `json:"versions"`
}

// ProtocolVersion represents a specific protocol version.
type ProtocolVersion struct {
	Name       string   `json:"name"`
	Version    string   `json:"version"`
	RFC        string   `json:"rfc,omitempty"`
	Year       int      `json:"year,omitempty"`
	Status     string   `json:"status"` // "active", "deprecated", "obsolete"
	Features   []string `json:"features,omitempty"`
	ReplacedBy string   `json:"replaced_by,omitempty"`
}

// DeprecationWarning represents a deprecation warning.
type DeprecationWarning struct {
	Protocol    string `json:"protocol"`
	Field       string `json:"field,omitempty"`
	Reason      string `json:"reason"`
	Alternative string `json:"alternative"`
	Severity    string `json:"severity"` // "info", "warning", "critical"
}

// Tracker tracks protocol evolution.
type Tracker struct {
	families map[string]*ProtocolFamily
}

// NewTracker creates a new evolution tracker.
func NewTracker() *Tracker {
	t := &Tracker{families: make(map[string]*ProtocolFamily)}
	t.registerBuiltins()
	return t
}

// GetFamily returns a protocol family by name.
func (t *Tracker) GetFamily(name string) *ProtocolFamily {
	return t.families[name]
}

// CheckDeprecation checks if a protocol is deprecated.
func (t *Tracker) CheckDeprecation(protoName string) []DeprecationWarning {
	var warnings []DeprecationWarning

	// Check known deprecations
	deprecations := map[string]DeprecationWarning{
		"SSLv2":   {Protocol: "SSLv2", Reason: "known security vulnerabilities", Alternative: "TLS 1.3", Severity: "critical"},
		"SSLv3":   {Protocol: "SSLv3", Reason: "POODLE attack vulnerability", Alternative: "TLS 1.2+", Severity: "critical"},
		"TLSv1.0": {Protocol: "TLSv1.0", Reason: "deprecated by RFC 8996", Alternative: "TLS 1.2+", Severity: "warning"},
		"TLSv1.1": {Protocol: "TLSv1.1", Reason: "deprecated by RFC 8996", Alternative: "TLS 1.2+", Severity: "warning"},
		"FTP":     {Protocol: "FTP", Reason: "plaintext credentials", Alternative: "SFTP/SCP", Severity: "warning"},
		"Telnet":  {Protocol: "Telnet", Reason: "plaintext communication", Alternative: "SSH", Severity: "warning"},
	}

	if w, ok := deprecations[protoName]; ok {
		warnings = append(warnings, w)
	}
	return warnings
}

func (t *Tracker) registerBuiltins() {
	t.families["IP"] = &ProtocolFamily{
		Name: "Internet Protocol",
		Versions: []ProtocolVersion{
			{Name: "IPv4", Version: "4", RFC: "RFC 791", Year: 1981, Status: "active"},
			{Name: "IPv6", Version: "6", RFC: "RFC 8200", Year: 2017, Status: "active"},
		},
	}
	t.families["HTTP"] = &ProtocolFamily{
		Name: "Hypertext Transfer Protocol",
		Versions: []ProtocolVersion{
			{Name: "HTTP/1.0", Version: "1.0", RFC: "RFC 1945", Year: 1996, Status: "deprecated", ReplacedBy: "HTTP/1.1"},
			{Name: "HTTP/1.1", Version: "1.1", RFC: "RFC 9110", Year: 2022, Status: "active"},
			{Name: "HTTP/2", Version: "2", RFC: "RFC 9113", Year: 2022, Status: "active"},
			{Name: "HTTP/3", Version: "3", RFC: "RFC 9114", Year: 2022, Status: "active"},
		},
	}
	t.families["TLS"] = &ProtocolFamily{
		Name: "Transport Layer Security",
		Versions: []ProtocolVersion{
			{Name: "SSLv3", Version: "3.0", RFC: "RFC 6101", Year: 1996, Status: "obsolete", ReplacedBy: "TLS 1.0"},
			{Name: "TLS 1.0", Version: "1.0", RFC: "RFC 2246", Year: 1999, Status: "deprecated", ReplacedBy: "TLS 1.2"},
			{Name: "TLS 1.2", Version: "1.2", RFC: "RFC 5246", Year: 2008, Status: "active"},
			{Name: "TLS 1.3", Version: "1.3", RFC: "RFC 8446", Year: 2018, Status: "active"},
		},
	}
}

// FormatFamily formats a protocol family as a timeline.
func FormatFamily(f *ProtocolFamily) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Protocol Family: %s\n", f.Name))
	b.WriteString(strings.Repeat("─", 50) + "\n")
	for _, v := range f.Versions {
		icon := "●"
		if v.Status == "deprecated" {
			icon = "○"
		} else if v.Status == "obsolete" {
			icon = "✗"
		}
		line := fmt.Sprintf("%s %d  %s %s", icon, v.Year, v.Name, v.RFC)
		if v.ReplacedBy != "" {
			line += fmt.Sprintf(" → %s", v.ReplacedBy)
		}
		b.WriteString(line + "\n")
	}
	return b.String()
}
