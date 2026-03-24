// Package pqcrypto provides post-quantum cryptography protocol analysis.
package pqcrypto

import (
	"fmt"
	"strings"
)

// Algorithm represents a post-quantum algorithm.
type Algorithm struct {
	Name     string `json:"name"`
	Type     string `json:"type"`           // kem, signature
	Security int    `json:"security_level"` // NIST level 1-5
	KeySize  int    `json:"key_size_bytes"`
}

// Assessment represents a quantum security assessment.
type Assessment struct {
	Protocol    string      `json:"protocol"`
	Algorithms  []Algorithm `json:"algorithms"`
	Risk        string      `json:"risk"` // low, medium, high, critical
	Suggestions []string    `json:"suggestions"`
}

// KnownAlgorithms returns known PQ algorithms.
func KnownAlgorithms() []Algorithm {
	return []Algorithm{
		{Name: "Kyber-768", Type: "kem", Security: 3, KeySize: 1184},
		{Name: "Kyber-1024", Type: "kem", Security: 5, KeySize: 1568},
		{Name: "Dilithium-3", Type: "signature", Security: 3, KeySize: 1952},
		{Name: "Dilithium-5", Type: "signature", Security: 5, KeySize: 2592},
		{Name: "SPHINCS+-256f", Type: "signature", Security: 5, KeySize: 64},
	}
}

// Assess performs a quantum security assessment.
func Assess(protocol string) *Assessment {
	a := &Assessment{
		Protocol:   protocol,
		Algorithms: KnownAlgorithms(),
		Risk:       "medium",
		Suggestions: []string{
			"Enable hybrid TLS 1.3 + PQ key exchange",
			"Migrate to Kyber-768 for key encapsulation",
			"Use Dilithium-3 for digital signatures",
		},
	}
	return a
}

// Describe returns assessment description.
func (a *Assessment) Describe() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Quantum Security Assessment: %s (risk: %s)\n\n", a.Protocol, a.Risk))
	b.WriteString("Recommended PQ Algorithms:\n")
	for _, alg := range a.Algorithms {
		b.WriteString(fmt.Sprintf("  %s (%s, NIST L%d, %d bytes)\n", alg.Name, alg.Type, alg.Security, alg.KeySize))
	}
	b.WriteString("\nMigration Suggestions:\n")
	for _, s := range a.Suggestions {
		b.WriteString(fmt.Sprintf("  → %s\n", s))
	}
	return b.String()
}
