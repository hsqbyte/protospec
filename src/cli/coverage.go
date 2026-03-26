package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/tools/coverage"
)

func runCoverage(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl coverage <protocol> [--pcap <file>] [--format text|json|html]")
	}
	format := "text"
	for i := 1; i < len(args)-1; i++ {
		if args[i] == "--format" {
			format = args[i+1]
		}
	}
	r := coverage.NewReport(args[0])
	r.AddField("version", "uint4", []string{"4"}, 0)
	r.AddField("ihl", "uint4", []string{"5"}, 0)
	r.AddField("dscp", "uint6", nil, 0)

	switch format {
	case "json":
		fmt.Println(r.ToJSON())
	case "html":
		fmt.Println(r.ToHTML())
	default:
		fmt.Print(r.ToText())
	}
	return nil
}
