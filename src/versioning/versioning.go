// Package versioning provides protocol version management and changelog generation.
package versioning

import (
	"fmt"
	"strings"

	"github.com/hsqbyte/protospec/src/schema"
)

// ChangeType represents the type of protocol change.
type ChangeType int

const (
	FieldAdded   ChangeType = iota // minor
	FieldRemoved                   // major
	FieldChanged                   // major
	MetaChanged                    // patch
)

// Change represents a single protocol change.
type Change struct {
	Type        ChangeType `json:"type"`
	Field       string     `json:"field"`
	Description string     `json:"description"`
}

// VersionBump represents a semver bump recommendation.
type VersionBump struct {
	Current   string   `json:"current"`
	Suggested string   `json:"suggested"`
	BumpType  string   `json:"bump_type"` // "major", "minor", "patch"
	Changes   []Change `json:"changes"`
	Breaking  bool     `json:"breaking"`
}

// CompareSchemas compares two protocol schemas and returns changes.
func CompareSchemas(old, new *schema.ProtocolSchema) *VersionBump {
	bump := &VersionBump{Current: old.Version}
	oldFields := fieldMap(old)
	newFields := fieldMap(new)

	// Check removed fields
	for name := range oldFields {
		if _, ok := newFields[name]; !ok {
			bump.Changes = append(bump.Changes, Change{
				Type: FieldRemoved, Field: name,
				Description: fmt.Sprintf("field %q removed", name),
			})
			bump.Breaking = true
		}
	}

	// Check added fields
	for name := range newFields {
		if _, ok := oldFields[name]; !ok {
			bump.Changes = append(bump.Changes, Change{
				Type: FieldAdded, Field: name,
				Description: fmt.Sprintf("field %q added", name),
			})
		}
	}

	// Check modified fields
	for name, of := range oldFields {
		if nf, ok := newFields[name]; ok {
			if of.BitWidth != nf.BitWidth || of.Type != nf.Type {
				bump.Changes = append(bump.Changes, Change{
					Type: FieldChanged, Field: name,
					Description: fmt.Sprintf("field %q type/width changed", name),
				})
				bump.Breaking = true
			}
		}
	}

	// Determine bump type
	if bump.Breaking {
		bump.BumpType = "major"
	} else if len(bump.Changes) > 0 {
		bump.BumpType = "minor"
	} else {
		bump.BumpType = "patch"
	}

	return bump
}

func fieldMap(s *schema.ProtocolSchema) map[string]schema.FieldDef {
	m := make(map[string]schema.FieldDef)
	for _, f := range s.Fields {
		if f.IsBitfieldGroup {
			for _, bf := range f.BitfieldFields {
				m[bf.Name] = bf
			}
		} else {
			m[f.Name] = f
		}
	}
	return m
}

// GenerateChangelog generates a changelog from version bumps.
func GenerateChangelog(bumps []*VersionBump) string {
	var b strings.Builder
	b.WriteString("# Changelog\n\n")
	for _, bump := range bumps {
		b.WriteString(fmt.Sprintf("## %s → %s (%s)\n\n", bump.Current, bump.Suggested, bump.BumpType))
		if bump.Breaking {
			b.WriteString("**BREAKING CHANGES**\n\n")
		}
		for _, c := range bump.Changes {
			prefix := "+"
			if c.Type == FieldRemoved {
				prefix = "-"
			} else if c.Type == FieldChanged {
				prefix = "~"
			}
			b.WriteString(fmt.Sprintf("- %s %s\n", prefix, c.Description))
		}
		b.WriteString("\n")
	}
	return b.String()
}
