package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/pqcrypto"
)

func runPQCrypto(ctx *Context, args []string) error {
	proto := "TLS"
	if len(args) >= 1 {
		proto = args[0]
	}
	a := pqcrypto.Assess(proto)
	fmt.Print(a.Describe())
	return nil
}
