package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/fsm"
)

func runFSM(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl fsm <protocol> [--format mermaid|dot]")
	}
	proto := args[0]
	format := "mermaid"
	for i := 1; i < len(args)-1; i++ {
		if args[i] == "--format" {
			format = args[i+1]
		}
	}
	sm := fsm.NewStateMachine(proto)
	sm.AddState("Init", true, false)
	sm.AddState("Connected", false, false)
	sm.AddState("Closed", false, true)
	sm.AddTransition("Init", "Connected", "SYN/ACK")
	sm.AddTransition("Connected", "Closed", "FIN")

	switch format {
	case "mermaid":
		fmt.Print(sm.ToMermaid())
	case "dot":
		fmt.Print(sm.ToDOT())
	default:
		fmt.Print(sm.ToMermaid())
	}
	return nil
}
