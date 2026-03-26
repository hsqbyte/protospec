// Package etl provides protocol data ETL pipeline.
package etl

import (
	"fmt"
	"strings"
)

// SourceType represents an ETL data source type.
type SourceType string

const (
	SourcePCAP      SourceType = "pcap"
	SourceInterface SourceType = "interface"
	SourceKafka     SourceType = "kafka"
	SourceFile      SourceType = "file"
)

// SinkType represents an ETL data sink type.
type SinkType string

const (
	SinkFile          SinkType = "file"
	SinkDatabase      SinkType = "database"
	SinkElasticsearch SinkType = "elasticsearch"
	SinkS3            SinkType = "s3"
)

// TransformOp represents a transform operation.
type TransformOp struct {
	Type   string `json:"type"` // decode, filter, aggregate, format
	Config string `json:"config"`
}

// Pipeline represents an ETL pipeline configuration.
type Pipeline struct {
	Name       string        `json:"name"`
	Source     SourceType    `json:"source"`
	SourceURI  string        `json:"source_uri"`
	Transforms []TransformOp `json:"transforms"`
	Sink       SinkType      `json:"sink"`
	SinkURI    string        `json:"sink_uri"`
}

// NewPipeline creates a new ETL pipeline.
func NewPipeline(name string) *Pipeline {
	return &Pipeline{Name: name}
}

// SetSource sets the pipeline source.
func (p *Pipeline) SetSource(t SourceType, uri string) *Pipeline {
	p.Source = t
	p.SourceURI = uri
	return p
}

// AddTransform adds a transform step.
func (p *Pipeline) AddTransform(op TransformOp) *Pipeline {
	p.Transforms = append(p.Transforms, op)
	return p
}

// SetSink sets the pipeline sink.
func (p *Pipeline) SetSink(t SinkType, uri string) *Pipeline {
	p.Sink = t
	p.SinkURI = uri
	return p
}

// Describe returns a human-readable pipeline description.
func (p *Pipeline) Describe() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Pipeline: %s\n", p.Name))
	b.WriteString(fmt.Sprintf("  Source: %s (%s)\n", p.Source, p.SourceURI))
	for i, t := range p.Transforms {
		b.WriteString(fmt.Sprintf("  Step %d: %s [%s]\n", i+1, t.Type, t.Config))
	}
	b.WriteString(fmt.Sprintf("  Sink: %s (%s)\n", p.Sink, p.SinkURI))
	return b.String()
}

// Run executes the ETL pipeline (stub).
func (p *Pipeline) Run() error {
	fmt.Printf("running ETL pipeline '%s'...\n", p.Name)
	fmt.Printf("  reading from %s: %s\n", p.Source, p.SourceURI)
	for _, t := range p.Transforms {
		fmt.Printf("  applying %s: %s\n", t.Type, t.Config)
	}
	fmt.Printf("  writing to %s: %s\n", p.Sink, p.SinkURI)
	return nil
}
