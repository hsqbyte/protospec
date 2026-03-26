package cli

import (
	"fmt"
	"os"

	"github.com/hsqbyte/protospec/src/tools/formatter"
)

func runFmt(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl fmt <file.psl> [--check]")
	}
	checkOnly := false
	for _, a := range args[1:] {
		if a == "--check" {
			checkOnly = true
		}
	}
	data, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}
	source := string(data)
	opts := formatter.DefaultOptions()
	if checkOnly {
		if formatter.Check(source, opts) {
			fmt.Printf("%s: formatted\n", args[0])
			return nil
		}
		return fmt.Errorf("%s: not formatted", args[0])
	}
	formatted := formatter.Format(source, opts)
	return os.WriteFile(args[0], []byte(formatted), 0644)
}
