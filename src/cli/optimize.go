package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/tools/optimize"
)

func runOptimize(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl optimize <protocol>")
	}

	s, err := ctx.Lib.Registry().GetSchema(args[0])
	if err != nil {
		return fmt.Errorf("protocol %q not found: %w", args[0], err)
	}

	analyzer := optimize.NewAnalyzer()
	suggestions := analyzer.Analyze(s)
	fmt.Print(optimize.FormatSuggestions(suggestions))
	return nil
}
