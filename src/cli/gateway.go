package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/integrations/gateway"
)

func runGateway(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl gateway <init|describe|config>")
	}
	switch args[0] {
	case "init":
		cfg := gateway.DefaultConfig()
		gw := gateway.NewGateway(cfg)
		fmt.Println(gw.ExportConfig())
	case "describe":
		cfg := gateway.DefaultConfig()
		gw := gateway.NewGateway(cfg)
		fmt.Print(gw.Describe())
	case "config":
		cfg := gateway.DefaultConfig()
		gw := gateway.NewGateway(cfg)
		fmt.Println(gw.ExportConfig())
	default:
		return fmt.Errorf("unknown gateway subcommand: %s", args[0])
	}
	return nil
}
