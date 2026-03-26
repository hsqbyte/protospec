package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/platform/metrics"
)

func runMetrics(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl metrics <export|dashboard>")
	}
	switch args[0] {
	case "export":
		c := metrics.NewCollector()
		fmt.Print(c.PrometheusExport())
	case "dashboard":
		fmt.Println(metrics.GrafanaDashboard())
	default:
		return fmt.Errorf("unknown metrics subcommand: %s", args[0])
	}
	return nil
}
