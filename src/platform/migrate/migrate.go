// Package migrate provides protocol version migration tools.
package migrate

import (
	"fmt"
	"strings"
)

// MigrationStep represents a single migration operation.
type MigrationStep struct {
	Type        string `json:"type"` // add, remove, rename, change_type
	Field       string `json:"field"`
	NewField    string `json:"new_field,omitempty"`
	NewType     string `json:"new_type,omitempty"`
	Description string `json:"description"`
}

// MigrationPlan describes how to migrate between protocol versions.
type MigrationPlan struct {
	Protocol    string          `json:"protocol"`
	FromVersion string          `json:"from_version"`
	ToVersion   string          `json:"to_version"`
	Steps       []MigrationStep `json:"steps"`
	Reversible  bool            `json:"reversible"`
}

// GeneratePlan generates a migration plan between two protocol versions.
func GeneratePlan(protocol, fromVer, toVer string) *MigrationPlan {
	return &MigrationPlan{
		Protocol:    protocol,
		FromVersion: fromVer,
		ToVersion:   toVer,
		Steps:       []MigrationStep{},
		Reversible:  true,
	}
}

// GenerateScript generates a migration script from a plan.
func GenerateScript(plan *MigrationPlan) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("# Migration: %s %s → %s\n", plan.Protocol, plan.FromVersion, plan.ToVersion))
	for i, step := range plan.Steps {
		b.WriteString(fmt.Sprintf("step %d: %s field '%s'", i+1, step.Type, step.Field))
		if step.NewField != "" {
			b.WriteString(fmt.Sprintf(" → '%s'", step.NewField))
		}
		b.WriteString("\n")
	}
	return b.String()
}

// Validate validates data against the target version after migration.
func Validate(plan *MigrationPlan, data map[string]interface{}) error {
	for _, step := range plan.Steps {
		switch step.Type {
		case "add":
			if _, ok := data[step.Field]; !ok {
				return fmt.Errorf("missing required field after migration: %s", step.Field)
			}
		case "remove":
			if _, ok := data[step.Field]; ok {
				return fmt.Errorf("field should be removed: %s", step.Field)
			}
		}
	}
	return nil
}

// Rollback generates a rollback plan from an existing migration plan.
func Rollback(plan *MigrationPlan) *MigrationPlan {
	if !plan.Reversible {
		return nil
	}
	rollback := &MigrationPlan{
		Protocol:    plan.Protocol,
		FromVersion: plan.ToVersion,
		ToVersion:   plan.FromVersion,
		Reversible:  true,
	}
	for i := len(plan.Steps) - 1; i >= 0; i-- {
		step := plan.Steps[i]
		switch step.Type {
		case "add":
			rollback.Steps = append(rollback.Steps, MigrationStep{Type: "remove", Field: step.Field})
		case "remove":
			rollback.Steps = append(rollback.Steps, MigrationStep{Type: "add", Field: step.Field})
		case "rename":
			rollback.Steps = append(rollback.Steps, MigrationStep{Type: "rename", Field: step.NewField, NewField: step.Field})
		default:
			rollback.Steps = append(rollback.Steps, step)
		}
	}
	return rollback
}
