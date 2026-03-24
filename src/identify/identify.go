// Package identify provides automatic protocol identification from binary data.
package identify

import (
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/hsqbyte/protospec/src/protocol"
)

// Candidate represents a protocol identification candidate.
type Candidate struct {
	Protocol   string  `json:"protocol"`
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason"`
}

// Identifier identifies protocols from binary data.
type Identifier struct {
	lib   *protocol.Library
	rules []SignatureRule
}

// SignatureRule defines a signature-based matching rule.
type SignatureRule struct {
	Protocol string
	Check    func(data []byte) (float64, string)
}

// NewIdentifier creates a new protocol identifier.
func NewIdentifier(lib *protocol.Library) *Identifier {
	id := &Identifier{lib: lib}
	id.registerBuiltinRules()
	return id
}

// Identify identifies the protocol of the given binary data.
func (id *Identifier) Identify(data []byte) []Candidate {
	var candidates []Candidate
	for _, rule := range id.rules {
		confidence, reason := rule.Check(data)
		if confidence > 0 {
			candidates = append(candidates, Candidate{
				Protocol:   rule.Protocol,
				Confidence: confidence,
				Reason:     reason,
			})
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Confidence > candidates[j].Confidence
	})
	if len(candidates) > 10 {
		candidates = candidates[:10]
	}
	return candidates
}

// IdentifyHex identifies protocol from hex string.
func (id *Identifier) IdentifyHex(hexStr string) ([]Candidate, error) {
	hexStr = strings.ReplaceAll(hexStr, " ", "")
	hexStr = strings.ReplaceAll(hexStr, ":", "")
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("invalid hex: %w", err)
	}
	return id.Identify(data), nil
}

func (id *Identifier) registerBuiltinRules() {
	id.rules = []SignatureRule{
		{Protocol: "IPv4", Check: func(d []byte) (float64, string) {
			if len(d) >= 20 && (d[0]>>4) == 4 {
				return 0.9, "version nibble = 4"
			}
			return 0, ""
		}},
		{Protocol: "IPv6", Check: func(d []byte) (float64, string) {
			if len(d) >= 40 && (d[0]>>4) == 6 {
				return 0.9, "version nibble = 6"
			}
			return 0, ""
		}},
		{Protocol: "Ethernet", Check: func(d []byte) (float64, string) {
			if len(d) >= 14 {
				etherType := uint16(d[12])<<8 | uint16(d[13])
				if etherType == 0x0800 || etherType == 0x0806 || etherType == 0x86DD {
					return 0.85, fmt.Sprintf("etherType=0x%04X", etherType)
				}
			}
			return 0, ""
		}},
		{Protocol: "TCP", Check: func(d []byte) (float64, string) {
			if len(d) >= 20 {
				dataOffset := (d[12] >> 4) * 4
				if dataOffset >= 20 && dataOffset <= 60 {
					return 0.5, fmt.Sprintf("data_offset=%d", dataOffset)
				}
			}
			return 0, ""
		}},
		{Protocol: "DNS", Check: func(d []byte) (float64, string) {
			if len(d) >= 12 {
				qdcount := uint16(d[4])<<8 | uint16(d[5])
				if qdcount >= 1 && qdcount <= 10 {
					return 0.6, fmt.Sprintf("qdcount=%d", qdcount)
				}
			}
			return 0, ""
		}},
		{Protocol: "TLS", Check: func(d []byte) (float64, string) {
			if len(d) >= 5 && d[0] >= 20 && d[0] <= 23 && d[1] == 3 {
				return 0.9, fmt.Sprintf("content_type=%d, version=3.%d", d[0], d[2])
			}
			return 0, ""
		}},
		{Protocol: "MQTT", Check: func(d []byte) (float64, string) {
			if len(d) >= 2 {
				pktType := d[0] >> 4
				if pktType >= 1 && pktType <= 14 {
					return 0.4, fmt.Sprintf("packet_type=%d", pktType)
				}
			}
			return 0, ""
		}},
		{Protocol: "RTP", Check: func(d []byte) (float64, string) {
			if len(d) >= 12 && (d[0]>>6) == 2 {
				return 0.6, "RTP version=2"
			}
			return 0, ""
		}},
	}
}
