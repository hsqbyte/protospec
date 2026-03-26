package cli

import (
	"fmt"
	"os"

	"github.com/hsqbyte/protospec/src/docs/apidoc"
)

func runAPIDoc(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl apidoc <protocol> [--format openapi|asyncapi]")
	}

	protoName := args[0]
	format := "openapi"
	for i, a := range args {
		if a == "--format" && i+1 < len(args) {
			format = args[i+1]
		}
	}

	ms := ctx.Lib.Message(protoName)
	if ms == nil {
		return fmt.Errorf("message protocol %q not found", protoName)
	}

	gen := apidoc.NewGenerator()
	var output string
	var err error

	switch format {
	case "asyncapi":
		output, err = gen.GenerateAsyncAPI(ms)
	default:
		output, err = gen.GenerateOpenAPI(ms)
	}
	if err != nil {
		return err
	}

	if ctx.OutputFile != "" {
		return os.WriteFile(ctx.OutputFile, []byte(output), 0644)
	}
	fmt.Println(output)
	return nil
}
