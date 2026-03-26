package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/platform/playground"
)

func runPlayground(ctx *Context, args []string) error {
	addr := "localhost:8090"
	if len(args) > 0 {
		addr = args[0]
	}

	fmt.Printf("Starting PSL Playground at http://%s\n", addr)
	server := playground.NewServer(ctx.Lib, addr)
	return server.Start()
}
