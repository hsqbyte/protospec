package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/psl4"
)

func runPSL4(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl psl4 <features|infer>")
	}
	switch args[0] {
	case "features":
		f := psl4.NewFeatures()
		fmt.Print(f.Describe())
	case "infer":
		fields := map[string]string{"src_port": "", "dst_addr": "", "payload": "bytes"}
		results := psl4.InferTypes(fields)
		for _, r := range results {
			fmt.Printf("  %s → %s (confidence: %.0f%%)\n", r.Field, r.InferredType, r.Confidence*100)
		}
	default:
		return fmt.Errorf("unknown psl4 subcommand: %s", args[0])
	}
	return nil
}
