package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/tenant"
)

func runTenant(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl tenant <create|list|audit>")
	}
	mgr := tenant.NewManager()
	switch args[0] {
	case "create":
		name := "default"
		if len(args) >= 2 {
			name = args[1]
		}
		return mgr.CreateTenant(name, name, tenant.Quota{MaxProtocols: 100, MaxRequests: 1000})
	case "list":
		fmt.Print(mgr.Describe())
	case "audit":
		id := "default"
		if len(args) >= 2 {
			id = args[1]
		}
		entries := mgr.GetAuditLog(id)
		for _, e := range entries {
			fmt.Printf("  [%s] %s — %s\n", e.TenantID, e.Action, e.Resource)
		}
	default:
		return fmt.Errorf("unknown tenant subcommand: %s", args[0])
	}
	return nil
}
