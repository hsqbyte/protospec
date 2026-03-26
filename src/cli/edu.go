package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/platform/edu"
)

func runEdu(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl edu <lessons|paths|badges>")
	}
	p := edu.NewPlatform()
	switch args[0] {
	case "lessons":
		fmt.Print(p.ListLessons())
	case "paths":
		fmt.Print(p.ListPaths())
	case "badges":
		fmt.Print(p.ListBadges())
	default:
		return fmt.Errorf("unknown edu subcommand: %s", args[0])
	}
	return nil
}
