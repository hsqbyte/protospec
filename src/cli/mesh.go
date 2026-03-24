package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/mesh"
)

func runMesh(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl mesh <config|envoy> <protocol>")
	}
	proto := "HTTP"
	if len(args) >= 2 {
		proto = args[1]
	}
	cfg := mesh.NewConfig(proto)
	switch args[0] {
	case "config":
		fmt.Print(cfg.Describe())
	case "envoy":
		fmt.Print(cfg.GenerateEnvoyConfig())
	default:
		return fmt.Errorf("unknown mesh subcommand: %s", args[0])
	}
	return nil
}
