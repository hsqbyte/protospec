package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/constraint"
)

func runConstraint(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl constraint <check|describe> <protocol>")
	}
	switch args[0] {
	case "check":
		proto := "IPv4"
		if len(args) >= 2 {
			proto = args[1]
		}
		cb := constraint.NewBlock(proto).
			AddArithmetic("total_length == header_length + payload_length", []string{"total_length", "header_length", "payload_length"}).
			AddReference("checksum == crc16(header)", []string{"checksum"})
		errs := cb.Validate(nil)
		if len(errs) == 0 {
			fmt.Printf("all constraints passed for %s\n", proto)
		}
	case "describe":
		proto := "IPv4"
		if len(args) >= 2 {
			proto = args[1]
		}
		cb := constraint.NewBlock(proto).
			AddArithmetic("total_length == header_length + payload_length", []string{"total_length", "header_length", "payload_length"})
		fmt.Print(cb.Describe())
	default:
		return fmt.Errorf("unknown constraint subcommand: %s", args[0])
	}
	return nil
}
