// Package mobile provides mobile platform protocol analysis support.
package mobile

import (
	"fmt"
	"strings"
)

// Platform represents a mobile platform.
type Platform string

const (
	PlatformIOS     Platform = "ios"
	PlatformAndroid Platform = "android"
)

// SDKConfig represents a mobile SDK configuration.
type SDKConfig struct {
	Platform Platform `json:"platform"`
	Language string   `json:"language"` // swift, kotlin
	Features []string `json:"features"`
}

// NewSDKConfig creates a mobile SDK config.
func NewSDKConfig(platform Platform) *SDKConfig {
	lang := "kotlin"
	if platform == PlatformIOS {
		lang = "swift"
	}
	return &SDKConfig{
		Platform: platform,
		Language: lang,
		Features: []string{"protocol-viewer", "pcap-analysis", "bluetooth-analysis", "wifi-analysis"},
	}
}

// GenerateSwiftSDK generates Swift SDK skeleton.
func GenerateSwiftSDK() string {
	return `import Foundation

public class PSLMobile {
    public static let shared = PSLMobile()
    
    public func decode(protocol: String, data: Data) -> [String: Any]? {
        // Protocol decoding via PSL engine
        return nil
    }
    
    public func analyzeBluetooth(data: Data) -> [String: Any]? {
        return decode(protocol: "Bluetooth_HCI", data: data)
    }
}
`
}

// GenerateKotlinSDK generates Kotlin SDK skeleton.
func GenerateKotlinSDK() string {
	return `package dev.psl.mobile

class PSLMobile {
    companion object {
        val instance = PSLMobile()
    }
    
    fun decode(protocol: String, data: ByteArray): Map<String, Any>? {
        // Protocol decoding via PSL engine
        return null
    }
    
    fun analyzeWifi(data: ByteArray): Map<String, Any>? {
        return decode("IEEE80211", data)
    }
}
`
}

// Describe returns SDK description.
func (c *SDKConfig) Describe() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Mobile SDK: %s (%s)\n", c.Platform, c.Language))
	b.WriteString("Features:\n")
	for _, f := range c.Features {
		b.WriteString(fmt.Sprintf("  • %s\n", f))
	}
	return b.String()
}
