package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hsqbyte/protospec/src/testgen"
)

func runTestGen(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl testgen <protocol> [--format json|go]")
	}

	protoName := args[0]
	format := "json"
	for i, a := range args {
		if a == "--format" && i+1 < len(args) {
			format = args[i+1]
		}
	}

	s, err := ctx.Lib.Registry().GetSchema(protoName)
	if err != nil {
		return fmt.Errorf("protocol %q not found: %w", protoName, err)
	}

	gen := testgen.NewGenerator()
	tests := gen.Generate(s)

	switch format {
	case "go":
		output := testgen.FormatGoTest(protoName, tests)
		if ctx.OutputFile != "" {
			return os.WriteFile(ctx.OutputFile, []byte(output), 0644)
		}
		fmt.Print(output)
	default:
		if ctx.OutputFile != "" {
			data, _ := json.MarshalIndent(tests, "", "  ")
			return os.WriteFile(ctx.OutputFile, data, 0644)
		}
		return json.NewEncoder(os.Stdout).Encode(tests)
	}
	return nil
}
