package cli

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TestCase defines a single protocol test case.
type TestCase struct {
	Name     string         `json:"name"`
	Protocol string         `json:"protocol"`
	Input    string         `json:"input"`  // hex-encoded bytes for decode test
	Expect   map[string]any `json:"expect"` // expected decoded fields
	Encode   map[string]any `json:"encode"` // fields to encode
	Output   string         `json:"output"` // expected hex-encoded bytes for encode test
}

// TestSuite is a collection of test cases.
type TestSuite struct {
	Protocol string     `json:"protocol"`
	Tests    []TestCase `json:"tests"`
}

// TestResult holds the result of a single test.
type TestResult struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Error   string `json:"error,omitempty"`
	Details string `json:"details,omitempty"`
}

func runTest(ctx *Context, args []string) error {
	var (
		all    bool
		format = "text"
		files  []string
	)

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--all":
			all = true
		case "--format":
			i++
			if i < len(args) {
				format = args[i]
			}
		default:
			files = append(files, args[i])
		}
	}

	if !all && len(files) == 0 {
		return fmt.Errorf("usage: psl test <file_test.json|--all> [--format text|json|junit]")
	}

	if all {
		// Find all *_test.json files in current directory
		matches, _ := filepath.Glob("*_test.json")
		files = append(files, matches...)
		// Also check psl/ subdirectories
		pslMatches, _ := filepath.Glob("psl/*/*_test.json")
		files = append(files, pslMatches...)
	}

	var allResults []TestResult
	for _, f := range files {
		results, err := runTestFile(ctx, f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading %s: %v\n", f, err)
			continue
		}
		allResults = append(allResults, results...)
	}

	switch format {
	case "json":
		data, _ := json.MarshalIndent(allResults, "", "  ")
		fmt.Println(string(data))
	case "junit":
		printJUnit(allResults)
	default:
		printTestResults(allResults)
	}

	// Exit with error if any test failed
	for _, r := range allResults {
		if !r.Passed {
			return fmt.Errorf("some tests failed")
		}
	}
	return nil
}

func runTestFile(ctx *Context, file string) ([]TestResult, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var suite TestSuite
	if err := json.Unmarshal(data, &suite); err != nil {
		return nil, err
	}

	var results []TestResult
	for _, tc := range suite.Tests {
		proto := tc.Protocol
		if proto == "" {
			proto = suite.Protocol
		}

		result := TestResult{Name: tc.Name}

		if tc.Input != "" && tc.Expect != nil {
			// Decode test
			inputBytes, err := hex.DecodeString(tc.Input)
			if err != nil {
				result.Error = fmt.Sprintf("invalid hex input: %v", err)
				results = append(results, result)
				continue
			}

			decoded, err := ctx.Lib.Decode(proto, inputBytes)
			if err != nil {
				result.Error = fmt.Sprintf("decode error: %v", err)
				results = append(results, result)
				continue
			}

			// Compare fields
			mismatches := compareFields(tc.Expect, decoded.Packet)
			if len(mismatches) > 0 {
				result.Error = "field mismatch"
				result.Details = strings.Join(mismatches, "; ")
			} else {
				result.Passed = true
			}
		} else if tc.Encode != nil && tc.Output != "" {
			// Encode test
			encoded, err := ctx.Lib.Encode(proto, tc.Encode)
			if err != nil {
				result.Error = fmt.Sprintf("encode error: %v", err)
				results = append(results, result)
				continue
			}

			actual := hex.EncodeToString(encoded)
			expected := strings.ToLower(tc.Output)
			if actual != expected {
				result.Error = fmt.Sprintf("encode mismatch: got %s, want %s", actual, expected)
			} else {
				result.Passed = true
			}
		} else {
			result.Error = "invalid test case: need input+expect or encode+output"
		}

		results = append(results, result)
	}
	return results, nil
}

func compareFields(expected, actual map[string]any) []string {
	var mismatches []string
	for k, ev := range expected {
		av, ok := actual[k]
		if !ok {
			mismatches = append(mismatches, fmt.Sprintf("missing field %q", k))
			continue
		}
		if fmt.Sprintf("%v", ev) != fmt.Sprintf("%v", av) {
			mismatches = append(mismatches, fmt.Sprintf("%s: got %v, want %v", k, av, ev))
		}
	}
	return mismatches
}

func printTestResults(results []TestResult) {
	passed := 0
	failed := 0
	for _, r := range results {
		if r.Passed {
			passed++
			fmt.Printf("  %s✓%s %s\n", cGreen, cReset, r.Name)
		} else {
			failed++
			fmt.Printf("  %s✗%s %s — %s\n", cYellow, cReset, r.Name, r.Error)
			if r.Details != "" {
				fmt.Printf("    %s%s%s\n", cDim, r.Details, cReset)
			}
		}
	}
	fmt.Printf("\n%d passed, %d failed, %d total\n", passed, failed, len(results))
}

func printJUnit(results []TestResult) {
	passed := 0
	failed := 0
	for _, r := range results {
		if r.Passed {
			passed++
		} else {
			failed++
		}
	}

	fmt.Printf("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	fmt.Printf("<testsuite tests=\"%d\" failures=\"%d\">\n", len(results), failed)
	for _, r := range results {
		if r.Passed {
			fmt.Printf("  <testcase name=\"%s\"/>\n", r.Name)
		} else {
			fmt.Printf("  <testcase name=\"%s\">\n", r.Name)
			fmt.Printf("    <failure message=\"%s\">%s</failure>\n", r.Error, r.Details)
			fmt.Printf("  </testcase>\n")
		}
	}
	fmt.Printf("</testsuite>\n")
}
