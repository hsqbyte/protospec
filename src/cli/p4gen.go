package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/p4gen"
)

func runP4Gen(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl p4 <protocol>")
	}
	fields := map[string]int{"version": 4, "ihl": 4, "total_length": 16, "protocol": 8}
	prog := p4gen.Generate(args[0], fields)
	fmt.Print(prog.ToP4())
	return nil
}
