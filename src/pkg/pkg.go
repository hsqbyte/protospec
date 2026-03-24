// Package pkg provides protocol package management.
package pkg

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PackageMeta describes a PSL package.
type PackageMeta struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	License     string   `json:"license"`
	Protocols   []string `json:"protocols"`
}

// Registry represents a package registry.
type Registry struct {
	URL      string
	Packages map[string]PackageMeta
}

// NewRegistry creates a new registry client.
func NewRegistry(url string) *Registry {
	return &Registry{URL: url, Packages: make(map[string]PackageMeta)}
}

// InitPackage initializes a new PSL package in the given directory.
func InitPackage(dir, name string) error {
	meta := PackageMeta{
		Name:    name,
		Version: "0.1.0",
		License: "MIT",
	}
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "psl-package.json"), data, 0644)
}

// Publish publishes a package to the registry.
func (r *Registry) Publish(meta PackageMeta) error {
	r.Packages[meta.Name] = meta
	return nil
}

// Install installs a package by name.
func (r *Registry) Install(name, destDir string) error {
	pkg, ok := r.Packages[name]
	if !ok {
		return fmt.Errorf("package not found: %s", name)
	}
	dir := filepath.Join(destDir, "psl_modules", pkg.Name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, _ := json.MarshalIndent(pkg, "", "  ")
	return os.WriteFile(filepath.Join(dir, "psl-package.json"), data, 0644)
}

// Search searches packages by query.
func (r *Registry) Search(query string) []PackageMeta {
	var results []PackageMeta
	for _, pkg := range r.Packages {
		if strings.Contains(pkg.Name, query) || strings.Contains(pkg.Description, query) {
			results = append(results, pkg)
		}
	}
	return results
}
