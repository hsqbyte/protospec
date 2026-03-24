package pdl

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// TransportRef represents a parsed transport reference.
type TransportRef struct {
	Name    string
	Version *string // nil means "latest"
}

// ParseTransportRef parses "jsonrpc" or "jsonrpc@2.0" into a TransportRef.
func ParseTransportRef(ref string) TransportRef {
	if idx := strings.Index(ref, "@"); idx >= 0 {
		name := ref[:idx]
		version := ref[idx+1:]
		return TransportRef{Name: name, Version: &version}
	}
	return TransportRef{Name: ref, Version: nil}
}

// ResolveVersion finds the best matching version directory under basePath.
// If requested is non-nil, it looks for an exact match.
// If requested is nil, it returns the latest version by semver ordering.
// If no version subdirectories exist, it returns basePath as-is (flat fallback).
func ResolveVersion(basePath string, requested *string) (string, error) {
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return "", fmt.Errorf("reading directory %s: %w", basePath, err)
	}

	var versions []string
	for _, e := range entries {
		if e.IsDir() && isVersionDir(e.Name()) {
			versions = append(versions, e.Name())
		}
	}

	if len(versions) == 0 {
		// No version subdirectories — flat structure fallback.
		return basePath, nil
	}

	if requested != nil {
		for _, v := range versions {
			if v == *requested {
				return filepath.Join(basePath, v), nil
			}
		}
		return "", fmt.Errorf("version %s not found under %s", *requested, basePath)
	}

	// Sort by semver and return the latest.
	sort.Slice(versions, func(i, j int) bool {
		return compareSemver(versions[i], versions[j]) < 0
	})
	latest := versions[len(versions)-1]
	return filepath.Join(basePath, latest), nil
}

// isVersionDir returns true if the directory name looks like a version number
// (e.g., "2.0", "1.0.1", "3", "2024-11-05").
func isVersionDir(name string) bool {
	if len(name) == 0 {
		return false
	}
	// Must start with a digit.
	if name[0] < '0' || name[0] > '9' {
		return false
	}
	// Allow digits, dots, and hyphens.
	for _, c := range name {
		if (c < '0' || c > '9') && c != '.' && c != '-' {
			return false
		}
	}
	return true
}

// compareSemver compares two version strings by splitting on "." and comparing
// each numeric segment. Returns -1, 0, or 1.
func compareSemver(a, b string) int {
	partsA := strings.Split(a, ".")
	partsB := strings.Split(b, ".")

	maxLen := len(partsA)
	if len(partsB) > maxLen {
		maxLen = len(partsB)
	}

	for i := 0; i < maxLen; i++ {
		var na, nb int
		if i < len(partsA) {
			na, _ = strconv.Atoi(partsA[i])
		}
		if i < len(partsB) {
			nb, _ = strconv.Atoi(partsB[i])
		}
		if na < nb {
			return -1
		}
		if na > nb {
			return 1
		}
	}
	return 0
}
