package cli

import (
	"fmt"
	"os"

	"github.com/hsqbyte/protospec/src/compliance"
)

func runCompliance(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl compliance <protocol> [--format html|json] [-o file]")
	}

	name := args[0]
	format := "text"
	outFile := ""

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--format":
			i++
			if i < len(args) {
				format = args[i]
			}
		case "-o":
			i++
			if i < len(args) {
				outFile = args[i]
			}
		}
	}

	runner := compliance.NewRunner(ctx.Lib)
	report, err := runner.RunCompliance(name, nil)
	if err != nil {
		return err
	}

	var output string
	switch format {
	case "html":
		output = report.ToHTML()
	case "json":
		output, err = report.ToJSON()
		if err != nil {
			return err
		}
	default:
		output = fmt.Sprintf("Compliance Report: %s\n", report.Protocol)
		output += fmt.Sprintf("Total: %d | PASS: %d | WARN: %d | FAIL: %d\n",
			report.Summary.Total, report.Summary.Passed, report.Summary.Warned, report.Summary.Failed)
		for _, r := range report.Results {
			output += fmt.Sprintf("  [%s] %s — %s\n", r.Level, r.TestCase.Name, r.Message)
		}
	}

	if outFile != "" {
		return os.WriteFile(outFile, []byte(output), 0o644)
	}
	fmt.Print(output)
	return nil
}
