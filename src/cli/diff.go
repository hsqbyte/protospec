package cli

import (
	"encoding/json"
	"fmt"

	"github.com/hsqbyte/protospec/src/core/schema"
)

func runDiff(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl diff <proto1.psl> <proto2.psl> [--format json]")
	}

	var file1, file2, format string
	format = "text"

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--format":
			i++
			if i < len(args) {
				format = args[i]
			}
		default:
			if file1 == "" {
				file1 = args[i]
			} else {
				file2 = args[i]
			}
		}
	}

	if file2 == "" {
		return fmt.Errorf("need two PSL files or protocol names to compare")
	}

	// Try loading as registered protocols first, then as files
	s1, err := loadSchema(ctx, file1)
	if err != nil {
		return fmt.Errorf("load %s: %w", file1, err)
	}
	s2, err := loadSchema(ctx, file2)
	if err != nil {
		return fmt.Errorf("load %s: %w", file2, err)
	}

	diffs := compareSchemas(s1, s2)

	switch format {
	case "json":
		data, _ := json.MarshalIndent(diffs, "", "  ")
		fmt.Println(string(data))
	default:
		printDiffs(s1.Name, s2.Name, diffs)
	}
	return nil
}

func loadSchema(ctx *Context, nameOrFile string) (*schema.ProtocolSchema, error) {
	// Try as registered protocol
	s, err := ctx.Lib.Registry().GetSchema(nameOrFile)
	if err == nil {
		return s, nil
	}
	// Try loading from file
	if err := ctx.Lib.LoadPSL(nameOrFile); err != nil {
		return nil, err
	}
	// Get the name from the file (last registered)
	names := ctx.Lib.AllNames()
	if len(names) > 0 {
		return ctx.Lib.Registry().GetSchema(names[len(names)-1])
	}
	return nil, fmt.Errorf("could not load schema from %s", nameOrFile)
}

// FieldDiff describes a difference between two protocol versions.
type FieldDiff struct {
	Type     string `json:"type"` // "added", "removed", "changed"
	Field    string `json:"field"`
	OldValue string `json:"old_value,omitempty"`
	NewValue string `json:"new_value,omitempty"`
	Detail   string `json:"detail,omitempty"`
}

func compareSchemas(s1, s2 *schema.ProtocolSchema) []FieldDiff {
	var diffs []FieldDiff

	// Build field maps
	fields1 := flattenFields(s1.Fields)
	fields2 := flattenFields(s2.Fields)

	map1 := make(map[string]schema.FieldDef)
	map2 := make(map[string]schema.FieldDef)
	for _, f := range fields1 {
		map1[f.Name] = f
	}
	for _, f := range fields2 {
		map2[f.Name] = f
	}

	// Check removed fields
	for _, f := range fields1 {
		if _, ok := map2[f.Name]; !ok {
			diffs = append(diffs, FieldDiff{
				Type:  "removed",
				Field: f.Name,
			})
		}
	}

	// Check added fields
	for _, f := range fields2 {
		if _, ok := map1[f.Name]; !ok {
			diffs = append(diffs, FieldDiff{
				Type:  "added",
				Field: f.Name,
			})
		}
	}

	// Check changed fields
	for _, f1 := range fields1 {
		f2, ok := map2[f1.Name]
		if !ok {
			continue
		}
		if f1.Type != f2.Type {
			diffs = append(diffs, FieldDiff{
				Type:     "changed",
				Field:    f1.Name,
				Detail:   "type",
				OldValue: f1.Type.String(),
				NewValue: f2.Type.String(),
			})
		}
		if f1.BitWidth != f2.BitWidth {
			diffs = append(diffs, FieldDiff{
				Type:     "changed",
				Field:    f1.Name,
				Detail:   "bit_width",
				OldValue: fmt.Sprintf("%d", f1.BitWidth),
				NewValue: fmt.Sprintf("%d", f2.BitWidth),
			})
		}
	}

	// Check byte order
	if s1.DefaultByteOrder != s2.DefaultByteOrder {
		diffs = append(diffs, FieldDiff{
			Type:     "changed",
			Field:    "(byte_order)",
			OldValue: s1.DefaultByteOrder.String(),
			NewValue: s2.DefaultByteOrder.String(),
		})
	}

	return diffs
}

func flattenFields(fields []schema.FieldDef) []schema.FieldDef {
	var result []schema.FieldDef
	for _, f := range fields {
		if f.IsBitfieldGroup {
			result = append(result, f.BitfieldFields...)
		} else {
			result = append(result, f)
		}
	}
	return result
}

func printDiffs(name1, name2 string, diffs []FieldDiff) {
	if len(diffs) == 0 {
		fmt.Printf("no differences between %s and %s\n", name1, name2)
		return
	}

	fmt.Printf("%sDiff: %s ↔ %s%s\n\n", cBold, name1, name2, cReset)
	for _, d := range diffs {
		switch d.Type {
		case "added":
			fmt.Printf("  %s+ %s%s (added)\n", cGreen, d.Field, cReset)
		case "removed":
			fmt.Printf("  %s- %s%s (removed)\n", cYellow, d.Field, cReset)
		case "changed":
			detail := d.Detail
			if detail == "" {
				detail = "value"
			}
			fmt.Printf("  %s~ %s%s %s: %s → %s\n", cCyan, d.Field, cReset, detail, d.OldValue, d.NewValue)
		}
	}

	// Compat summary
	hasBreaking := false
	for _, d := range diffs {
		if d.Type == "removed" || (d.Type == "changed" && (d.Detail == "type" || d.Detail == "bit_width")) {
			hasBreaking = true
			break
		}
	}
	fmt.Println()
	if hasBreaking {
		fmt.Printf("  %s⚠ Breaking changes detected — suggest major version bump%s\n", cYellow, cReset)
	} else {
		addedOnly := true
		for _, d := range diffs {
			if d.Type != "added" {
				addedOnly = false
				break
			}
		}
		if addedOnly {
			fmt.Printf("  %s✓ Backward compatible — suggest minor version bump%s\n", cGreen, cReset)
		}
	}
}

func runCompat(ctx *Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: psl compat <old.psl> <new.psl>")
	}
	// Reuse diff with compat analysis
	return runDiff(ctx, append(args, "--format", "text"))
}
