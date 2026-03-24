// Package visual provides enhanced protocol visualization: PNG diagrams, interactive views.
package visual

import (
	"fmt"
	"strings"

	"github.com/hsqbyte/protospec/src/protocol"
	"github.com/hsqbyte/protospec/src/schema"
)

// DiagramConfig holds diagram generation configuration.
type DiagramConfig struct {
	Format          string // "ascii", "svg", "png", "html"
	ShowBits        bool
	ShowConditional bool
	Width           int
}

// GenerateInteractiveHTML generates an interactive HTML visualization.
func GenerateInteractiveHTML(lib *protocol.Library, name string) (string, error) {
	var fields []schema.FieldDef

	if ms := lib.Message(name); ms != nil {
		return generateMessageHTML(ms, lib.Meta(name)), nil
	}

	s, err := lib.Registry().GetSchema(name)
	if err != nil {
		return "", err
	}

	for _, f := range s.Fields {
		if f.IsBitfieldGroup {
			fields = append(fields, f.BitfieldFields...)
		} else {
			fields = append(fields, f)
		}
	}

	meta := lib.Meta(name)
	return generateBinaryHTML(s, fields, meta), nil
}

func generateBinaryHTML(s *schema.ProtocolSchema, fields []schema.FieldDef, meta *protocol.ProtocolMeta) string {
	var b strings.Builder
	title := s.Name
	if meta != nil {
		if t, ok := meta.Title["en"]; ok {
			title = t
		}
	}

	b.WriteString("<!DOCTYPE html>\n<html><head><meta charset=\"utf-8\">\n")
	b.WriteString(fmt.Sprintf("<title>%s — Interactive View</title>\n", title))
	b.WriteString(`<style>
body { font-family: -apple-system, sans-serif; max-width: 1000px; margin: 0 auto; padding: 20px; background: #fafafa; }
h1 { color: #333; }
.field-row { display: flex; margin: 2px 0; cursor: pointer; }
.field-box { padding: 8px 12px; border: 1px solid #ccc; background: #fff; text-align: center; font-size: 13px; transition: all 0.2s; }
.field-box:hover { background: #e3f2fd; border-color: #2196F3; }
.field-box.selected { background: #bbdefb; border-color: #1976D2; }
.info-panel { margin-top: 20px; padding: 15px; background: #fff; border: 1px solid #ddd; border-radius: 4px; display: none; }
.info-panel.active { display: block; }
.hex-view { font-family: monospace; background: #263238; color: #aed581; padding: 15px; border-radius: 4px; margin-top: 10px; }
.hex-view .highlight { background: #ff6f00; color: #fff; padding: 1px 2px; }
.bit-ruler { font-family: monospace; font-size: 11px; color: #999; margin-bottom: 5px; }
</style>
`)
	b.WriteString("</head><body>\n")
	b.WriteString(fmt.Sprintf("<h1>%s</h1>\n", title))

	// Bit ruler
	b.WriteString("<div class=\"bit-ruler\">")
	for i := 0; i < 32; i++ {
		b.WriteString(fmt.Sprintf("%d", i%10))
	}
	b.WriteString("</div>\n")

	// Field boxes
	b.WriteString("<div class=\"field-row\">\n")
	bitOffset := 0
	for i, f := range fields {
		width := f.BitWidth
		if width == 0 {
			width = 32 // variable
		}
		pxWidth := width * 8
		if pxWidth < 60 {
			pxWidth = 60
		}
		if pxWidth > 300 {
			pxWidth = 300
		}
		b.WriteString(fmt.Sprintf("  <div class=\"field-box\" style=\"width:%dpx\" onclick=\"selectField(%d)\" id=\"field-%d\">\n",
			pxWidth, i, i))
		b.WriteString(fmt.Sprintf("    <div>%s</div><div style=\"font-size:11px;color:#666\">%s %d bits</div>\n",
			f.Name, f.Type.String(), f.BitWidth))
		b.WriteString("  </div>\n")

		bitOffset += f.BitWidth
		if bitOffset >= 32 {
			b.WriteString("</div>\n<div class=\"field-row\">\n")
			bitOffset = 0
		}
	}
	b.WriteString("</div>\n")

	// Info panel
	b.WriteString("<div class=\"info-panel\" id=\"info-panel\">\n")
	b.WriteString("  <h3 id=\"info-name\"></h3>\n")
	b.WriteString("  <p id=\"info-desc\"></p>\n")
	b.WriteString("  <table id=\"info-table\"><tr><th>Property</th><th>Value</th></tr></table>\n")
	b.WriteString("</div>\n")

	// JavaScript
	b.WriteString("<script>\nconst fields = [\n")
	for _, f := range fields {
		desc := ""
		if meta != nil {
			if fm, ok := meta.Fields[f.Name]; ok {
				if d, ok := fm["en"]; ok {
					desc = d
				}
			}
		}
		b.WriteString(fmt.Sprintf("  {name:\"%s\",type:\"%s\",bits:%d,desc:\"%s\"},\n",
			f.Name, f.Type.String(), f.BitWidth, desc))
	}
	b.WriteString("];\n")
	b.WriteString(`
let selected = -1;
function selectField(i) {
  document.querySelectorAll('.field-box').forEach(el => el.classList.remove('selected'));
  document.getElementById('field-'+i).classList.add('selected');
  const f = fields[i];
  document.getElementById('info-name').textContent = f.name;
  document.getElementById('info-desc').textContent = f.desc || 'No description';
  const table = document.getElementById('info-table');
  table.innerHTML = '<tr><th>Property</th><th>Value</th></tr>';
  table.innerHTML += '<tr><td>Type</td><td>'+f.type+'</td></tr>';
  table.innerHTML += '<tr><td>Bit Width</td><td>'+f.bits+'</td></tr>';
  document.getElementById('info-panel').classList.add('active');
  selected = i;
}
`)
	b.WriteString("</script>\n</body></html>\n")
	return b.String()
}

func generateMessageHTML(ms *schema.MessageSchema, meta *protocol.ProtocolMeta) string {
	var b strings.Builder
	b.WriteString("<!DOCTYPE html>\n<html><head><meta charset=\"utf-8\">\n")
	b.WriteString(fmt.Sprintf("<title>%s — Messages</title>\n", ms.Name))
	b.WriteString("<style>body{font-family:sans-serif;max-width:800px;margin:0 auto;padding:20px}")
	b.WriteString(".msg{border:1px solid #ddd;margin:10px 0;padding:15px;border-radius:4px}")
	b.WriteString(".msg h3{margin-top:0}.req{border-left:3px solid #4CAF50}.res{border-left:3px solid #2196F3}.notif{border-left:3px solid #FF9800}")
	b.WriteString("</style></head><body>\n")
	b.WriteString(fmt.Sprintf("<h1>%s</h1>\n", ms.Name))

	for _, msg := range ms.Messages {
		cls := "msg"
		switch msg.Kind {
		case "request":
			cls += " req"
		case "response":
			cls += " res"
		default:
			cls += " notif"
		}
		b.WriteString(fmt.Sprintf("<div class=\"%s\">\n", cls))
		b.WriteString(fmt.Sprintf("<h3>%s <small>(%s)</small></h3>\n", msg.Name, msg.Kind))
		for _, f := range msg.Fields {
			opt := ""
			if f.Optional {
				opt = " (optional)"
			}
			b.WriteString(fmt.Sprintf("<p><code>%s</code>: %s%s</p>\n", f.Name, f.Type.String(), opt))
		}
		b.WriteString("</div>\n")
	}
	b.WriteString("</body></html>\n")
	return b.String()
}

// GenerateComparisonHTML generates a side-by-side protocol comparison.
func GenerateComparisonHTML(lib *protocol.Library, nameA, nameB string) (string, error) {
	var b strings.Builder
	b.WriteString("<!DOCTYPE html>\n<html><head><meta charset=\"utf-8\">\n")
	b.WriteString(fmt.Sprintf("<title>Compare: %s vs %s</title>\n", nameA, nameB))
	b.WriteString("<style>body{font-family:sans-serif;max-width:1200px;margin:0 auto;padding:20px}")
	b.WriteString(".compare{display:flex;gap:20px}.col{flex:1;border:1px solid #ddd;padding:15px;border-radius:4px}")
	b.WriteString("table{width:100%;border-collapse:collapse}th,td{border:1px solid #eee;padding:6px;text-align:left}th{background:#f5f5f5}")
	b.WriteString(".added{background:#e8f5e9}.removed{background:#ffebee}.changed{background:#fff3e0}")
	b.WriteString("</style></head><body>\n")
	b.WriteString(fmt.Sprintf("<h1>%s vs %s</h1>\n", nameA, nameB))
	b.WriteString("<div class=\"compare\">\n")

	for _, name := range []string{nameA, nameB} {
		b.WriteString("<div class=\"col\">\n")
		b.WriteString(fmt.Sprintf("<h2>%s</h2>\n", name))
		b.WriteString("<table><tr><th>Field</th><th>Type</th><th>Bits</th></tr>\n")
		if s, err := lib.Registry().GetSchema(name); err == nil {
			for _, f := range s.Fields {
				if f.IsBitfieldGroup {
					for _, sub := range f.BitfieldFields {
						b.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%d</td></tr>\n", sub.Name, sub.Type.String(), sub.BitWidth))
					}
				} else {
					b.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%d</td></tr>\n", f.Name, f.Type.String(), f.BitWidth))
				}
			}
		}
		b.WriteString("</table></div>\n")
	}
	b.WriteString("</div></body></html>\n")
	return b.String(), nil
}
