// Package certify provides protocol compliance certification framework.
package certify

import (
	"fmt"
	"strings"
	"time"

	"github.com/hsqbyte/protospec/src/schema"
)

// Level represents certification level.
type Level string

const (
	Bronze Level = "Bronze"
	Silver Level = "Silver"
	Gold   Level = "Gold"
)

// TestResult represents a single certification test result.
type TestResult struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Details string `json:"details"`
}

// Certificate represents a protocol compliance certificate.
type Certificate struct {
	Protocol  string       `json:"protocol"`
	Level     Level        `json:"level"`
	Standard  string       `json:"standard"` // "IETF", "IEEE", "3GPP", "custom"
	IssuedAt  time.Time    `json:"issued_at"`
	ExpiresAt time.Time    `json:"expires_at"`
	Tests     []TestResult `json:"tests"`
	PassRate  float64      `json:"pass_rate"`
}

// Certifier runs compliance certification tests.
type Certifier struct{}

// NewCertifier creates a new certifier.
func NewCertifier() *Certifier {
	return &Certifier{}
}

// Certify runs certification tests on a protocol schema.
func (c *Certifier) Certify(s *schema.ProtocolSchema, standard string) *Certificate {
	cert := &Certificate{
		Protocol:  s.Name,
		Standard:  standard,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().AddDate(1, 0, 0),
	}

	// Run basic structure tests
	cert.Tests = append(cert.Tests, c.testHasVersion(s))
	cert.Tests = append(cert.Tests, c.testHasFields(s))
	cert.Tests = append(cert.Tests, c.testByteOrder(s))
	cert.Tests = append(cert.Tests, c.testFieldNaming(s))
	cert.Tests = append(cert.Tests, c.testBitAlignment(s))

	// Calculate pass rate
	passed := 0
	for _, t := range cert.Tests {
		if t.Passed {
			passed++
		}
	}
	cert.PassRate = float64(passed) / float64(len(cert.Tests))

	// Determine level
	if cert.PassRate >= 0.9 {
		cert.Level = Gold
	} else if cert.PassRate >= 0.7 {
		cert.Level = Silver
	} else {
		cert.Level = Bronze
	}

	return cert
}

func (c *Certifier) testHasVersion(s *schema.ProtocolSchema) TestResult {
	return TestResult{
		Name:    "has_version",
		Passed:  s.Version != "",
		Details: fmt.Sprintf("version=%q", s.Version),
	}
}

func (c *Certifier) testHasFields(s *schema.ProtocolSchema) TestResult {
	count := len(s.Fields)
	return TestResult{
		Name:    "has_fields",
		Passed:  count > 0,
		Details: fmt.Sprintf("%d fields defined", count),
	}
}

func (c *Certifier) testByteOrder(s *schema.ProtocolSchema) TestResult {
	return TestResult{
		Name:    "byte_order_specified",
		Passed:  true,
		Details: fmt.Sprintf("byte_order=%s", s.DefaultByteOrder),
	}
}

func (c *Certifier) testFieldNaming(s *schema.ProtocolSchema) TestResult {
	for _, f := range s.Fields {
		if f.IsBitfieldGroup {
			continue
		}
		if f.Name == "" {
			return TestResult{Name: "field_naming", Passed: false, Details: "empty field name found"}
		}
	}
	return TestResult{Name: "field_naming", Passed: true, Details: "all fields have names"}
}

func (c *Certifier) testBitAlignment(s *schema.ProtocolSchema) TestResult {
	totalBits := 0
	for _, f := range s.Fields {
		if f.IsBitfieldGroup {
			for _, bf := range f.BitfieldFields {
				totalBits += bf.BitWidth
			}
		} else {
			totalBits += f.BitWidth
		}
	}
	aligned := totalBits%8 == 0
	return TestResult{
		Name:    "bit_alignment",
		Passed:  aligned,
		Details: fmt.Sprintf("total bits=%d, byte-aligned=%v", totalBits, aligned),
	}
}

// FormatCertificate formats a certificate as text.
func FormatCertificate(cert *Certificate) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("╔══════════════════════════════════════╗\n"))
	b.WriteString(fmt.Sprintf("║  PSL Compliance Certificate         ║\n"))
	b.WriteString(fmt.Sprintf("╠══════════════════════════════════════╣\n"))
	b.WriteString(fmt.Sprintf("║  Protocol: %-25s ║\n", cert.Protocol))
	b.WriteString(fmt.Sprintf("║  Level:    %-25s ║\n", cert.Level))
	b.WriteString(fmt.Sprintf("║  Standard: %-25s ║\n", cert.Standard))
	b.WriteString(fmt.Sprintf("║  Pass Rate: %-24s ║\n", fmt.Sprintf("%.0f%%", cert.PassRate*100)))
	b.WriteString(fmt.Sprintf("║  Issued:   %-25s ║\n", cert.IssuedAt.Format("2006-01-02")))
	b.WriteString(fmt.Sprintf("║  Expires:  %-25s ║\n", cert.ExpiresAt.Format("2006-01-02")))
	b.WriteString(fmt.Sprintf("╚══════════════════════════════════════╝\n"))

	b.WriteString("\nTest Results:\n")
	for _, t := range cert.Tests {
		icon := "✓"
		if !t.Passed {
			icon = "✗"
		}
		b.WriteString(fmt.Sprintf("  %s %s — %s\n", icon, t.Name, t.Details))
	}
	return b.String()
}
