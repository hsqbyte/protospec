package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/hsqbyte/protospec/src/identify"
)

func runIdentify(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl identify <hex-data>")
	}

	hexData := strings.Join(args, "")
	identifier := identify.NewIdentifier(ctx.Lib)
	candidates, err := identifier.IdentifyHex(hexData)
	if err != nil {
		return err
	}

	if len(candidates) == 0 {
		fmt.Println("No protocol identified.")
		return nil
	}

	if ctx.Raw {
		return json.NewEncoder(os.Stdout).Encode(candidates)
	}

	fmt.Println("Protocol identification results:")
	for i, c := range candidates {
		fmt.Printf("  %d. %s (confidence: %.0f%%) — %s\n", i+1, c.Protocol, c.Confidence*100, c.Reason)
	}
	return nil
}
