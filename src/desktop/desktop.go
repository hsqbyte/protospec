// Package desktop provides cross-platform desktop application support.
package desktop

import (
	"encoding/json"
	"fmt"
	"strings"
)

// WindowConfig represents a desktop window configuration.
type WindowConfig struct {
	Title  string `json:"title"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Theme  string `json:"theme"` // light, dark
}

// AppConfig represents the desktop application configuration.
type AppConfig struct {
	Name     string       `json:"name"`
	Version  string       `json:"version"`
	Window   WindowConfig `json:"window"`
	Features []string     `json:"features"`
}

// DefaultConfig returns the default desktop app configuration.
func DefaultConfig() *AppConfig {
	return &AppConfig{
		Name:     "PSL Desktop",
		Version:  "1.0.0",
		Window:   WindowConfig{Title: "PSL Protocol Browser", Width: 1280, Height: 800, Theme: "dark"},
		Features: []string{"protocol-browser", "drag-drop-designer", "integrated-terminal", "repl"},
	}
}

// GenerateWailsConfig generates a Wails project configuration.
func GenerateWailsConfig() string {
	cfg := map[string]any{
		"name":             "psl-desktop",
		"outputfilename":   "psl-desktop",
		"frontend:install": "npm install",
		"frontend:build":   "npm run build",
		"author":           map[string]string{"name": "PSL Team"},
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	return string(data)
}

// GenerateTauriConfig generates a Tauri project configuration.
func GenerateTauriConfig() string {
	cfg := map[string]any{
		"productName": "PSL Desktop",
		"version":     "1.0.0",
		"identifier":  "dev.psl.desktop",
		"build":       map[string]string{"beforeBuildCommand": "npm run build", "frontendDist": "../dist"},
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	return string(data)
}

// Describe returns app description.
func (c *AppConfig) Describe() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("PSL Desktop v%s\n", c.Version))
	b.WriteString(fmt.Sprintf("  Window: %dx%d (%s theme)\n", c.Window.Width, c.Window.Height, c.Window.Theme))
	b.WriteString("  Features:\n")
	for _, f := range c.Features {
		b.WriteString(fmt.Sprintf("    • %s\n", f))
	}
	return b.String()
}
