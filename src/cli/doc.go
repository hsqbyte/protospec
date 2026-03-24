package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hsqbyte/protospec/src/protocol"
	"github.com/hsqbyte/protospec/src/schema"
)

func runDoc(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl doc <protocol|--all> [--format md|html] [-o dir]")
	}

	var (
		protocols []string
		all       bool
		format    = "md"
		outDir    = ""
	)

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--all":
			all = true
		case "--format":
			i++
			if i < len(args) {
				format = args[i]
			}
		case "-o":
			i++
			if i < len(args) {
				outDir = args[i]
			}
		default:
			protocols = append(protocols, args[i])
		}
	}

	if all {
		protocols = ctx.Lib.AllNames()
	}

	for _, name := range protocols {
		var doc string
		if ms := ctx.Lib.Message(name); ms != nil {
			doc = generateMessageDoc(ms, ctx.Lib.Meta(name), ctx.Lang, format)
		} else {
			s, err := ctx.Lib.Registry().GetSchema(name)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: skip %s: %v\n", name, err)
				continue
			}
			doc = generateBinaryDoc(s, ctx.Lib.Meta(name), ctx.Lang, format)
		}

		if outDir != "" {
			os.MkdirAll(outDir, 0o755)
			ext := ".md"
			if format == "html" {
				ext = ".html"
			}
			path := filepath.Join(outDir, strings.ToLower(name)+ext)
			if err := os.WriteFile(path, []byte(doc), 0o644); err != nil {
				return err
			}
			fmt.Printf("generated %s\n", path)
		} else {
			fmt.Print(doc)
		}
	}
	return nil
}

func generateBinaryDoc(s *schema.ProtocolSchema, meta *protocol.ProtocolMeta, lang, format string) string {
	if format == "html" {
		return generateBinaryDocHTML(s, meta, lang)
	}
	return generateBinaryDocMD(s, meta, lang)
}

func generateBinaryDocMD(s *schema.ProtocolSchema, meta *protocol.ProtocolMeta, lang string) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# %s\n\n", s.Name))

	if meta != nil {
		if title := metaLang(meta.Title, lang); title != "" {
			b.WriteString(fmt.Sprintf("> %s\n\n", title))
		}
		if desc := metaLang(meta.Description, lang); desc != "" {
			b.WriteString(desc + "\n\n")
		}
		if meta.RFC != "" {
			b.WriteString(fmt.Sprintf("- RFC: %s\n", meta.RFC))
		}
		if meta.Layer != "" {
			b.WriteString(fmt.Sprintf("- Layer: %s\n", meta.Layer))
		}
		if meta.Status != "" {
			b.WriteString(fmt.Sprintf("- Status: %s\n", meta.Status))
		}
		if meta.URL != "" {
			b.WriteString(fmt.Sprintf("- URL: %s\n", meta.URL))
		}
		b.WriteString("\n")
	}

	b.WriteString(fmt.Sprintf("- Version: %s\n", s.Version))
	b.WriteString(fmt.Sprintf("- Byte Order: %s\n\n", s.DefaultByteOrder))

	// Fields table
	b.WriteString("## Fields\n\n")
	b.WriteString("| Field | Type | Bits | Description |\n")
	b.WriteString("|-------|------|------|-------------|\n")

	var fields []schema.FieldDef
	for _, f := range s.Fields {
		if f.IsBitfieldGroup {
			fields = append(fields, f.BitfieldFields...)
		} else {
			fields = append(fields, f)
		}
	}

	for _, f := range fields {
		desc := ""
		if meta != nil {
			desc = fieldDescText(f, lang, meta)
		}
		bitsStr := fmt.Sprintf("%d", f.BitWidth)
		if f.BitWidth == 0 {
			bitsStr = "variable"
		}
		b.WriteString(fmt.Sprintf("| `%s` | %s | %s | %s |\n",
			f.Name, f.Type, bitsStr, desc))
	}
	b.WriteString("\n")

	return b.String()
}

func generateBinaryDocHTML(s *schema.ProtocolSchema, meta *protocol.ProtocolMeta, lang string) string {
	var b strings.Builder

	b.WriteString("<!DOCTYPE html>\n<html><head>\n")
	b.WriteString(fmt.Sprintf("<title>%s Protocol</title>\n", s.Name))
	b.WriteString(`<style>
body { font-family: -apple-system, sans-serif; max-width: 800px; margin: 40px auto; padding: 0 20px; color: #333; }
h1 { color: #1a73e8; }
table { border-collapse: collapse; width: 100%; margin: 20px 0; }
th, td { border: 1px solid #ddd; padding: 8px 12px; text-align: left; }
th { background: #f5f5f5; }
code { background: #f0f0f0; padding: 2px 6px; border-radius: 3px; }
.meta { color: #666; font-size: 0.9em; }
</style>
</head><body>
`)
	b.WriteString(fmt.Sprintf("<h1>%s</h1>\n", s.Name))

	if meta != nil {
		if title := metaLang(meta.Title, lang); title != "" {
			b.WriteString(fmt.Sprintf("<p class=\"meta\">%s</p>\n", title))
		}
		if desc := metaLang(meta.Description, lang); desc != "" {
			b.WriteString(fmt.Sprintf("<p>%s</p>\n", desc))
		}
	}

	b.WriteString(fmt.Sprintf("<p>Version: %s | Byte Order: %s</p>\n", s.Version, s.DefaultByteOrder))

	b.WriteString("<h2>Fields</h2>\n<table>\n")
	b.WriteString("<tr><th>Field</th><th>Type</th><th>Bits</th><th>Description</th></tr>\n")

	var fields []schema.FieldDef
	for _, f := range s.Fields {
		if f.IsBitfieldGroup {
			fields = append(fields, f.BitfieldFields...)
		} else {
			fields = append(fields, f)
		}
	}

	for _, f := range fields {
		desc := ""
		if meta != nil {
			desc = fieldDescText(f, lang, meta)
		}
		bitsStr := fmt.Sprintf("%d", f.BitWidth)
		if f.BitWidth == 0 {
			bitsStr = "variable"
		}
		b.WriteString(fmt.Sprintf("<tr><td><code>%s</code></td><td>%s</td><td>%s</td><td>%s</td></tr>\n",
			f.Name, f.Type, bitsStr, desc))
	}
	b.WriteString("</table>\n</body></html>\n")

	return b.String()
}

func generateMessageDoc(ms *schema.MessageSchema, meta *protocol.ProtocolMeta, lang, format string) string {
	if format == "html" {
		return generateMessageDocHTML(ms, meta, lang)
	}
	return generateMessageDocMD(ms, meta, lang)
}

func generateMessageDocMD(ms *schema.MessageSchema, meta *protocol.ProtocolMeta, lang string) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# %s\n\n", ms.Name))
	if meta != nil {
		if title := metaLang(meta.Title, lang); title != "" {
			b.WriteString(fmt.Sprintf("> %s\n\n", title))
		}
	}
	b.WriteString(fmt.Sprintf("- Transport: %s\n", ms.Transport))
	b.WriteString(fmt.Sprintf("- Version: %s\n\n", ms.Version))

	if td := ms.TransportDef; td != nil {
		b.WriteString("## Transport Definition\n\n")
		b.WriteString(fmt.Sprintf("> Transport: %s v%s\n\n", td.Name, td.Version))
		for _, mt := range td.MessageTypes {
			b.WriteString(fmt.Sprintf("### %s\n\n", mt.Name))
			if len(mt.Fields) > 0 {
				b.WriteString("| Field | Type | Default |\n")
				b.WriteString("|-------|------|---------|\n")
				for _, f := range mt.Fields {
					def := ""
					if f.DefaultValue != nil {
						def = fmt.Sprintf("%v", f.DefaultValue)
					}
					b.WriteString(fmt.Sprintf("| `%s` | %s | %s |\n", f.Name, f.Type, def))
				}
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("## Messages\n\n")
	for _, msg := range ms.Messages {
		b.WriteString(fmt.Sprintf("### %s (%s)\n\n", msg.Name, msg.Kind))
		if len(msg.Fields) > 0 {
			b.WriteString("| Field | Type | Required |\n")
			b.WriteString("|-------|------|----------|\n")
			for _, f := range msg.Fields {
				req := "yes"
				if f.Optional {
					req = "no"
				}
				b.WriteString(fmt.Sprintf("| `%s` | %s | %s |\n", f.Name, f.Type, req))
			}
		}
		b.WriteString("\n")
		if msg.Response != nil {
			b.WriteString(fmt.Sprintf("#### %s Response\n\n", msg.Name))
			if len(msg.Response.Fields) > 0 {
				b.WriteString("| Field | Type | Required |\n")
				b.WriteString("|-------|------|----------|\n")
				for _, f := range msg.Response.Fields {
					req := "yes"
					if f.Optional {
						req = "no"
					}
					b.WriteString(fmt.Sprintf("| `%s` | %s | %s |\n", f.Name, f.Type, req))
				}
			}
			b.WriteString("\n")
		}
	}
	return b.String()
}

func generateMessageDocHTML(ms *schema.MessageSchema, meta *protocol.ProtocolMeta, lang string) string {
	var b strings.Builder
	b.WriteString("<!DOCTYPE html>\n<html><head>\n")
	b.WriteString(fmt.Sprintf("<title>%s Protocol</title>\n", ms.Name))
	b.WriteString(`<style>
body { font-family: -apple-system, sans-serif; max-width: 800px; margin: 40px auto; padding: 0 20px; }
table { border-collapse: collapse; width: 100%; margin: 10px 0; }
th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
th { background: #f5f5f5; }
code { background: #f0f0f0; padding: 2px 6px; border-radius: 3px; }
</style></head><body>
`)
	b.WriteString(fmt.Sprintf("<h1>%s</h1>\n", ms.Name))
	b.WriteString(fmt.Sprintf("<p>Transport: %s | Version: %s</p>\n", ms.Transport, ms.Version))

	if td := ms.TransportDef; td != nil {
		b.WriteString("<h2>Transport Definition</h2>\n")
		b.WriteString(fmt.Sprintf("<p>Transport: %s v%s</p>\n", td.Name, td.Version))
		for _, mt := range td.MessageTypes {
			b.WriteString(fmt.Sprintf("<h3>%s</h3>\n", mt.Name))
			if len(mt.Fields) > 0 {
				b.WriteString("<table><tr><th>Field</th><th>Type</th><th>Default</th></tr>\n")
				for _, f := range mt.Fields {
					def := ""
					if f.DefaultValue != nil {
						def = fmt.Sprintf("%v", f.DefaultValue)
					}
					b.WriteString(fmt.Sprintf("<tr><td><code>%s</code></td><td>%s</td><td>%s</td></tr>\n", f.Name, f.Type, def))
				}
				b.WriteString("</table>\n")
			}
		}
	}

	b.WriteString("<h2>Messages</h2>\n")
	for _, msg := range ms.Messages {
		b.WriteString(fmt.Sprintf("<h3>%s <small>(%s)</small></h3>\n", msg.Name, msg.Kind))
		if len(msg.Fields) > 0 {
			b.WriteString("<table><tr><th>Field</th><th>Type</th><th>Required</th></tr>\n")
			for _, f := range msg.Fields {
				req := "yes"
				if f.Optional {
					req = "no"
				}
				b.WriteString(fmt.Sprintf("<tr><td><code>%s</code></td><td>%s</td><td>%s</td></tr>\n", f.Name, f.Type, req))
			}
			b.WriteString("</table>\n")
		}
		if msg.Response != nil {
			b.WriteString(fmt.Sprintf("<h4>%s Response</h4>\n", msg.Name))
			if len(msg.Response.Fields) > 0 {
				b.WriteString("<table><tr><th>Field</th><th>Type</th><th>Required</th></tr>\n")
				for _, f := range msg.Response.Fields {
					req := "yes"
					if f.Optional {
						req = "no"
					}
					b.WriteString(fmt.Sprintf("<tr><td><code>%s</code></td><td>%s</td><td>%s</td></tr>\n", f.Name, f.Type, req))
				}
				b.WriteString("</table>\n")
			}
		}
	}
	b.WriteString("</body></html>\n")
	return b.String()
}
