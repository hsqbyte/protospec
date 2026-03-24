// Package scanner provides protocol security scanning and vulnerability detection.
package scanner

import (
	"fmt"
	"strings"
	"time"
)

// Vulnerability represents a detected vulnerability.
type Vulnerability struct {
	ID          string `json:"id"`
	Protocol    string `json:"protocol"`
	Severity    string `json:"severity"` // "low", "medium", "high", "critical"
	Title       string `json:"title"`
	Description string `json:"description"`
	Remediation string `json:"remediation"`
	CVE         string `json:"cve,omitempty"`
}

// ScanResult holds the results of a security scan.
type ScanResult struct {
	Target          string          `json:"target"`
	Protocol        string          `json:"protocol"`
	StartTime       time.Time       `json:"start_time"`
	EndTime         time.Time       `json:"end_time"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities"`
	RiskLevel       string          `json:"risk_level"` // "low", "medium", "high", "critical"
	Score           float64         `json:"score"`      // 0-10
}

// Scanner performs protocol security scans.
type Scanner struct {
	rules []ScanRule
}

// ScanRule defines a security scanning rule.
type ScanRule struct {
	ID       string
	Name     string
	Severity string
	Check    func(protocol string, data []byte) *Vulnerability
}

// NewScanner creates a new security scanner.
func NewScanner() *Scanner {
	s := &Scanner{}
	s.registerRules()
	return s
}

// Scan performs a security scan on protocol data.
func (s *Scanner) Scan(protocol string, data []byte) *ScanResult {
	result := &ScanResult{
		Protocol:  protocol,
		StartTime: time.Now(),
	}

	for _, rule := range s.rules {
		if vuln := rule.Check(protocol, data); vuln != nil {
			result.Vulnerabilities = append(result.Vulnerabilities, *vuln)
		}
	}

	result.EndTime = time.Now()
	result.RiskLevel = calculateRisk(result.Vulnerabilities)
	result.Score = calculateScore(result.Vulnerabilities)
	return result
}

func (s *Scanner) registerRules() {
	s.rules = []ScanRule{
		{ID: "WEAK-CRYPTO", Name: "Weak Cryptography", Severity: "high",
			Check: func(proto string, data []byte) *Vulnerability {
				weakProtos := []string{"SSLv2", "SSLv3", "TLSv1.0", "RC4", "DES", "MD5"}
				for _, wp := range weakProtos {
					if strings.Contains(proto, wp) {
						return &Vulnerability{
							ID: "WEAK-CRYPTO-001", Protocol: proto, Severity: "high",
							Title:       "Weak cryptographic protocol detected",
							Description: fmt.Sprintf("Protocol %s uses weak cryptography", proto),
							Remediation: "Upgrade to TLS 1.2 or higher",
						}
					}
				}
				return nil
			}},
		{ID: "PLAINTEXT", Name: "Plaintext Credentials", Severity: "critical",
			Check: func(proto string, data []byte) *Vulnerability {
				plainProtos := []string{"FTP", "Telnet", "HTTP"}
				for _, pp := range plainProtos {
					if proto == pp {
						return &Vulnerability{
							ID: "PLAINTEXT-001", Protocol: proto, Severity: "critical",
							Title:       "Plaintext protocol detected",
							Description: fmt.Sprintf("%s transmits data in plaintext", proto),
							Remediation: "Use encrypted alternative (SFTP, SSH, HTTPS)",
						}
					}
				}
				return nil
			}},
	}
}

func calculateRisk(vulns []Vulnerability) string {
	maxSev := "low"
	for _, v := range vulns {
		if severityRank(v.Severity) > severityRank(maxSev) {
			maxSev = v.Severity
		}
	}
	return maxSev
}

func calculateScore(vulns []Vulnerability) float64 {
	if len(vulns) == 0 {
		return 0
	}
	total := 0.0
	for _, v := range vulns {
		switch v.Severity {
		case "critical":
			total += 9.0
		case "high":
			total += 7.0
		case "medium":
			total += 4.0
		case "low":
			total += 2.0
		}
	}
	score := total / float64(len(vulns))
	if score > 10 {
		score = 10
	}
	return score
}

func severityRank(s string) int {
	switch s {
	case "critical":
		return 4
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	}
	return 0
}

// FormatReport formats a scan result as a security report.
func FormatReport(r *ScanResult) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Security Scan Report — %s\n", r.Protocol))
	b.WriteString(strings.Repeat("═", 50) + "\n")
	b.WriteString(fmt.Sprintf("Risk Level: %s (Score: %.1f/10)\n", strings.ToUpper(r.RiskLevel), r.Score))
	b.WriteString(fmt.Sprintf("Duration: %v\n\n", r.EndTime.Sub(r.StartTime)))

	if len(r.Vulnerabilities) == 0 {
		b.WriteString("No vulnerabilities found.\n")
	} else {
		for i, v := range r.Vulnerabilities {
			b.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, strings.ToUpper(v.Severity), v.Title))
			b.WriteString(fmt.Sprintf("   %s\n", v.Description))
			b.WriteString(fmt.Sprintf("   Fix: %s\n\n", v.Remediation))
		}
	}
	return b.String()
}
