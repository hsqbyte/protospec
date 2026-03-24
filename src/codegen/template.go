package codegen

import (
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/hsqbyte/protospec/src/protocol"
	"github.com/hsqbyte/protospec/src/schema"
)

// TemplateData holds the data passed to custom templates.
type TemplateData struct {
	Name      string
	Version   string
	ByteOrder string
	Fields    []TemplateField
}

// TemplateField holds field data for templates.
type TemplateField struct {
	Name      string
	Type      string
	BitWidth  int
	ByteWidth int
	Comment   string
}

// GenerateFromTemplate generates code using a custom Go text/template file.
func (g *Generator) GenerateFromTemplate(name, templateFile string) (string, error) {
	s, err := g.lib.Registry().GetSchema(name)
	if err != nil {
		return "", err
	}

	tmplContent, err := os.ReadFile(templateFile)
	if err != nil {
		return "", fmt.Errorf("read template: %w", err)
	}

	tmpl, err := template.New("custom").Funcs(templateFuncs()).Parse(string(tmplContent))
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	data := buildTemplateData(s, g.lib.Meta(s.Name))
	var b strings.Builder
	if err := tmpl.Execute(&b, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return b.String(), nil
}

func buildTemplateData(s *schema.ProtocolSchema, meta *protocol.ProtocolMeta) TemplateData {
	td := TemplateData{
		Name:      s.Name,
		Version:   s.Version,
		ByteOrder: s.DefaultByteOrder.String(),
	}

	var fields []schema.FieldDef
	for _, f := range s.Fields {
		if f.IsBitfieldGroup {
			fields = append(fields, f.BitfieldFields...)
		} else {
			fields = append(fields, f)
		}
	}

	for _, f := range fields {
		tf := TemplateField{
			Name:     f.Name,
			Type:     f.Type.String(),
			BitWidth: f.BitWidth,
		}
		if f.BitWidth > 0 {
			tf.ByteWidth = (f.BitWidth + 7) / 8
		}
		tf.Comment = fieldComment(f, meta)
		td.Fields = append(td.Fields, tf)
	}
	return td
}

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"upper":  strings.ToUpper,
		"lower":  strings.ToLower,
		"title":  strings.Title,
		"pascal": goExportName,
		"snake":  rustFieldName,
	}
}
