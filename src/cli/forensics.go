package cli

import (
	"fmt"
	"os"

	"github.com/hsqbyte/protospec/src/integrations/forensics"
)

func runForensics(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl forensics <capture.pcap>")
	}

	data, err := os.ReadFile(args[0])
	if err != nil {
		return fmt.Errorf("read capture: %w", err)
	}

	analyzer := forensics.NewAnalyzer()
	report := analyzer.AnalyzeData(data)

	if ctx.OutputFile != "" {
		output := forensics.FormatReport(report)
		return os.WriteFile(ctx.OutputFile, []byte(output), 0644)
	}

	fmt.Print(forensics.FormatReport(report))
	return nil
}
