package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/platform/standards"
)

func runStandards(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl standards <list|find> [protocol|org]")
	}
	reg := standards.NewRegistry()
	switch args[0] {
	case "list":
		fmt.Print(reg.Describe())
	case "find":
		if len(args) < 2 {
			return fmt.Errorf("usage: psl standards find <protocol>")
		}
		results := reg.FindByProtocol(args[1])
		for _, s := range results {
			fmt.Printf("  [%s] %s — %s\n    %s\n", s.Org, s.ID, s.Title, s.URL)
		}
		if len(results) == 0 {
			fmt.Println("no standards found")
		}
	default:
		return fmt.Errorf("unknown standards subcommand: %s", args[0])
	}
	return nil
}
