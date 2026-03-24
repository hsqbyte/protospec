package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/chaos"
)

func runChaos(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl chaos <protocol> [--target host:port]")
	}
	target := "localhost:8080"
	for i := 1; i < len(args)-1; i++ {
		if args[i] == "--target" {
			target = args[i+1]
		}
	}
	s := chaos.NewScenario("chaos-1", args[0], target)
	s.AddFault(chaos.FaultCorrupt, 0.1, "random field")
	s.AddFault(chaos.FaultDelay, 0.2, "50-200ms")
	s.AddFault(chaos.FaultDrop, 0.05, "")
	fmt.Print(s.Describe())
	return nil
}
