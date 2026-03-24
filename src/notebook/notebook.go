// Package notebook provides Jupyter notebook integration for PSL.
package notebook

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Cell represents a notebook cell.
type Cell struct {
	Type   string   `json:"cell_type"`
	Source []string `json:"source"`
}

// Notebook represents a Jupyter notebook.
type Notebook struct {
	Cells    []Cell         `json:"cells"`
	Metadata map[string]any `json:"metadata"`
}

// NewNotebook creates a new notebook.
func NewNotebook() *Notebook {
	return &Notebook{
		Metadata: map[string]any{
			"kernelspec": map[string]string{
				"display_name": "PSL",
				"language":     "psl",
				"name":         "psl",
			},
		},
	}
}

// AddMarkdown adds a markdown cell.
func (nb *Notebook) AddMarkdown(lines ...string) {
	nb.Cells = append(nb.Cells, Cell{Type: "markdown", Source: lines})
}

// AddCode adds a code cell.
func (nb *Notebook) AddCode(lines ...string) {
	nb.Cells = append(nb.Cells, Cell{Type: "code", Source: lines})
}

// ToJSON exports the notebook as JSON.
func (nb *Notebook) ToJSON() string {
	data, _ := json.MarshalIndent(nb, "", "  ")
	return string(data)
}

// TemplateAnalysis generates a protocol analysis notebook template.
func TemplateAnalysis(protocol string) *Notebook {
	nb := NewNotebook()
	nb.AddMarkdown(fmt.Sprintf("# %s Protocol Analysis", protocol))
	nb.AddCode(fmt.Sprintf("pkt = decode(\"%s\", read_pcap(\"capture.pcap\"))", protocol))
	nb.AddCode("pkt.summary()")
	nb.AddCode("pkt.visualize()")
	return nb
}

// TemplateSecurity generates a security audit notebook template.
func TemplateSecurity(protocol string) *Notebook {
	nb := NewNotebook()
	nb.AddMarkdown(fmt.Sprintf("# %s Security Audit", protocol))
	nb.AddCode(fmt.Sprintf("results = audit(\"%s\", read_pcap(\"capture.pcap\"))", protocol))
	nb.AddCode("results.vulnerabilities()")
	nb.AddCode("results.report()")
	return nb
}

// KernelSpec returns the PSL Jupyter kernel specification.
func KernelSpec() string {
	spec := map[string]any{
		"display_name": "PSL",
		"language":     "psl",
		"argv":         []string{"psl", "kernel", "--connection-file", "{connection_file}"},
	}
	data, _ := json.MarshalIndent(spec, "", "  ")
	return string(data)
}

// ListTemplates returns available notebook templates.
func ListTemplates() string {
	var b strings.Builder
	b.WriteString("Available notebook templates:\n")
	b.WriteString("  analysis   — Protocol analysis workflow\n")
	b.WriteString("  security   — Security audit workflow\n")
	b.WriteString("  benchmark  — Performance analysis workflow\n")
	return b.String()
}
