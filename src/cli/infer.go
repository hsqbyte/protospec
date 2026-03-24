package cli

import (
	"encoding/json"
	"fmt"

	"github.com/hsqbyte/protospec/src/infer"
)

func runInfer(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl infer <capture.pcap> [--name ProtoName] [--format psl|json]")
	}

	name := "Inferred"
	format := "psl"

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--name":
			i++
			if i < len(args) {
				name = args[i]
			}
		case "--format":
			i++
			if i < len(args) {
				format = args[i]
			}
		}
	}

	// For now, generate a sample inference with empty data
	// Real implementation would read PCAP
	result, err := infer.Infer([][]byte{{0x45, 0x00, 0x00, 0x3c, 0x00, 0x00, 0x40, 0x00}})
	if err != nil {
		return err
	}

	switch format {
	case "psl":
		fmt.Print(result.ToPSL(name))
	case "json":
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
	return nil
}
