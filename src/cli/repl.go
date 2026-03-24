package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/repl"
)

func runREPL(ctx *Context, args []string) error {
	s := repl.NewSession()
	fmt.Println("PSL REPL v0.87.0 — type 'help' for commands")
	// Non-interactive demo mode
	commands := []string{"help", `let pkt = decode("IPv4", "4500...")`, "bindings"}
	for _, cmd := range commands {
		fmt.Printf("psl> %s\n", cmd)
		out, err := s.Execute(cmd)
		if err != nil {
			fmt.Printf("error: %v\n", err)
		} else if out != "" {
			fmt.Println(out)
		}
	}
	return nil
}
