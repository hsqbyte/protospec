package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/integrations/dpdk"
)

func runDPDK(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl dpdk <pipeline|generate> [protocol]")
	}
	switch args[0] {
	case "pipeline":
		p := dpdk.NewPipeline("default", 4)
		p.AddStage("rx", "rx", "port=0")
		p.AddStage("parse", "parse", "protocol=auto")
		p.AddStage("classify", "classify", "5-tuple")
		p.AddStage("tx", "tx", "port=1")
		fmt.Print(p.Describe())
	case "generate":
		proto := "IPv4"
		if len(args) >= 2 {
			proto = args[1]
		}
		fmt.Print(dpdk.GenerateC(proto))
	default:
		return fmt.Errorf("unknown dpdk subcommand: %s", args[0])
	}
	return nil
}
