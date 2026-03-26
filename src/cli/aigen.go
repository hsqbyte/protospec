package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/platform/aigen"
)

func runAIGen(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl ai <generate|review|testgen> [args]")
	}
	switch args[0] {
	case "generate":
		desc := "a simple binary protocol with version, type and length"
		if len(args) >= 2 {
			desc = args[1]
		}
		fmt.Print(aigen.GeneratePSL(desc))
	case "review":
		content := "protocol Test version \"1.0\" { header { type: uint8; } }"
		r := aigen.ReviewPSL(content)
		fmt.Print(aigen.FormatReview(r))
	case "testgen":
		proto := "IPv4"
		if len(args) >= 2 {
			proto = args[1]
		}
		tests := aigen.GenerateTests(proto)
		for _, t := range tests {
			fmt.Printf("  %s: %s (input=%s, expected=%s)\n", t.Name, t.Description, t.Input, t.Expected)
		}
	default:
		return fmt.Errorf("unknown ai subcommand: %s", args[0])
	}
	return nil
}
