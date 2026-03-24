package cli

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/hsqbyte/protospec/src/stack"
)

func runDecapsulate(ctx *Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: psl decapsulate <start-protocol> <hex-data>")
	}

	hexStr := strings.ReplaceAll(args[1], " ", "")
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		return fmt.Errorf("invalid hex: %w", err)
	}

	engine := stack.NewEngine(ctx.Lib)
	layers, err := engine.Decode(args[0], data)
	if err != nil {
		return err
	}

	if ctx.Raw {
		return json.NewEncoder(os.Stdout).Encode(layers)
	}

	for i, l := range layers {
		fmt.Printf("Layer %d: %s (%d bytes)\n", i+1, l.Protocol, l.Bytes)
		for k, v := range l.Fields {
			fmt.Printf("  %s: %v\n", k, v)
		}
	}
	return nil
}
