package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/contract"
)

func runContract(ctx *Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: psl contract <provider.psl> <consumer.psl>")
	}
	c := &contract.Contract{Provider: args[0], Consumer: args[1], Version: "1.0"}
	provider := []string{"version", "header_length", "total_length", "src_ip", "dst_ip"}
	consumer := []string{"version", "src_ip", "dst_ip"}
	r := contract.Verify(provider, consumer)
	fmt.Print(contract.FormatResult(c, r))
	return nil
}
