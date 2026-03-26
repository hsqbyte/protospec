package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/integrations/autocomply"
)

func runAutoComply(ctx *Context, args []string) error {
	proto := "IPv4"
	if len(args) >= 1 {
		proto = args[0]
	}
	rules := autocomply.DefaultRules()
	report := autocomply.Check(proto, rules)
	fmt.Print(report.GenerateReport())
	return nil
}
