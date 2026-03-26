// Package fingerprint provides protocol fingerprinting and traffic classification.
package fingerprint

import (
	"fmt"
	"sort"
	"strings"
)

// Signature represents a protocol fingerprint signature.
type Signature struct {
	Protocol string   `json:"protocol"`
	Magic    []byte   `json:"magic,omitempty"`
	Offset   int      `json:"offset"`
	Patterns []string `json:"patterns,omitempty"`
}

// ClassifyResult represents a traffic classification result.
type ClassifyResult struct {
	Protocol   string  `json:"protocol"`
	Confidence float64 `json:"confidence"`
	Method     string  `json:"method"` // "magic", "pattern", "heuristic"
}

// Classifier classifies network traffic by protocol.
type Classifier struct {
	signatures []Signature
}

// NewClassifier creates a new traffic classifier.
func NewClassifier() *Classifier {
	c := &Classifier{}
	c.registerDefaults()
	return c
}

// Classify classifies a packet and returns candidates.
func (c *Classifier) Classify(data []byte) []ClassifyResult {
	var results []ClassifyResult
	for _, sig := range c.signatures {
		if conf := c.matchSignature(data, sig); conf > 0 {
			results = append(results, ClassifyResult{
				Protocol: sig.Protocol, Confidence: conf, Method: "magic",
			})
		}
	}
	sort.Slice(results, func(i, j int) bool { return results[i].Confidence > results[j].Confidence })
	return results
}

func (c *Classifier) matchSignature(data []byte, sig Signature) float64 {
	if len(sig.Magic) > 0 && sig.Offset+len(sig.Magic) <= len(data) {
		match := true
		for i, b := range sig.Magic {
			if data[sig.Offset+i] != b {
				match = false
				break
			}
		}
		if match {
			return 0.95
		}
	}
	return 0
}

func (c *Classifier) registerDefaults() {
	c.signatures = []Signature{
		{Protocol: "TLS", Magic: []byte{0x16, 0x03}, Offset: 0},
		{Protocol: "HTTP", Magic: []byte("HTTP"), Offset: 0},
		{Protocol: "SSH", Magic: []byte("SSH-"), Offset: 0},
		{Protocol: "DNS", Magic: nil}, // heuristic-based
		{Protocol: "MQTT", Magic: []byte{0x10}, Offset: 0},
	}
}

// DeviceFingerprint represents a device fingerprint.
type DeviceFingerprint struct {
	OS         string `json:"os"`
	TTL        int    `json:"ttl"`
	WindowSize int    `json:"window_size"`
	MSS        int    `json:"mss"`
}

// IdentifyOS identifies OS from TCP/IP stack characteristics.
func IdentifyOS(ttl int, windowSize int) string {
	switch {
	case ttl <= 64:
		return "Linux/Unix"
	case ttl <= 128:
		return "Windows"
	case ttl <= 255:
		return "Cisco/Network Device"
	default:
		return "Unknown"
	}
}

// FormatClassification formats classification results.
func FormatClassification(results []ClassifyResult) string {
	var b strings.Builder
	for i, r := range results {
		b.WriteString(fmt.Sprintf("%d. %s (%.0f%% confidence, method: %s)\n", i+1, r.Protocol, r.Confidence*100, r.Method))
	}
	return b.String()
}
