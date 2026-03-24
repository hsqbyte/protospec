package cli

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
)

func runEncode(ctx *Context, name string, jsonInput string) error {
	var raw []byte
	var err error

	if jsonInput != "" {
		raw = []byte(jsonInput)
	} else if ctx.InputFile != "" {
		raw, err = os.ReadFile(ctx.InputFile)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("%s", msg(ctx, "encode.no_input"))
	}

	var packet map[string]any
	if err := json.Unmarshal(raw, &packet); err != nil {
		return fmt.Errorf("%s: %w", msg(ctx, "encode.invalid_json"), err)
	}

	// Convert JSON values for the codec:
	// - float64 → uint64 (JSON numbers)
	// - base64 string for bytes fields → []byte
	for k, v := range packet {
		switch val := v.(type) {
		case float64:
			packet[k] = uint64(val)
		case string:
			// Try base64 decode for potential bytes fields
			if decoded, err := base64.StdEncoding.DecodeString(val); err == nil && len(val) > 0 && len(val)%4 == 0 {
				packet[k] = decoded
			}
		}
	}

	data, err := ctx.Lib.Encode(name, packet)
	if err != nil {
		return err
	}

	return writeOutput(ctx.OutputFile, data)
}
