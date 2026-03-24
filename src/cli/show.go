package cli

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/hsqbyte/protospec/src/pdl"
	"github.com/hsqbyte/protospec/src/protocol"
	"github.com/hsqbyte/protospec/src/schema"

	"golang.org/x/term"
)

// ANSI color helpers
const (
	cReset  = "\033[0m"
	cBold   = "\033[1m"
	cDim    = "\033[2m"
	cCyan   = "\033[36m"
	cGreen  = "\033[32m"
	cYellow = "\033[33m"
)

// termWidth returns the terminal width, defaulting to 80.
func termWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return 80
	}
	return w
}

func runShow(ctx *Context, name string, raw bool) error {
	// Check if it's a message protocol
	if ms := ctx.Lib.Message(name); ms != nil {
		if raw {
			printer := &pdl.PDLPrinter{}
			fmt.Print(printer.PrintMessage(ms))
			return nil
		}
		meta := ctx.Lib.Meta(name)
		return printMessageTable(ctx, ms, meta)
	}

	s, err := ctx.Lib.Registry().GetSchema(name)
	if err != nil {
		return err
	}
	if raw {
		printer := &pdl.PDLPrinter{}
		fmt.Print(printer.Print(s))
		return nil
	}
	meta := ctx.Lib.Meta(name)
	return printTable(ctx, s, meta)
}

// displayWidth returns the terminal display width of a string.
func displayWidth(s string) int {
	w := 0
	for _, r := range s {
		if isCJK(r) {
			w += 2
		} else {
			w++
		}
	}
	return w
}

func isCJK(r rune) bool {
	return unicode.Is(unicode.Han, r) ||
		unicode.Is(unicode.Hangul, r) ||
		unicode.Is(unicode.Katakana, r) ||
		unicode.Is(unicode.Hiragana, r) ||
		(r >= 0xFF01 && r <= 0xFF60) ||
		(r >= 0xFFE0 && r <= 0xFFE6)
}

// padRight pads s to the given display width.
func padRight(s string, width int) string {
	dw := displayWidth(s)
	if dw >= width {
		return s
	}
	return s + strings.Repeat(" ", width-dw)
}

// truncateToWidth truncates s so its display width fits within max columns.
func truncateToWidth(s string, max int) string {
	if max <= 0 {
		return ""
	}
	w := 0
	for i, r := range s {
		rw := 1
		if isCJK(r) {
			rw = 2
		}
		if w+rw > max {
			if max >= 2 {
				return s[:i] + "…"
			}
			return s[:i]
		}
		w += rw
	}
	return s
}

func printTable(ctx *Context, s *schema.ProtocolSchema, meta *protocol.ProtocolMeta) error {
	lang := ctx.Lang
	tw := termWidth()

	// ── Protocol header ──
	fmt.Printf("%s%s%s %s(version %s)%s  byte_order: %s\n",
		cBold+cCyan, s.Name, cReset,
		cDim, s.Version, cReset,
		s.DefaultByteOrder)

	if meta != nil {
		title := metaTitle(meta, lang)
		if title != "" {
			fmt.Printf("%sRFC %s%s - %s\n", cYellow, meta.RFC, cReset, title)
		}
		if meta.URL != "" {
			fmt.Printf("%s%s%s\n", cDim, meta.URL, cReset)
		}
		// Protocol description
		if desc := metaLang(meta.Description, lang); desc != "" {
			fmt.Printf("%s%s%s\n", cDim, desc, cReset)
		}
		var info []string
		if meta.Status != "" {
			sc := cGreen
			if meta.Status == "obsoleted" {
				sc = cYellow
			}
			info = append(info, fmt.Sprintf("status: %s%s%s", sc, meta.Status, cReset))
		}
		if meta.Layer != "" {
			info = append(info, fmt.Sprintf("layer: %s%s%s", cCyan, meta.Layer, cReset))
		}
		if len(meta.SeeAlso) > 0 {
			info = append(info, fmt.Sprintf("see also: %s%s%s", cDim, strings.Join(meta.SeeAlso, ", "), cReset))
		}
		if len(info) > 0 {
			fmt.Println(strings.Join(info, "  "))
		}
		// Protocol stack chain
		if chain := buildProtocolChain(ctx, s.Name); chain != "" {
			fmt.Printf("%s%s%s\n", cDim, chain, cReset)
		}
	}
	fmt.Println()

	// ── Column layout ──
	const (
		pad     = 2 // space between columns
		colName = 22
		colType = 6
		colBits = 5
	)
	fixedWidth := 2 + colName + pad + colType + pad + colBits + pad // 2 for left indent
	descWidth := max(tw-fixedWidth, 10)

	hName := msg(ctx, "show.col.name")
	hType := msg(ctx, "show.col.type")
	hBits := msg(ctx, "show.col.bits")
	hDesc := msg(ctx, "show.col.desc")

	// Header
	fmt.Printf("  %s%s  %s  %s  %s%s\n",
		cBold,
		padRight(hName, colName), padRight(hType, colType), padRight(hBits, colBits), hDesc,
		cReset)
	lineWidth := min(tw, 100)
	fmt.Printf("  %s%s%s\n", cDim, strings.Repeat("─", lineWidth-2), cReset)

	// ── Fields ──
	var fields []schema.FieldDef
	for _, f := range s.Fields {
		if f.IsBitfieldGroup {
			fields = append(fields, f.BitfieldFields...)
		} else {
			fields = append(fields, f)
		}
	}

	lblVar := msg(ctx, "show.variable")
	lblRem := msg(ctx, "show.remaining")

	for _, f := range fields {
		typeName := f.Type.String()
		bitsStr := fmt.Sprintf("%d", f.BitWidth)
		if f.Type == schema.Bytes || f.Type == schema.String {
			if f.FixedLength > 0 {
				bitsStr = fmt.Sprintf("%d", f.FixedLength*8)
			} else if f.LengthRef != nil {
				bitsStr = lblVar
			} else {
				bitsStr = lblRem
			}
		}

		// Build description
		desc := fieldDescText(f, lang, meta)
		mods := fieldModifiers(f)

		var descCol string
		if desc != "" && mods != "" {
			modsWidth := displayWidth(mods)
			availDesc := descWidth - modsWidth - 2
			if availDesc < 6 {
				// Not enough room for both, just show desc
				descCol = truncateToWidth(desc, descWidth)
			} else {
				descCol = truncateToWidth(desc, availDesc) + "  " + cDim + mods + cReset
			}
		} else if desc != "" {
			descCol = truncateToWidth(desc, descWidth)
		} else if mods != "" {
			descCol = cDim + truncateToWidth(mods, descWidth) + cReset
		}

		fmt.Printf("  %s%s%s  %s  %s  %s\n",
			cGreen, padRight(f.Name, colName), cReset,
			padRight(typeName, colType),
			padRight(bitsStr, colBits),
			descCol)
	}

	return nil
}

func metaTitle(meta *protocol.ProtocolMeta, lang string) string {
	return metaLang(meta.Title, lang)
}

// metaLang returns the localized string from a map, with fallback to English.
func metaLang(m map[string]string, lang string) string {
	if t, ok := m[lang]; ok {
		return t
	}
	if t, ok := m["en"]; ok {
		return t
	}
	return ""
}

func fieldDescText(f schema.FieldDef, lang string, meta *protocol.ProtocolMeta) string {
	if meta == nil {
		return ""
	}
	if fm, ok := meta.Fields[f.Name]; ok {
		if d, ok := fm[lang]; ok {
			return d
		}
		if d, ok := fm["en"]; ok {
			return d
		}
	}
	return ""
}

func fieldModifiers(f schema.FieldDef) string {
	var parts []string
	if f.Checksum != nil {
		parts = append(parts, "✓"+f.Checksum.Algorithm)
	}
	if f.LengthRef != nil {
		parts = append(parts, "↔"+f.LengthRef.FieldName)
	}
	if f.DisplayFormat != "" {
		parts = append(parts, "◈"+f.DisplayFormat)
	}
	if f.Condition != nil {
		parts = append(parts, fmt.Sprintf("?%s%s%v", f.Condition.FieldName, f.Condition.Operator, f.Condition.Value))
	}
	return strings.Join(parts, " ")
}

// buildProtocolChain builds a protocol stack chain string like "Ethernet → IPv4 → TCP → HTTP".
func buildProtocolChain(ctx *Context, name string) string {
	// Walk up the depends_on chain
	var chain []string
	visited := make(map[string]bool)
	current := name

	for {
		if visited[current] {
			break
		}
		visited[current] = true
		chain = append([]string{current}, chain...)

		meta := ctx.Lib.Meta(current)
		if meta == nil || len(meta.DependsOn) == 0 {
			break
		}
		// Follow the first dependency
		current = meta.DependsOn[0]
	}

	if len(chain) <= 1 {
		return ""
	}

	return strings.Join(chain, " → ")
}

// printMessageTable displays a message protocol in table format.
func printMessageTable(ctx *Context, ms *schema.MessageSchema, meta *protocol.ProtocolMeta) error {
	lang := ctx.Lang
	tw := termWidth()

	// Header
	fmt.Printf("%s%s%s %s(version %s)%s  transport: %s\n",
		cBold+cCyan, ms.Name, cReset,
		cDim, ms.Version, cReset,
		ms.Transport)

	if meta != nil {
		title := metaTitle(meta, lang)
		if title != "" {
			if meta.RFC != "" {
				fmt.Printf("%sRFC %s%s - %s\n", cYellow, meta.RFC, cReset, title)
			} else {
				fmt.Printf("%s%s%s\n", cCyan, title, cReset)
			}
		}
		if meta.URL != "" {
			fmt.Printf("%s%s%s\n", cDim, meta.URL, cReset)
		}
		if desc := metaLang(meta.Description, lang); desc != "" {
			fmt.Printf("%s%s%s\n", cDim, desc, cReset)
		}
		// Protocol stack chain
		if chain := buildProtocolChain(ctx, ms.Name); chain != "" {
			fmt.Printf("%s%s%s\n", cDim, chain, cReset)
		}
	}
	fmt.Println()

	// Messages
	for _, msg := range ms.Messages {
		kindColor := cGreen
		if msg.Kind == "response" {
			kindColor = cYellow
		} else if msg.Kind == "notification" {
			kindColor = cCyan
		}
		fmt.Printf("  %s%s%s %s%s%s\n", kindColor, msg.Kind, cReset, cBold, msg.Name, cReset)

		if len(msg.Fields) == 0 {
			fmt.Printf("    %s(no fields)%s\n", cDim, cReset)
		} else {
			const colName = 22
			const colType = 10
			lineWidth := min(tw, 80)
			fmt.Printf("    %s%s%s\n", cDim, strings.Repeat("─", lineWidth-4), cReset)
			for _, f := range msg.Fields {
				printMessageFieldRow(f, "    ", colName, colType, lang, meta)
			}
		}
		fmt.Println()
	}

	return nil
}

func printMessageFieldRow(f schema.MessageFieldDef, indent string, colName, colType int, lang string, meta *protocol.ProtocolMeta) {
	typeName := f.Type.String()
	if f.Optional {
		typeName += "?"
	}

	desc := ""
	if meta != nil {
		if fm, ok := meta.Fields[f.Name]; ok {
			if d, ok := fm[lang]; ok {
				desc = d
			} else if d, ok := fm["en"]; ok {
				desc = d
			}
		}
	}

	fmt.Printf("%s%s%s%s  %s%s%s", indent,
		cGreen, padRight(f.Name, colName), cReset,
		padRight(typeName, colType), cDim, desc)
	if desc != "" {
		fmt.Print(cReset)
	}
	fmt.Println()

	// Print nested fields
	if f.Type == schema.MsgObject && len(f.Fields) > 0 {
		for _, sub := range f.Fields {
			printMessageFieldRow(sub, indent+"  ", colName-2, colType, lang, meta)
		}
	}
}
