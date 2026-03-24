package cli

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/hsqbyte/protospec/src/pipe"
)

func runPipe(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl pipe <pipeline-notation> [hex-data]\nExample: psl pipe 'decode:IPv4' 4500003c...")
	}

	pipeline, err := pipe.ParsePipelineShort(args[0])
	if err != nil {
		return err
	}

	var input []byte
	if len(args) > 1 {
		hexStr := strings.ReplaceAll(args[1], " ", "")
		input, err = hex.DecodeString(hexStr)
		if err != nil {
			return fmt.Errorf("invalid hex input: %w", err)
		}
	}

	engine := pipe.NewEngine(ctx.Lib)
	results, err := engine.Execute(pipeline, input)
	if err != nil && len(results) == 0 {
		return err
	}

	if ctx.Raw {
		return json.NewEncoder(os.Stdout).Encode(results)
	}

	for _, r := range results {
		status := "✓"
		if !r.Success {
			status = "✗"
		}
		fmt.Printf("[%s] %s %s\n", status, r.Stage, formatPipeData(r.Data))
	}
	return nil
}

func formatPipeData(data any) string {
	if data == nil {
		return ""
	}
	switch v := data.(type) {
	case string:
		return v
	case map[string]any:
		b, _ := json.MarshalIndent(v, "  ", "  ")
		return "\n  " + string(b)
	default:
		return fmt.Sprintf("%v", v)
	}
}
