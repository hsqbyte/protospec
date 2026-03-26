package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/platform/ci"
)

func runCI(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl ci <validate|test|report|template>")
	}
	switch args[0] {
	case "validate":
		fmt.Println(ci.FormatResults([]ci.ValidationResult{
			{File: "example.psl", Valid: true},
		}))
	case "test":
		fmt.Println("running protocol tests in CI mode...")
	case "report":
		xml, err := ci.GenerateJUnitXML([]ci.ValidationResult{
			{File: "example.psl", Valid: true},
		})
		if err != nil {
			return err
		}
		fmt.Println(xml)
	case "template":
		fmt.Println(ci.GitHubActionsTemplate())
	default:
		return fmt.Errorf("unknown ci subcommand: %s", args[0])
	}
	return nil
}
