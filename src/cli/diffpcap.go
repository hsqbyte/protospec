package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hsqbyte/protospec/src/diffpcap"
)

func runDiffPcap(ctx *Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: psl diff-pcap <a.pcap> <b.pcap>")
	}

	// Placeholder: in real implementation, would read and decode PCAP files
	fmt.Printf("Comparing %s vs %s\n", args[0], args[1])

	// Demo with empty comparison
	result := diffpcap.ComparePackets(nil, nil)

	if ctx.Raw {
		return json.NewEncoder(os.Stdout).Encode(result)
	}

	fmt.Print(diffpcap.FormatDiffResult(result))
	return nil
}
