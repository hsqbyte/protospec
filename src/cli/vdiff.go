package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/tools/vdiff"
)

func runVisualDiff(ctx *Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: psl vdiff <a.psl> <b.psl> [--format text|html]")
	}
	format := "text"
	for i := 2; i < len(args)-1; i++ {
		if args[i] == "--format" {
			format = args[i+1]
		}
	}
	d := vdiff.Compare(args[0], args[1], []string{"version", "ihl", "total_length"}, []string{"version", "traffic_class", "flow_label"})
	switch format {
	case "html":
		fmt.Print(d.ToHTML())
	default:
		fmt.Print(d.ToText())
		fmt.Println(d.Summary())
	}
	return nil
}
