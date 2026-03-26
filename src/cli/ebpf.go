package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/integrations/ebpf"
)

func runEBPF(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl ebpf <protocol> [--type xdp|tc]")
	}
	progType := ebpf.ProgXDP
	for i := 1; i < len(args)-1; i++ {
		if args[i] == "--type" && args[i+1] == "tc" {
			progType = ebpf.ProgTC
		}
	}
	prog := ebpf.Generate(args[0], progType)
	fmt.Print(prog.ToC())
	return nil
}
