package cli

import (
	"fmt"
	"os"

	"github.com/hsqbyte/protospec/psl"
	"github.com/hsqbyte/protospec/src/lint"
)

func runLint(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl lint <file.psl>")
	}
	data, err := os.ReadFile(args[0])
	if err != nil {
		return err
	}
	l := lint.NewLinterWithTransport(nil, psl.FS)
	issues := l.LintContent(args[0], string(data))
	fmt.Print(lint.FormatIssues(issues))
	if len(issues) > 0 {
		return fmt.Errorf("lint failed with %d issue(s)", len(issues))
	}
	return nil
}
