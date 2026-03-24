package cli

import (
	"encoding/json"
	"fmt"

	"github.com/hsqbyte/protospec/src/schema"
)

// AuditFinding represents a security finding.
type AuditFinding struct {
	Severity string `json:"severity"` // "high", "medium", "low", "info"
	Field    string `json:"field"`
	Rule     string `json:"rule"`
	Message  string `json:"message"`
}

func runAudit(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl audit <protocol> [--format json]")
	}

	name := args[0]
	format := "text"
	for i := 1; i < len(args); i++ {
		if args[i] == "--format" && i+1 < len(args) {
			format = args[i+1]
			i++
		}
	}

	s, err := ctx.Lib.Registry().GetSchema(name)
	if err != nil {
		return err
	}

	findings := auditProtocol(s)

	switch format {
	case "json":
		data, _ := json.MarshalIndent(findings, "", "  ")
		fmt.Println(string(data))
	default:
		printAuditFindings(name, findings)
	}
	return nil
}

func auditProtocol(s *schema.ProtocolSchema) []AuditFinding {
	var findings []AuditFinding

	var fields []schema.FieldDef
	for _, f := range s.Fields {
		if f.IsBitfieldGroup {
			fields = append(fields, f.BitfieldFields...)
		} else {
			fields = append(fields, f)
		}
	}

	// Build field name set for reference checking
	fieldNames := make(map[string]bool)
	for _, f := range fields {
		fieldNames[f.Name] = true
	}

	hasChecksum := false
	for _, f := range fields {
		// Rule 1: Length fields without range constraints
		if f.LengthRef == nil && isLikelyLengthField(f) && f.RangeMin == nil {
			findings = append(findings, AuditFinding{
				Severity: "medium",
				Field:    f.Name,
				Rule:     "unconstrained-length",
				Message:  "Length-like field without range constraint — potential integer overflow",
			})
		}

		// Rule 2: Variable-length fields with unconstrained length_ref
		if f.LengthRef != nil {
			refField := findField(fields, f.LengthRef.FieldName)
			if refField != nil && refField.RangeMin == nil && refField.RangeMax == nil {
				findings = append(findings, AuditFinding{
					Severity: "high",
					Field:    f.Name,
					Rule:     "unbounded-variable-length",
					Message:  fmt.Sprintf("Variable-length field references %q which has no range constraint — potential buffer overflow", f.LengthRef.FieldName),
				})
			}
		}

		// Rule 3: Checksum coverage
		if f.Checksum != nil {
			hasChecksum = true
			if len(f.Checksum.CoverFields) == 0 {
				findings = append(findings, AuditFinding{
					Severity: "high",
					Field:    f.Name,
					Rule:     "empty-checksum-coverage",
					Message:  "Checksum field covers no fields",
				})
			}
		}

		// Rule 4: Large integer fields without validation
		if (f.Type == schema.Uint || f.Type == schema.Int) && f.BitWidth >= 32 && f.RangeMin == nil && f.EnumMap == nil {
			findings = append(findings, AuditFinding{
				Severity: "low",
				Field:    f.Name,
				Rule:     "unconstrained-large-int",
				Message:  fmt.Sprintf("%d-bit integer without range or enum constraint", f.BitWidth),
			})
		}
	}

	// Rule 5: No checksum at all
	if !hasChecksum && len(fields) > 4 {
		findings = append(findings, AuditFinding{
			Severity: "info",
			Field:    "(protocol)",
			Rule:     "no-checksum",
			Message:  "Protocol has no checksum field — data integrity not verified",
		})
	}

	return findings
}

func isLikelyLengthField(f schema.FieldDef) bool {
	name := f.Name
	return (f.Type == schema.Uint || f.Type == schema.Int) &&
		(contains(name, "length") || contains(name, "len") || contains(name, "size") || contains(name, "count"))
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func findField(fields []schema.FieldDef, name string) *schema.FieldDef {
	for i := range fields {
		if fields[i].Name == name {
			return &fields[i]
		}
	}
	return nil
}

func printAuditFindings(name string, findings []AuditFinding) {
	if len(findings) == 0 {
		fmt.Printf("%s✓ %s: no issues found%s\n", cGreen, name, cReset)
		return
	}

	fmt.Printf("%sAudit: %s%s (%d findings)\n\n", cBold, name, cReset, len(findings))

	sevColor := map[string]string{
		"high":   "\033[31m", // red
		"medium": cYellow,
		"low":    cCyan,
		"info":   cDim,
	}

	for _, f := range findings {
		color := sevColor[f.Severity]
		if color == "" {
			color = cDim
		}
		fmt.Printf("  %s[%s]%s %s — %s\n", color, f.Severity, cReset, f.Field, f.Message)
		fmt.Printf("         %srule: %s%s\n", cDim, f.Rule, cReset)
	}
}
