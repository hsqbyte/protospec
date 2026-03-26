package format

import (
	"fmt"
	"math/big"
	"net"
	"strings"
)

// IPv4Formatter converts between a 32-bit unsigned integer and a dotted-decimal
// string such as "192.168.1.1".
type IPv4Formatter struct{}

// Encode converts a numeric value (uint32, uint64, int, int64, etc.) to a
// dotted-decimal IPv4 string.
func (f *IPv4Formatter) Encode(value any) (string, error) {
	v, err := toUint64(value)
	if err != nil {
		return "", fmt.Errorf("ipv4 encode: %w", err)
	}
	if v > 0xFFFFFFFF {
		return "", fmt.Errorf("ipv4 encode: value %d exceeds 32-bit range", v)
	}
	u := uint32(v)
	return fmt.Sprintf("%d.%d.%d.%d",
		(u>>24)&0xFF,
		(u>>16)&0xFF,
		(u>>8)&0xFF,
		u&0xFF,
	), nil
}

// Decode converts a dotted-decimal IPv4 string to a uint64 value.
func (f *IPv4Formatter) Decode(display string) (any, error) {
	ip := net.ParseIP(strings.TrimSpace(display))
	if ip == nil {
		return nil, fmt.Errorf("ipv4 decode: invalid address %q", display)
	}
	ip4 := ip.To4()
	if ip4 == nil {
		return nil, fmt.Errorf("ipv4 decode: not an IPv4 address %q", display)
	}
	val := uint64(ip4[0])<<24 | uint64(ip4[1])<<16 | uint64(ip4[2])<<8 | uint64(ip4[3])
	return val, nil
}

// MACFormatter converts between a 48-bit integer and a colon-separated hex
// string such as "aa:bb:cc:dd:ee:ff".
type MACFormatter struct{}

// Encode converts a numeric value (48-bit range) to a colon-separated hex MAC
// string.
func (f *MACFormatter) Encode(value any) (string, error) {
	v, err := toUint64(value)
	if err != nil {
		return "", fmt.Errorf("mac encode: %w", err)
	}
	if v > 0xFFFFFFFFFFFF {
		return "", fmt.Errorf("mac encode: value %d exceeds 48-bit range", v)
	}
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
		(v>>40)&0xFF,
		(v>>32)&0xFF,
		(v>>24)&0xFF,
		(v>>16)&0xFF,
		(v>>8)&0xFF,
		v&0xFF,
	), nil
}

// Decode converts a colon-separated hex MAC string to a uint64 value.
func (f *MACFormatter) Decode(display string) (any, error) {
	hw, err := net.ParseMAC(strings.TrimSpace(display))
	if err != nil {
		return nil, fmt.Errorf("mac decode: %w", err)
	}
	if len(hw) != 6 {
		return nil, fmt.Errorf("mac decode: expected 6 bytes, got %d", len(hw))
	}
	val := uint64(hw[0])<<40 | uint64(hw[1])<<32 | uint64(hw[2])<<24 |
		uint64(hw[3])<<16 | uint64(hw[4])<<8 | uint64(hw[5])
	return val, nil
}

// IPv6Formatter converts between a 128-bit value and a colon-separated hex
// group string such as "2001:0db8:0000:0000:0000:0000:0000:0001".
//
// The Encode method accepts [16]byte, []byte (length 16), *big.Int, or
// big.Int. The Decode method returns a *big.Int.
type IPv6Formatter struct{}

// Encode converts a 128-bit value to a colon-separated hex group string.
func (f *IPv6Formatter) Encode(value any) (string, error) {
	var b [16]byte

	switch v := value.(type) {
	case [16]byte:
		b = v
	case []byte:
		if len(v) != 16 {
			return "", fmt.Errorf("ipv6 encode: expected 16 bytes, got %d", len(v))
		}
		copy(b[:], v)
	case *big.Int:
		if v.Sign() < 0 {
			return "", fmt.Errorf("ipv6 encode: negative big.Int not allowed")
		}
		buf := v.Bytes()
		if len(buf) > 16 {
			return "", fmt.Errorf("ipv6 encode: big.Int exceeds 128 bits")
		}
		// Right-align into 16 bytes.
		copy(b[16-len(buf):], buf)
	case big.Int:
		return f.Encode(&v)
	default:
		return "", fmt.Errorf("ipv6 encode: unsupported type %T", value)
	}

	groups := make([]string, 8)
	for i := 0; i < 8; i++ {
		hi := b[i*2]
		lo := b[i*2+1]
		groups[i] = fmt.Sprintf("%02x%02x", hi, lo)
	}
	return strings.Join(groups, ":"), nil
}

// Decode converts a colon-separated hex group string to a *big.Int.
func (f *IPv6Formatter) Decode(display string) (any, error) {
	ip := net.ParseIP(strings.TrimSpace(display))
	if ip == nil {
		return nil, fmt.Errorf("ipv6 decode: invalid address %q", display)
	}
	ip6 := ip.To16()
	if ip6 == nil {
		return nil, fmt.Errorf("ipv6 decode: not an IPv6 address %q", display)
	}
	val := new(big.Int).SetBytes([]byte(ip6))
	return val, nil
}

// toUint64 converts various numeric types to uint64.
func toUint64(v any) (uint64, error) {
	switch n := v.(type) {
	case uint64:
		return n, nil
	case uint32:
		return uint64(n), nil
	case uint16:
		return uint64(n), nil
	case uint8:
		return uint64(n), nil
	case uint:
		return uint64(n), nil
	case int:
		if n < 0 {
			return 0, fmt.Errorf("negative value %d", n)
		}
		return uint64(n), nil
	case int64:
		if n < 0 {
			return 0, fmt.Errorf("negative value %d", n)
		}
		return uint64(n), nil
	case int32:
		if n < 0 {
			return 0, fmt.Errorf("negative value %d", n)
		}
		return uint64(n), nil
	case int16:
		if n < 0 {
			return 0, fmt.Errorf("negative value %d", n)
		}
		return uint64(n), nil
	case int8:
		if n < 0 {
			return 0, fmt.Errorf("negative value %d", n)
		}
		return uint64(n), nil
	default:
		return 0, fmt.Errorf("unsupported numeric type %T", v)
	}
}

// NewDefaultFormatRegistry creates a FormatRegistry pre-loaded with the
// built-in formatters: "ipv4", "mac", and "ipv6".
func NewDefaultFormatRegistry() *FormatRegistry {
	r := NewFormatRegistry()
	r.Register("ipv4", &IPv4Formatter{})
	r.Register("mac", &MACFormatter{})
	r.Register("ipv6", &IPv6Formatter{})
	return r
}
