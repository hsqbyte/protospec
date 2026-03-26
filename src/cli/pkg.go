package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/platform/pkg"
)

func runPkg(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl pkg <init|publish|install|search>")
	}
	switch args[0] {
	case "init":
		name := "my-protocol"
		if len(args) >= 2 {
			name = args[1]
		}
		return pkg.InitPackage(".", name)
	case "publish":
		fmt.Println("publishing package to registry...")
		return nil
	case "install":
		if len(args) < 2 {
			return fmt.Errorf("usage: psl pkg install <name>")
		}
		r := pkg.NewRegistry("https://registry.psl.dev")
		return r.Install(args[1], ".")
	case "search":
		if len(args) < 2 {
			return fmt.Errorf("usage: psl pkg search <query>")
		}
		r := pkg.NewRegistry("https://registry.psl.dev")
		results := r.Search(args[1])
		for _, p := range results {
			fmt.Printf("%s@%s — %s\n", p.Name, p.Version, p.Description)
		}
		if len(results) == 0 {
			fmt.Println("no packages found")
		}
		return nil
	default:
		return fmt.Errorf("unknown pkg subcommand: %s", args[0])
	}
}
