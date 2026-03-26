package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/platform/cloud"
)

func runCloud(ctx *Context, args []string) error {
	addr := "localhost:8091"
	if len(args) > 0 {
		addr = args[0]
	}

	fmt.Printf("Starting PSL Cloud API at http://%s\n", addr)
	server := cloud.NewServer(ctx.Lib, addr)
	return server.Start()
}
