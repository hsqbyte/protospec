package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/integrations/monitor"
)

func runMonitor(ctx *Context, args []string) error {
	iface := "eth0"
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--interface", "-i":
			i++
			if i < len(args) {
				iface = args[i]
			}
		}
	}

	stats := monitor.NewStats()
	_ = stats
	return fmt.Errorf("run manually: psl monitor requires long-running process on interface %s", iface)
}
