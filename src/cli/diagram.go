package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/hsqbyte/protospec/src/core/schema"
)

func runDiagram(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl diagram <protocol> [--format ascii|svg]")
	}

	name := args[0]
	format := "ascii"
	outFile := ""

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--format":
			i++
			if i < len(args) {
				format = args[i]
			}
		case "-o":
			i++
			if i < len(args) {
				outFile = args[i]
			}
		}
	}

	s, err := ctx.Lib.Registry().GetSchema(name)
	if err != nil {
		// Try message protocol fallback
		ms := ctx.Lib.Message(name)
		if ms != nil {
			output := generateMessageDiagram(ms)
			if outFile != "" {
				return os.WriteFile(outFile, []byte(output), 0o644)
			}
			fmt.Print(output)
			return nil
		}
		return err
	}

	var output string
	switch format {
	case "ascii":
		output = generateASCIIDiagram(s)
	case "svg":
		output = generateSVGDiagram(s)
	default:
		return fmt.Errorf("unsupported format: %s (supported: ascii, svg)", format)
	}

	if outFile != "" {
		return os.WriteFile(outFile, []byte(output), 0o644)
	}
	fmt.Print(output)
	return nil
}

// generateASCIIDiagram generates an RFC-style ASCII protocol header diagram.
func generateASCIIDiagram(s *schema.ProtocolSchema) string {
	var fields []schema.FieldDef
	for _, f := range s.Fields {
		if f.IsBitfieldGroup {
			fields = append(fields, f.BitfieldFields...)
		} else {
			fields = append(fields, f)
		}
	}
	return generateASCIIDiagramSimple(s, fields)
}

func generateASCIIDiagramSimple(s *schema.ProtocolSchema, fields []schema.FieldDef) string {
	var b strings.Builder
	rowWidth := 32 // bits per row

	b.WriteString(fmt.Sprintf("  %s (v%s)\n", s.Name, s.Version))
	b.WriteString("  Byte order: " + s.DefaultByteOrder.String() + "\n\n")

	// Bit ruler
	b.WriteString("   0                   1                   2                   3\n")
	b.WriteString("   0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1\n")

	bitPos := 0
	type cell struct {
		name  string
		start int
		bits  int
	}

	// Collect cells per row
	var rows [][]cell
	var currentRow []cell

	for _, f := range fields {
		bits := f.BitWidth
		if bits == 0 {
			if f.FixedLength > 0 {
				bits = f.FixedLength * 8
			} else {
				// Variable: show as one row
				bits = rowWidth - (bitPos % rowWidth)
				if bits == 0 || bits == rowWidth {
					bits = rowWidth
				}
			}
		}

		remaining := bits
		for remaining > 0 {
			colStart := bitPos % rowWidth
			available := rowWidth - colStart
			take := remaining
			if take > available {
				take = available
			}

			currentRow = append(currentRow, cell{name: f.Name, start: colStart, bits: take})
			bitPos += take
			remaining -= take

			if bitPos%rowWidth == 0 {
				rows = append(rows, currentRow)
				currentRow = nil
			}
		}
	}
	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}

	// Render rows
	for _, row := range rows {
		// Top border
		b.WriteString("  +")
		for _, c := range row {
			_ = c.start
			w := c.bits*2 - 1
			b.WriteString(strings.Repeat("-", w) + "+")
		}
		// Fill remaining if row doesn't reach 32 bits
		totalBits := 0
		for _, c := range row {
			totalBits += c.bits
		}
		if totalBits < rowWidth {
			rem := rowWidth - totalBits
			b.WriteString(strings.Repeat("-", rem*2-1) + "+")
		}
		b.WriteString("\n")

		// Content
		b.WriteString("  |")
		for _, c := range row {
			w := c.bits*2 - 1
			label := c.name
			if len(label) > w {
				label = label[:w]
			}
			pad := w - len(label)
			left := pad / 2
			right := pad - left
			b.WriteString(strings.Repeat(" ", left) + label + strings.Repeat(" ", right) + "|")
		}
		if totalBits < rowWidth {
			rem := rowWidth - totalBits
			w := rem*2 - 1
			b.WriteString(strings.Repeat(" ", w) + "|")
		}
		b.WriteString("\n")
	}

	// Bottom border
	b.WriteString("  +")
	b.WriteString(strings.Repeat("-+", rowWidth))
	b.WriteString("\n")

	return b.String()
}

func generateSVGDiagram(s *schema.ProtocolSchema) string {
	var b strings.Builder
	rowWidth := 32
	cellW := 20 // pixels per bit
	cellH := 30
	marginX := 10
	marginY := 50

	var fields []schema.FieldDef
	for _, f := range s.Fields {
		if f.IsBitfieldGroup {
			fields = append(fields, f.BitfieldFields...)
		} else {
			fields = append(fields, f)
		}
	}

	// Calculate rows
	type cell struct {
		name  string
		start int
		bits  int
	}
	var rows [][]cell
	var currentRow []cell
	bitPos := 0

	for _, f := range fields {
		bits := f.BitWidth
		if bits == 0 {
			if f.FixedLength > 0 {
				bits = f.FixedLength * 8
			} else {
				bits = rowWidth - (bitPos % rowWidth)
				if bits == 0 || bits == rowWidth {
					bits = rowWidth
				}
			}
		}
		remaining := bits
		for remaining > 0 {
			colStart := bitPos % rowWidth
			available := rowWidth - colStart
			take := remaining
			if take > available {
				take = available
			}
			currentRow = append(currentRow, cell{name: f.Name, start: colStart, bits: take})
			bitPos += take
			remaining -= take
			if bitPos%rowWidth == 0 {
				rows = append(rows, currentRow)
				currentRow = nil
			}
		}
	}
	if len(currentRow) > 0 {
		rows = append(rows, currentRow)
	}

	totalW := rowWidth*cellW + marginX*2
	totalH := len(rows)*cellH + marginY + 20

	b.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d">`, totalW, totalH))
	b.WriteString("\n")
	b.WriteString(`<style>
  text { font-family: monospace; font-size: 11px; }
  .title { font-size: 14px; font-weight: bold; }
  .ruler { font-size: 9px; fill: #666; }
  rect { fill: #f8f9fa; stroke: #333; stroke-width: 1; }
</style>`)
	b.WriteString("\n")

	// Title
	b.WriteString(fmt.Sprintf(`<text x="%d" y="20" class="title">%s (v%s)</text>`, marginX, s.Name, s.Version))
	b.WriteString("\n")

	// Bit ruler
	for i := 0; i < rowWidth; i++ {
		x := marginX + i*cellW + cellW/2
		b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" class="ruler" text-anchor="middle">%d</text>`, x, marginY-5, i))
		b.WriteString("\n")
	}

	// Rows
	for ri, row := range rows {
		y := marginY + ri*cellH
		for _, c := range row {
			x := marginX + c.start*cellW
			w := c.bits * cellW
			b.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d"/>`, x, y, w, cellH))
			b.WriteString("\n")
			tx := x + w/2
			ty := y + cellH/2 + 4
			label := c.name
			maxChars := w / 7
			if maxChars < 1 {
				maxChars = 1
			}
			if len(label) > maxChars {
				label = label[:maxChars]
			}
			b.WriteString(fmt.Sprintf(`<text x="%d" y="%d" text-anchor="middle">%s</text>`, tx, ty, label))
			b.WriteString("\n")
		}
	}

	b.WriteString("</svg>\n")
	return b.String()
}

// generateMessageDiagram generates a Mermaid flowchart for a message protocol.
func generateMessageDiagram(ms *schema.MessageSchema) string {
	var b strings.Builder
	b.WriteString("```mermaid\nflowchart TD\n")

	// Transport node
	transportLabel := ms.Transport
	if ms.TransportDef != nil {
		transportLabel = fmt.Sprintf("%s v%s", ms.TransportDef.Name, ms.TransportDef.Version)
	}
	b.WriteString(fmt.Sprintf("    T[\"%s\\n%s\"]\n", ms.Name, transportLabel))

	for i, msg := range ms.Messages {
		msgID := fmt.Sprintf("M%d", i)
		// Build field list
		var fields []string
		for _, f := range msg.Fields {
			fields = append(fields, fmt.Sprintf("%s: %s", f.Name, f.Type))
		}
		fieldStr := ""
		if len(fields) > 0 {
			fieldStr = "\\n" + strings.Join(fields, "\\n")
		}
		b.WriteString(fmt.Sprintf("    %s[\"%s (%s)%s\"]\n", msgID, msg.Name, msg.Kind, fieldStr))
		b.WriteString(fmt.Sprintf("    T --> %s\n", msgID))

		if msg.Response != nil {
			respID := fmt.Sprintf("R%d", i)
			var rFields []string
			for _, f := range msg.Response.Fields {
				rFields = append(rFields, fmt.Sprintf("%s: %s", f.Name, f.Type))
			}
			rFieldStr := ""
			if len(rFields) > 0 {
				rFieldStr = "\\n" + strings.Join(rFields, "\\n")
			}
			b.WriteString(fmt.Sprintf("    %s[\"%s%s\"]\n", respID, msg.Response.Name, rFieldStr))
			b.WriteString(fmt.Sprintf("    %s --> %s\n", msgID, respID))
		}
	}

	b.WriteString("```\n")
	return b.String()
}
