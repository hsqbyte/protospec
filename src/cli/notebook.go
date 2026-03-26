package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/platform/notebook"
)

func runNotebook(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl notebook <new|template|kernel>")
	}
	switch args[0] {
	case "new":
		proto := "IPv4"
		if len(args) >= 2 {
			proto = args[1]
		}
		nb := notebook.TemplateAnalysis(proto)
		fmt.Println(nb.ToJSON())
	case "template":
		fmt.Print(notebook.ListTemplates())
	case "kernel":
		fmt.Println(notebook.KernelSpec())
	default:
		return fmt.Errorf("unknown notebook subcommand: %s", args[0])
	}
	return nil
}
