// Package diffpcap provides PCAP file comparison and protocol audit trail.
package diffpcap

import (
	"fmt"
	"strings"
)

// PacketDiff represents the difference between two packets.
type PacketDiff struct {
	Index   int            `json:"index"`
	FieldA  map[string]any `json:"field_a,omitempty"`
	FieldB  map[string]any `json:"field_b,omitempty"`
	Changes []FieldChange  `json:"changes"`
}

// FieldChange represents a single field-level change.
type FieldChange struct {
	Field    string `json:"field"`
	OldValue any    `json:"old_value"`
	NewValue any    `json:"new_value"`
	Type     string `json:"type"` // "modified", "added", "removed"
}

// DiffResult holds the result of comparing two packet sets.
type DiffResult struct {
	TotalA    int          `json:"total_a"`
	TotalB    int          `json:"total_b"`
	Identical int          `json:"identical"`
	Modified  int          `json:"modified"`
	OnlyInA   int          `json:"only_in_a"`
	OnlyInB   int          `json:"only_in_b"`
	Diffs     []PacketDiff `json:"diffs,omitempty"`
}

// ComparePackets compares two sets of decoded packets.
func ComparePackets(a, b []map[string]any) *DiffResult {
	result := &DiffResult{
		TotalA: len(a),
		TotalB: len(b),
	}

	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}

	for i := 0; i < maxLen; i++ {
		if i >= len(a) {
			result.OnlyInB++
			continue
		}
		if i >= len(b) {
			result.OnlyInA++
			continue
		}

		changes := compareFields(a[i], b[i])
		if len(changes) == 0 {
			result.Identical++
		} else {
			result.Modified++
			result.Diffs = append(result.Diffs, PacketDiff{
				Index:   i,
				FieldA:  a[i],
				FieldB:  b[i],
				Changes: changes,
			})
		}
	}
	return result
}

func compareFields(a, b map[string]any) []FieldChange {
	var changes []FieldChange

	for k, va := range a {
		vb, ok := b[k]
		if !ok {
			changes = append(changes, FieldChange{Field: k, OldValue: va, Type: "removed"})
			continue
		}
		if fmt.Sprintf("%v", va) != fmt.Sprintf("%v", vb) {
			changes = append(changes, FieldChange{Field: k, OldValue: va, NewValue: vb, Type: "modified"})
		}
	}
	for k, vb := range b {
		if _, ok := a[k]; !ok {
			changes = append(changes, FieldChange{Field: k, NewValue: vb, Type: "added"})
		}
	}
	return changes
}

// AuditEntry represents a protocol change audit entry.
type AuditEntry struct {
	Timestamp string `json:"timestamp"`
	User      string `json:"user"`
	Protocol  string `json:"protocol"`
	Action    string `json:"action"` // "created", "modified", "deleted"
	Details   string `json:"details"`
}

// FormatDiffResult formats a DiffResult as a human-readable string.
func FormatDiffResult(r *DiffResult) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Packets: A=%d, B=%d\n", r.TotalA, r.TotalB))
	b.WriteString(fmt.Sprintf("Identical: %d, Modified: %d\n", r.Identical, r.Modified))
	b.WriteString(fmt.Sprintf("Only in A: %d, Only in B: %d\n", r.OnlyInA, r.OnlyInB))

	for _, d := range r.Diffs {
		b.WriteString(fmt.Sprintf("\nPacket #%d:\n", d.Index))
		for _, c := range d.Changes {
			switch c.Type {
			case "modified":
				b.WriteString(fmt.Sprintf("  ~ %s: %v → %v\n", c.Field, c.OldValue, c.NewValue))
			case "added":
				b.WriteString(fmt.Sprintf("  + %s: %v\n", c.Field, c.NewValue))
			case "removed":
				b.WriteString(fmt.Sprintf("  - %s: %v\n", c.Field, c.OldValue))
			}
		}
	}
	return b.String()
}
