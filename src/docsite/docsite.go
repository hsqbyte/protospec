// Package docsite generates a static documentation website from PSL protocols.
package docsite

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hsqbyte/protospec/src/protocol"
	"github.com/hsqbyte/protospec/src/schema"
)

// Config holds documentation site generation configuration.
type Config struct {
	Title   string
	Lang    string // "en" or "zh"
	OutDir  string
	BaseURL string
	Lib     *protocol.Library
}

// Generate creates a static documentation site.
func Generate(cfg *Config) error {
	if err := os.MkdirAll(cfg.OutDir, 0o755); err != nil {
		return err
	}

	// Generate index page
	if err := generateIndex(cfg); err != nil {
		return err
	}

	// Generate per-protocol pages
	names := cfg.Lib.AllNames()
	for _, name := range names {
		if err := generateProtocolPage(cfg, name); err != nil {
			fmt.Fprintf(os.Stderr, "warning: skip %s: %v\n", name, err)
		}
	}

	// Generate language reference
	if err := generateLangRef(cfg); err != nil {
		return err
	}

	// Generate API reference
	if err := generateAPIRef(cfg); err != nil {
		return err
	}

	return nil
}

func generateIndex(cfg *Config) error {
	names := cfg.Lib.AllNames()
	var b strings.Builder

	b.WriteString("<!DOCTYPE html>\n<html lang=\"" + cfg.Lang + "\">\n<head>\n")
	b.WriteString("<meta charset=\"utf-8\">\n")
	b.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n")
	b.WriteString(fmt.Sprintf("<title>%s</title>\n", cfg.Title))
	b.WriteString("<link rel=\"stylesheet\" href=\"style.css\">\n")
	b.WriteString("</head>\n<body>\n")
	b.WriteString("<nav><a href=\"index.html\">Home</a> | <a href=\"reference.html\">PSL Reference</a> | <a href=\"api.html\">API</a></nav>\n")
	b.WriteString(fmt.Sprintf("<h1>%s</h1>\n", cfg.Title))

	// Group by layer
	layers := map[string][]string{}
	for _, name := range names {
		meta := cfg.Lib.Meta(name)
		layer := "other"
		if meta != nil && meta.Layer != "" {
			layer = meta.Layer
		}
		layers[layer] = append(layers[layer], name)
	}

	layerOrder := []string{"link", "network", "transport", "application", "message", "other"}
	for _, layer := range layerOrder {
		protos, ok := layers[layer]
		if !ok || len(protos) == 0 {
			continue
		}
		sort.Strings(protos)
		b.WriteString(fmt.Sprintf("<h2>%s</h2>\n<ul>\n", strings.Title(layer)))
		for _, name := range protos {
			meta := cfg.Lib.Meta(name)
			title := name
			if meta != nil {
				if t, ok := meta.Title[cfg.Lang]; ok {
					title = t
				}
			}
			b.WriteString(fmt.Sprintf("  <li><a href=\"%s.html\">%s</a> — %s</li>\n",
				strings.ToLower(name), name, title))
		}
		b.WriteString("</ul>\n")
	}

	b.WriteString("</body>\n</html>\n")

	// Write CSS
	css := `body { font-family: -apple-system, sans-serif; max-width: 900px; margin: 0 auto; padding: 20px; }
nav { border-bottom: 1px solid #ddd; padding-bottom: 10px; margin-bottom: 20px; }
nav a { margin-right: 15px; text-decoration: none; color: #0066cc; }
h1 { color: #333; } h2 { color: #555; border-bottom: 1px solid #eee; padding-bottom: 5px; }
table { border-collapse: collapse; width: 100%; margin: 10px 0; }
th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
th { background: #f5f5f5; }
code { background: #f0f0f0; padding: 2px 6px; border-radius: 3px; }
pre { background: #f5f5f5; padding: 15px; border-radius: 5px; overflow-x: auto; }
.field-type { color: #0066cc; } .field-bits { color: #888; }
`
	if err := os.WriteFile(filepath.Join(cfg.OutDir, "style.css"), []byte(css), 0o644); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(cfg.OutDir, "index.html"), []byte(b.String()), 0o644)
}

func generateProtocolPage(cfg *Config, name string) error {
	meta := cfg.Lib.Meta(name)
	var b strings.Builder

	title := name
	desc := ""
	if meta != nil {
		if t, ok := meta.Title[cfg.Lang]; ok {
			title = t
		}
		if d, ok := meta.Description[cfg.Lang]; ok {
			desc = d
		}
	}

	b.WriteString("<!DOCTYPE html>\n<html lang=\"" + cfg.Lang + "\">\n<head>\n")
	b.WriteString("<meta charset=\"utf-8\">\n")
	b.WriteString(fmt.Sprintf("<title>%s — %s</title>\n", name, cfg.Title))
	b.WriteString("<link rel=\"stylesheet\" href=\"style.css\">\n")
	b.WriteString("</head>\n<body>\n")
	b.WriteString("<nav><a href=\"index.html\">Home</a> | <a href=\"reference.html\">PSL Reference</a> | <a href=\"api.html\">API</a></nav>\n")
	b.WriteString(fmt.Sprintf("<h1>%s</h1>\n", title))

	if desc != "" {
		b.WriteString(fmt.Sprintf("<p>%s</p>\n", desc))
	}

	// Meta info
	if meta != nil {
		b.WriteString("<table>\n")
		if meta.RFC != "" {
			b.WriteString(fmt.Sprintf("<tr><th>RFC</th><td><a href=\"%s\">%s</a></td></tr>\n", meta.URL, meta.RFC))
		}
		if meta.Layer != "" {
			b.WriteString(fmt.Sprintf("<tr><th>Layer</th><td>%s</td></tr>\n", meta.Layer))
		}
		if meta.Status != "" {
			b.WriteString(fmt.Sprintf("<tr><th>Status</th><td>%s</td></tr>\n", meta.Status))
		}
		if len(meta.DependsOn) > 0 {
			deps := strings.Join(meta.DependsOn, ", ")
			b.WriteString(fmt.Sprintf("<tr><th>Depends On</th><td>%s</td></tr>\n", deps))
		}
		b.WriteString("</table>\n")
	}

	// Check if message protocol
	if ms := cfg.Lib.Message(name); ms != nil {
		b.WriteString(generateMessageDoc(ms, meta, cfg.Lang))
	} else {
		// Binary protocol fields
		s, err := cfg.Lib.Registry().GetSchema(name)
		if err == nil {
			b.WriteString(generateBinaryDoc(s, meta, cfg.Lang))
		}
	}

	b.WriteString("</body>\n</html>\n")

	filename := strings.ToLower(name) + ".html"
	return os.WriteFile(filepath.Join(cfg.OutDir, filename), []byte(b.String()), 0o644)
}

func generateBinaryDoc(s *schema.ProtocolSchema, meta *protocol.ProtocolMeta, lang string) string {
	var b strings.Builder
	b.WriteString("<h2>Fields</h2>\n")
	b.WriteString("<table>\n<tr><th>Name</th><th>Type</th><th>Bits</th><th>Description</th></tr>\n")

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
			if fm, ok := meta.Fields[f.Name]; ok {
				if d, ok := fm[lang]; ok {
					desc = d
				} else if d, ok := fm["en"]; ok {
					desc = d
				}
			}
		}
		b.WriteString(fmt.Sprintf("<tr><td><code>%s</code></td><td class=\"field-type\">%s</td><td class=\"field-bits\">%d</td><td>%s</td></tr>\n",
			f.Name, f.Type.String(), f.BitWidth, desc))
	}
	b.WriteString("</table>\n")

	// Constants
	if len(s.Constants) > 0 {
		b.WriteString("<h2>Constants</h2>\n<table>\n<tr><th>Name</th><th>Value</th></tr>\n")
		for k, v := range s.Constants {
			b.WriteString(fmt.Sprintf("<tr><td><code>%s</code></td><td>%d (0x%X)</td></tr>\n", k, v, v))
		}
		b.WriteString("</table>\n")
	}

	return b.String()
}

func generateMessageDoc(ms *schema.MessageSchema, meta *protocol.ProtocolMeta, lang string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("<p>Transport: <code>%s</code></p>\n", ms.Transport))
	b.WriteString("<h2>Messages</h2>\n")

	for _, msg := range ms.Messages {
		b.WriteString(fmt.Sprintf("<h3>%s <small>(%s)</small></h3>\n", msg.Name, msg.Kind))
		b.WriteString("<table>\n<tr><th>Field</th><th>Type</th><th>Required</th></tr>\n")
		for _, f := range msg.Fields {
			req := "✓"
			if f.Optional {
				req = ""
			}
			b.WriteString(fmt.Sprintf("<tr><td><code>%s</code></td><td>%s</td><td>%s</td></tr>\n",
				f.Name, f.Type.String(), req))
		}
		b.WriteString("</table>\n")
	}

	return b.String()
}

func generateLangRef(cfg *Config) error {
	var b strings.Builder
	b.WriteString("<!DOCTYPE html>\n<html lang=\"" + cfg.Lang + "\">\n<head>\n")
	b.WriteString("<meta charset=\"utf-8\">\n")
	b.WriteString(fmt.Sprintf("<title>PSL Language Reference — %s</title>\n", cfg.Title))
	b.WriteString("<link rel=\"stylesheet\" href=\"style.css\">\n")
	b.WriteString("</head>\n<body>\n")
	b.WriteString("<nav><a href=\"index.html\">Home</a> | <a href=\"reference.html\">PSL Reference</a> | <a href=\"api.html\">API</a></nav>\n")
	b.WriteString("<h1>PSL Language Reference</h1>\n")

	b.WriteString(`<h2>Binary Protocol</h2>
<pre>protocol MyProtocol {
    version "1.0";
    byte_order big-endian;

    const MAGIC = 0xCAFE;

    field version: uint8;
    field length: uint16;

    bitfield {
        field flags: uint4;
        field type: uint4;
    }

    field payload: bytes length_ref=length;
    field checksum: uint16 checksum="internet-checksum" covers=version..payload;
}</pre>

<h2>Message Protocol</h2>
<pre>message MyAPI {
    version "1.0";
    transport jsonrpc;

    request GetUser {
        field id: number;
    }

    response GetUserResult {
        field name: string;
        field email: string;
    }
}</pre>

<h2>Keywords</h2>
<table>
<tr><th>Keyword</th><th>Description</th></tr>
<tr><td><code>protocol</code></td><td>Define a binary protocol</td></tr>
<tr><td><code>message</code></td><td>Define a message protocol</td></tr>
<tr><td><code>field</code></td><td>Define a field</td></tr>
<tr><td><code>bitfield</code></td><td>Group of sub-byte fields</td></tr>
<tr><td><code>const</code></td><td>Define a constant</td></tr>
<tr><td><code>version</code></td><td>Protocol version</td></tr>
<tr><td><code>byte_order</code></td><td>Default byte order</td></tr>
<tr><td><code>checksum</code></td><td>Checksum algorithm</td></tr>
<tr><td><code>when</code></td><td>Conditional field</td></tr>
<tr><td><code>enum</code></td><td>Value-to-name mapping</td></tr>
<tr><td><code>import</code></td><td>Import another protocol</td></tr>
<tr><td><code>extends</code></td><td>Inherit from protocol</td></tr>
<tr><td><code>type</code></td><td>Type alias</td></tr>
</table>

<h2>Types</h2>
<table>
<tr><th>Type</th><th>Description</th></tr>
<tr><td><code>uint8..uint64</code></td><td>Unsigned integer (bit width)</td></tr>
<tr><td><code>int8..int64</code></td><td>Signed integer</td></tr>
<tr><td><code>bool</code></td><td>Boolean (1 bit)</td></tr>
<tr><td><code>bytes</code></td><td>Byte sequence</td></tr>
<tr><td><code>bytes[N]</code></td><td>Fixed-length bytes</td></tr>
<tr><td><code>string</code></td><td>UTF-8 string</td></tr>
</table>
`)

	b.WriteString("</body>\n</html>\n")
	return os.WriteFile(filepath.Join(cfg.OutDir, "reference.html"), []byte(b.String()), 0o644)
}

func generateAPIRef(cfg *Config) error {
	var b strings.Builder
	b.WriteString("<!DOCTYPE html>\n<html lang=\"" + cfg.Lang + "\">\n<head>\n")
	b.WriteString("<meta charset=\"utf-8\">\n")
	b.WriteString(fmt.Sprintf("<title>Go SDK API — %s</title>\n", cfg.Title))
	b.WriteString("<link rel=\"stylesheet\" href=\"style.css\">\n")
	b.WriteString("</head>\n<body>\n")
	b.WriteString("<nav><a href=\"index.html\">Home</a> | <a href=\"reference.html\">PSL Reference</a> | <a href=\"api.html\">API</a></nav>\n")
	b.WriteString("<h1>Go SDK API Reference</h1>\n")

	b.WriteString(`<h2>protocol.Library</h2>
<pre>lib, err := protocol.NewLibrary()

// Decode binary data
result, err := lib.Decode("IPv4", data)
fmt.Println(result.Packet)

// Encode fields
data, err := lib.Encode("UDP", map[string]any{
    "source_port": 8080,
    "dest_port":   53,
    "length":      12,
    "checksum":    0,
})

// List protocols
names := lib.AllNames()

// Load custom PSL
err = lib.LoadPSL("my_protocol.psl")

// Create standalone codec
codec, err := lib.CreateCodec(pslText)
result, err := codec.Decode(data)
</pre>

<h2>Multi-language SDK</h2>
<p>Generate SDK bindings:</p>
<pre>psl sdk python -o ./sdk/    # Python ctypes binding
psl sdk rust -o ./sdk/      # Rust FFI binding
psl sdk typescript -o ./sdk/ # TypeScript/Node.js binding
psl sdk header -o ./sdk/    # C header file
psl sdk cmake -o ./sdk/     # CMake example
psl sdk all -o ./sdk/       # All bindings</pre>
`)

	b.WriteString("</body>\n</html>\n")
	return os.WriteFile(filepath.Join(cfg.OutDir, "api.html"), []byte(b.String()), 0o644)
}

// GenerateJSON exports all protocol metadata as JSON for doc site consumption.
func GenerateJSON(lib *protocol.Library, outDir string) error {
	type protoInfo struct {
		Name        string                       `json:"name"`
		Title       map[string]string            `json:"title,omitempty"`
		Description map[string]string            `json:"description,omitempty"`
		RFC         string                       `json:"rfc,omitempty"`
		Layer       string                       `json:"layer,omitempty"`
		Status      string                       `json:"status,omitempty"`
		Type        string                       `json:"type,omitempty"`
		DependsOn   []string                     `json:"depends_on,omitempty"`
		Fields      map[string]map[string]string `json:"fields,omitempty"`
	}

	var protos []protoInfo
	for _, name := range lib.AllNames() {
		info := protoInfo{Name: name}
		if meta := lib.Meta(name); meta != nil {
			info.Title = meta.Title
			info.Description = meta.Description
			info.RFC = meta.RFC
			info.Layer = meta.Layer
			info.Status = meta.Status
			info.Type = meta.Type
			info.DependsOn = meta.DependsOn
			info.Fields = meta.Fields
		}
		protos = append(protos, info)
	}

	data, err := json.MarshalIndent(protos, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(outDir, "protocols.json"), data, 0o644)
}

// unused import guard
var _ = sort.Strings
