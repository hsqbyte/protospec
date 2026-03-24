// Package plugin manages PSL protocol packages (install, load, publish).
package plugin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// PluginDir returns the default plugin directory (~/.psl/plugins/).
func PluginDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".psl", "plugins")
}

// PackageManifest describes a PSL protocol package.
type PackageManifest struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      string `json:"author"`
	License     string `json:"license"`
}

// Install downloads and installs a protocol package from a URL or local path.
func Install(source string) error {
	pluginDir := PluginDir()

	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		return installFromURL(source, pluginDir)
	}

	// Local directory install
	return installFromLocal(source, pluginDir)
}

func installFromURL(url, pluginDir string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	// Read manifest to get package name
	tmpDir, err := os.MkdirTemp("", "psl-install-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Expect a tar.gz or directory structure
	// For now, treat as a single PSL file
	name := filepath.Base(url)
	name = strings.TrimSuffix(name, filepath.Ext(name))

	destDir := filepath.Join(pluginDir, name)
	os.MkdirAll(destDir, 0o755)

	return os.WriteFile(filepath.Join(destDir, name+".psl"), data, 0o644)
}

func installFromLocal(source, pluginDir string) error {
	info, err := os.Stat(source)
	if err != nil {
		return err
	}

	var name string
	if info.IsDir() {
		name = filepath.Base(source)
	} else {
		name = strings.TrimSuffix(filepath.Base(source), filepath.Ext(source))
	}

	destDir := filepath.Join(pluginDir, name)
	os.MkdirAll(destDir, 0o755)

	if info.IsDir() {
		return copyDir(source, destDir)
	}
	return copyFile(source, filepath.Join(destDir, filepath.Base(source)))
}

// ListInstalled returns all installed plugin names.
func ListInstalled() ([]string, error) {
	pluginDir := PluginDir()
	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// Uninstall removes an installed plugin.
func Uninstall(name string) error {
	dir := filepath.Join(PluginDir(), name)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("plugin %q not installed", name)
	}
	return os.RemoveAll(dir)
}

// InitPackage creates a new protocol package scaffold.
func InitPackage(dir, name string) error {
	os.MkdirAll(dir, 0o755)

	manifest := PackageManifest{
		Name:    name,
		Version: "0.1.0",
	}
	data, _ := json.MarshalIndent(manifest, "", "  ")
	if err := os.WriteFile(filepath.Join(dir, "package.json"), data, 0o644); err != nil {
		return err
	}

	// Create template PSL file
	psl := fmt.Sprintf(`protocol %s version "0.1.0" {
  byte_order big-endian;

  // Add fields here
}
`, name)
	if err := os.WriteFile(filepath.Join(dir, name+".psl"), []byte(psl), 0o644); err != nil {
		return err
	}

	// Create meta.json
	meta := map[string]any{
		"title":       map[string]string{"en": name + " protocol"},
		"description": map[string]string{"en": ""},
		"type":        "binary",
		"layer":       "",
	}
	metaData, _ := json.MarshalIndent(meta, "", "  ")
	return os.WriteFile(filepath.Join(dir, "meta.json"), metaData, 0o644)
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}

func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		srcPath := filepath.Join(src, e.Name())
		dstPath := filepath.Join(dst, e.Name())
		if e.IsDir() {
			os.MkdirAll(dstPath, 0o755)
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}
