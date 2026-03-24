// Package forensics provides network forensics analysis capabilities.
package forensics

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"time"
)

// TimelineEvent represents an event in the forensics timeline.
type TimelineEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Protocol  string    `json:"protocol"`
	Source    string    `json:"source"`
	Dest      string    `json:"dest"`
	EventType string    `json:"event_type"` // "connection", "data", "credential", "file", "dns"
	Details   string    `json:"details"`
	Severity  string    `json:"severity"` // "info", "warning", "critical"
}

// ForensicsReport holds the complete forensics analysis report.
type ForensicsReport struct {
	Title        string          `json:"title"`
	GeneratedAt  time.Time       `json:"generated_at"`
	Hash         string          `json:"hash"`
	Timeline     []TimelineEvent `json:"timeline"`
	Credentials  []Credential    `json:"credentials,omitempty"`
	DNSQueries   []DNSQuery      `json:"dns_queries,omitempty"`
	FileExtracts []FileExtract   `json:"file_extracts,omitempty"`
	Summary      string          `json:"summary"`
}

// Credential represents a detected credential in traffic.
type Credential struct {
	Protocol string `json:"protocol"`
	Type     string `json:"type"` // "basic_auth", "form", "ftp", "telnet"
	Source   string `json:"source"`
	Details  string `json:"details"`
}

// DNSQuery represents a DNS query found in traffic.
type DNSQuery struct {
	Timestamp time.Time `json:"timestamp"`
	Domain    string    `json:"domain"`
	Type      string    `json:"type"` // "A", "AAAA", "CNAME", "MX"
	Response  string    `json:"response,omitempty"`
}

// FileExtract represents a file extracted from traffic.
type FileExtract struct {
	Filename string `json:"filename"`
	MimeType string `json:"mime_type"`
	Size     int64  `json:"size"`
	Hash     string `json:"hash"`
}

// Analyzer performs forensics analysis.
type Analyzer struct{}

// NewAnalyzer creates a new forensics analyzer.
func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

// AnalyzeData performs forensics analysis on raw packet data.
func (a *Analyzer) AnalyzeData(data []byte) *ForensicsReport {
	hash := sha256.Sum256(data)
	return &ForensicsReport{
		Title:       "Network Forensics Report",
		GeneratedAt: time.Now(),
		Hash:        fmt.Sprintf("%x", hash),
		Summary:     fmt.Sprintf("Analyzed %d bytes of capture data", len(data)),
	}
}

// FormatReport formats a forensics report as text.
func FormatReport(r *ForensicsReport) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("=== %s ===\n", r.Title))
	b.WriteString(fmt.Sprintf("Generated: %s\n", r.GeneratedAt.Format(time.RFC3339)))
	b.WriteString(fmt.Sprintf("Data Hash: %s\n", r.Hash))
	b.WriteString(fmt.Sprintf("\n%s\n", r.Summary))

	if len(r.Timeline) > 0 {
		b.WriteString("\n--- Timeline ---\n")
		for _, e := range r.Timeline {
			b.WriteString(fmt.Sprintf("[%s] %s %s→%s %s: %s\n",
				e.Severity, e.Timestamp.Format("15:04:05.000"),
				e.Source, e.Dest, e.Protocol, e.Details))
		}
	}

	if len(r.Credentials) > 0 {
		b.WriteString("\n--- Credentials Detected ---\n")
		for _, c := range r.Credentials {
			b.WriteString(fmt.Sprintf("  [%s] %s from %s: %s\n", c.Type, c.Protocol, c.Source, c.Details))
		}
	}

	if len(r.DNSQueries) > 0 {
		b.WriteString("\n--- DNS Queries ---\n")
		for _, d := range r.DNSQueries {
			b.WriteString(fmt.Sprintf("  %s %s → %s\n", d.Type, d.Domain, d.Response))
		}
	}

	return b.String()
}
