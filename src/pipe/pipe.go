// Package pipe provides protocol processing pipelines.
package pipe

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hsqbyte/protospec/src/protocol"
)

// Stage represents a pipeline processing stage.
type Stage struct {
	Type     string            `json:"type"` // "decode", "filter", "transform", "encode"
	Protocol string            `json:"protocol,omitempty"`
	Config   map[string]string `json:"config,omitempty"`
}

// Pipeline is a sequence of processing stages.
type Pipeline struct {
	Name   string  `json:"name"`
	Stages []Stage `json:"stages"`
}

// Engine executes pipelines.
type Engine struct {
	lib *protocol.Library
}

// NewEngine creates a new pipeline engine.
func NewEngine(lib *protocol.Library) *Engine {
	return &Engine{lib: lib}
}

// Result holds the output of a pipeline execution.
type Result struct {
	Stage   string `json:"stage"`
	Type    string `json:"type"`
	Data    any    `json:"data"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// Execute runs a pipeline on the given input data.
func (e *Engine) Execute(p *Pipeline, input []byte) ([]Result, error) {
	var results []Result
	current := input

	for i, stage := range p.Stages {
		stageName := fmt.Sprintf("%d:%s", i+1, stage.Type)
		switch stage.Type {
		case "decode":
			decoded, err := e.lib.Decode(stage.Protocol, current)
			if err != nil {
				results = append(results, Result{Stage: stageName, Type: "decode", Success: false, Error: err.Error()})
				return results, err
			}
			results = append(results, Result{Stage: stageName, Type: "decode", Data: decoded.Packet, Success: true})

		case "encode":
			// Re-encode from last decoded fields
			if len(results) == 0 {
				return results, fmt.Errorf("encode stage requires prior decode")
			}
			lastData, ok := results[len(results)-1].Data.(map[string]any)
			if !ok {
				return results, fmt.Errorf("encode stage requires map data")
			}
			encoded, err := e.lib.Encode(stage.Protocol, lastData)
			if err != nil {
				results = append(results, Result{Stage: stageName, Type: "encode", Success: false, Error: err.Error()})
				return results, err
			}
			current = encoded
			results = append(results, Result{Stage: stageName, Type: "encode", Data: hex.EncodeToString(encoded), Success: true})

		case "filter":
			// Filter based on field value
			field := stage.Config["field"]
			value := stage.Config["value"]
			if len(results) > 0 {
				if lastData, ok := results[len(results)-1].Data.(map[string]any); ok {
					if fmt.Sprintf("%v", lastData[field]) != value {
						results = append(results, Result{Stage: stageName, Type: "filter", Data: "filtered out", Success: true})
						return results, nil
					}
				}
			}
			results = append(results, Result{Stage: stageName, Type: "filter", Data: "passed", Success: true})

		case "transform":
			results = append(results, Result{Stage: stageName, Type: "transform", Data: "passthrough", Success: true})

		default:
			return results, fmt.Errorf("unknown stage type: %s", stage.Type)
		}
	}
	return results, nil
}

// ParsePipeline parses a pipeline definition from JSON.
func ParsePipeline(data string) (*Pipeline, error) {
	var p Pipeline
	if err := json.Unmarshal([]byte(data), &p); err != nil {
		return nil, fmt.Errorf("parse pipeline: %w", err)
	}
	return &p, nil
}

// ParsePipelineShort parses a short pipeline notation like "decode:IPv4|filter:version=4|encode:IPv4".
func ParsePipelineShort(notation string) (*Pipeline, error) {
	p := &Pipeline{Name: "inline"}
	parts := strings.Split(notation, "|")
	for _, part := range parts {
		kv := strings.SplitN(part, ":", 2)
		if len(kv) < 2 {
			return nil, fmt.Errorf("invalid stage: %s", part)
		}
		stage := Stage{Type: kv[0]}
		// Check for filter config
		if stage.Type == "filter" {
			fv := strings.SplitN(kv[1], "=", 2)
			if len(fv) == 2 {
				stage.Config = map[string]string{"field": fv[0], "value": fv[1]}
			}
		} else {
			stage.Protocol = kv[1]
		}
		p.Stages = append(p.Stages, stage)
	}
	return p, nil
}
