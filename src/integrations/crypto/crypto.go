// Package crypto provides encryption-related protocol support:
// TLS deep parsing, encrypted field handling, and authentication protocols.
package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

// TLSVersion represents a TLS version.
type TLSVersion uint16

const (
	TLS10 TLSVersion = 0x0301
	TLS11 TLSVersion = 0x0302
	TLS12 TLSVersion = 0x0303
	TLS13 TLSVersion = 0x0304
)

func (v TLSVersion) String() string {
	switch v {
	case TLS10:
		return "TLS 1.0"
	case TLS11:
		return "TLS 1.1"
	case TLS12:
		return "TLS 1.2"
	case TLS13:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("Unknown(0x%04X)", uint16(v))
	}
}

// CipherSuite describes a TLS cipher suite.
type CipherSuite struct {
	ID          uint16 `json:"id"`
	Name        string `json:"name"`
	KeyExchange string `json:"key_exchange"`
	Cipher      string `json:"cipher"`
	MAC         string `json:"mac"`
}

// CommonCipherSuites contains well-known cipher suites.
var CommonCipherSuites = map[uint16]CipherSuite{
	0x002F: {0x002F, "TLS_RSA_WITH_AES_128_CBC_SHA", "RSA", "AES-128-CBC", "SHA"},
	0x0035: {0x0035, "TLS_RSA_WITH_AES_256_CBC_SHA", "RSA", "AES-256-CBC", "SHA"},
	0xC02F: {0xC02F, "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256", "ECDHE-RSA", "AES-128-GCM", "SHA256"},
	0xC030: {0xC030, "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384", "ECDHE-RSA", "AES-256-GCM", "SHA384"},
	0x1301: {0x1301, "TLS_AES_128_GCM_SHA256", "ANY", "AES-128-GCM", "SHA256"},
	0x1302: {0x1302, "TLS_AES_256_GCM_SHA384", "ANY", "AES-256-GCM", "SHA384"},
	0x1303: {0x1303, "TLS_CHACHA20_POLY1305_SHA256", "ANY", "CHACHA20-POLY1305", "SHA256"},
	0xCCA8: {0xCCA8, "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256", "ECDHE-RSA", "CHACHA20-POLY1305", "SHA256"},
}

// LookupCipherSuite returns info about a cipher suite by ID.
func LookupCipherSuite(id uint16) (CipherSuite, bool) {
	cs, ok := CommonCipherSuites[id]
	return cs, ok
}

// TLSHandshakeType represents a TLS handshake message type.
type TLSHandshakeType uint8

const (
	HandshakeClientHello     TLSHandshakeType = 1
	HandshakeServerHello     TLSHandshakeType = 2
	HandshakeCertificate     TLSHandshakeType = 11
	HandshakeServerKeyExch   TLSHandshakeType = 12
	HandshakeServerHelloDone TLSHandshakeType = 14
	HandshakeClientKeyExch   TLSHandshakeType = 16
	HandshakeFinished        TLSHandshakeType = 20
)

func (t TLSHandshakeType) String() string {
	names := map[TLSHandshakeType]string{
		1: "ClientHello", 2: "ServerHello", 11: "Certificate",
		12: "ServerKeyExchange", 14: "ServerHelloDone",
		16: "ClientKeyExchange", 20: "Finished",
	}
	if n, ok := names[t]; ok {
		return n
	}
	return fmt.Sprintf("Unknown(%d)", t)
}

// ParseTLSRecord parses a TLS record header.
func ParseTLSRecord(data []byte) (map[string]any, error) {
	if len(data) < 5 {
		return nil, fmt.Errorf("TLS record too short: need 5 bytes, got %d", len(data))
	}
	contentType := data[0]
	version := TLSVersion(uint16(data[1])<<8 | uint16(data[2]))
	length := uint16(data[3])<<8 | uint16(data[4])

	ctName := "unknown"
	switch contentType {
	case 20:
		ctName = "ChangeCipherSpec"
	case 21:
		ctName = "Alert"
	case 22:
		ctName = "Handshake"
	case 23:
		ctName = "ApplicationData"
	}

	result := map[string]any{
		"content_type":      contentType,
		"content_type_name": ctName,
		"version":           version.String(),
		"length":            length,
	}

	// Parse handshake if applicable
	if contentType == 22 && len(data) > 5 {
		hs := parseHandshake(data[5:], version)
		result["handshake"] = hs
	}

	return result, nil
}

func parseHandshake(data []byte, version TLSVersion) map[string]any {
	if len(data) < 4 {
		return nil
	}
	hsType := TLSHandshakeType(data[0])
	hsLen := int(data[1])<<16 | int(data[2])<<8 | int(data[3])

	result := map[string]any{
		"type":   hsType.String(),
		"length": hsLen,
	}

	if hsType == HandshakeClientHello && len(data) > 38 {
		result["client_version"] = TLSVersion(uint16(data[4])<<8 | uint16(data[5])).String()
		result["random"] = fmt.Sprintf("%x", data[6:38])
	}

	return result
}

// --- JWT Support ---

// JWTHeader represents a JWT header.
type JWTHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

// JWTPayload represents decoded JWT claims.
type JWTPayload map[string]any

// ParseJWT parses a JWT token without verification.
func ParseJWT(token string) (*JWTHeader, JWTPayload, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, nil, fmt.Errorf("invalid JWT: expected 3 parts, got %d", len(parts))
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, nil, fmt.Errorf("decode header: %w", err)
	}

	var header JWTHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, nil, fmt.Errorf("parse header: %w", err)
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, nil, fmt.Errorf("decode payload: %w", err)
	}

	var payload JWTPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, nil, fmt.Errorf("parse payload: %w", err)
	}

	return &header, payload, nil
}

// VerifyHMAC verifies an HMAC signature.
func VerifyHMAC(message, signature, key []byte) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	expected := mac.Sum(nil)
	return hmac.Equal(expected, signature)
}

// SSLKeyLog represents a parsed SSLKEYLOGFILE entry.
type SSLKeyLog struct {
	Label        string
	ClientRandom string
	Secret       string
}

// ParseSSLKeyLog parses SSLKEYLOGFILE format entries.
func ParseSSLKeyLog(content string) []SSLKeyLog {
	var entries []SSLKeyLog
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) == 3 {
			entries = append(entries, SSLKeyLog{
				Label:        parts[0],
				ClientRandom: parts[1],
				Secret:       parts[2],
			})
		}
	}
	return entries
}
