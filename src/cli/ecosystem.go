package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/platform/ecosystem"
)

func runEcosystem(ctx *Context, args []string) error {
	d := ecosystem.NewDashboard()
	fmt.Print(d.Describe())
	return nil
}
