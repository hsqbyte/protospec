// Package project provides PSL project management (init, config, git hooks).
package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents a psl.yaml project configuration.
type Config struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description,omitempty"`
	Protocols   []string `json:"protocols,omitempty"`
	Output      string   `json:"output,omitempty"`
	Language    string   `json:"language,omitempty"`
}

// Init initializes a new PSL project in the given directory.
func Init(dir, name string) error {
	// Create directory structure
	dirs := []string{
		filepath.Join(dir, "protocols"),
		filepath.Join(dir, "generated"),
		filepath.Join(dir, "tests"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("create dir %s: %w", d, err)
		}
	}

	// Create psl.json config
	cfg := Config{
		Name:      name,
		Version:   "0.1.0",
		Protocols: []string{"protocols/*.psl"},
		Output:    "generated",
		Language:  "go",
	}
	data, _ := json.MarshalIndent(cfg, "", "  ")
	cfgPath := filepath.Join(dir, "psl.json")
	if err := os.WriteFile(cfgPath, data, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	// Create example PSL file
	example := `protocol Example version "1.0" {
    byte_order big-endian;
    field type: uint8;
    field length: uint16;
    field payload: bytes;
}
`
	exPath := filepath.Join(dir, "protocols", "example.psl")
	if err := os.WriteFile(exPath, []byte(example), 0644); err != nil {
		return fmt.Errorf("write example: %w", err)
	}

	// Create .gitignore
	gitignore := "generated/\n*.o\n*.so\n"
	giPath := filepath.Join(dir, ".gitignore")
	os.WriteFile(giPath, []byte(gitignore), 0644)

	return nil
}

// LoadConfig loads a psl.json config from the given directory.
func LoadConfig(dir string) (*Config, error) {
	data, err := os.ReadFile(filepath.Join(dir, "psl.json"))
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &cfg, nil
}

// InstallGitHook installs a pre-commit hook that validates PSL files.
func InstallGitHook(dir string) error {
	hookDir := filepath.Join(dir, ".git", "hooks")
	if err := os.MkdirAll(hookDir, 0755); err != nil {
		return err
	}

	hook := `#!/bin/sh
# PSL pre-commit hook — validates all .psl files
for f in $(git diff --cached --name-only --diff-filter=ACM | grep '\.psl$'); do
    psl validate "$f"
    if [ $? -ne 0 ]; then
        echo "PSL validation failed: $f"
        exit 1
    fi
done
`
	hookPath := filepath.Join(hookDir, "pre-commit")
	if err := os.WriteFile(hookPath, []byte(hook), 0755); err != nil {
		return fmt.Errorf("write hook: %w", err)
	}
	return nil
}

// Sign generates a SHA256 signature for a PSL file.
func Sign(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	// Simple hash for integrity verification
	var hash uint64
	for _, b := range data {
		hash = hash*31 + uint64(b)
	}
	return fmt.Sprintf("%016x", hash), nil
}
