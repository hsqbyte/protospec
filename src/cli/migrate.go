package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/migrate"
)

func runMigrate(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl migrate <protocol> --from <v1> --to <v2>")
	}
	proto := args[0]
	fromVer, toVer := "1.0", "2.0"
	for i := 1; i < len(args)-1; i++ {
		switch args[i] {
		case "--from":
			fromVer = args[i+1]
		case "--to":
			toVer = args[i+1]
		}
	}
	plan := migrate.GeneratePlan(proto, fromVer, toVer)
	fmt.Printf("migration plan: %s %s → %s (%d steps)\n", plan.Protocol, plan.FromVersion, plan.ToVersion, len(plan.Steps))
	fmt.Println(migrate.GenerateScript(plan))
	return nil
}
