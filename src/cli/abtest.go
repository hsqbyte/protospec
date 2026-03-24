package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/abtest"
)

func runABTest(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl abtest <create|analyze> <protocol>")
	}
	proto := "HTTP"
	if len(args) >= 2 {
		proto = args[1]
	}
	switch args[0] {
	case "create":
		exp := abtest.NewExperiment("test-1", proto)
		exp.AddVariant("control", "1.1", 0.5)
		exp.AddVariant("candidate", "2.0", 0.5)
		fmt.Printf("created experiment for %s with %d variants\n", proto, len(exp.Variants))
	case "analyze":
		exp := abtest.NewExperiment("test-1", proto)
		exp.AddVariant("control", "1.1", 0.5)
		exp.AddVariant("candidate", "2.0", 0.5)
		exp.Variants[0].Metrics = abtest.Metrics{Latency: 12.5, ErrorRate: 0.01, Throughput: 1000}
		exp.Variants[1].Metrics = abtest.Metrics{Latency: 8.3, ErrorRate: 0.005, Throughput: 1500}
		fmt.Print(exp.Analyze())
	default:
		return fmt.Errorf("unknown abtest subcommand: %s", args[0])
	}
	return nil
}
