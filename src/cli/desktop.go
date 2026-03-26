package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/platform/desktop"
)

func runDesktop(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl desktop <init|config|wails|tauri>")
	}
	switch args[0] {
	case "init", "config":
		cfg := desktop.DefaultConfig()
		fmt.Print(cfg.Describe())
	case "wails":
		fmt.Println(desktop.GenerateWailsConfig())
	case "tauri":
		fmt.Println(desktop.GenerateTauriConfig())
	default:
		return fmt.Errorf("unknown desktop subcommand: %s", args[0])
	}
	return nil
}
