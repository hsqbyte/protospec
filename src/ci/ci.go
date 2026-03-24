// Package ci provides CI/CD integration tools for PSL projects.
package ci

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

// TestSuite represents a JUnit XML test suite.
type TestSuite struct {
	XMLName  xml.Name   `xml:"testsuite"`
	Name     string     `xml:"name,attr"`
	Tests    int        `xml:"tests,attr"`
	Failures int        `xml:"failures,attr"`
	Time     float64    `xml:"time,attr"`
	Cases    []TestCase `xml:"testcase"`
}

// TestCase represents a JUnit XML test case.
type TestCase struct {
	Name    string   `xml:"name,attr"`
	Time    float64  `xml:"time,attr"`
	Failure *Failure `xml:"failure,omitempty"`
}

// Failure represents a test failure.
type Failure struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Body    string `xml:",chardata"`
}

// ValidationResult holds the result of validating a PSL file.
type ValidationResult struct {
	File    string        `json:"file"`
	Valid   bool          `json:"valid"`
	Error   string        `json:"error,omitempty"`
	Elapsed time.Duration `json:"elapsed"`
}

// GenerateJUnitXML generates JUnit XML from validation results.
func GenerateJUnitXML(results []ValidationResult) (string, error) {
	suite := TestSuite{Name: "PSL Validation", Tests: len(results)}
	for _, r := range results {
		tc := TestCase{Name: r.File, Time: r.Elapsed.Seconds()}
		if !r.Valid {
			suite.Failures++
			tc.Failure = &Failure{
				Message: r.Error,
				Type:    "ValidationError",
				Body:    r.Error,
			}
		}
		suite.Cases = append(suite.Cases, tc)
	}
	suite.Time = 0
	for _, tc := range suite.Cases {
		suite.Time += tc.Time
	}

	data, err := xml.MarshalIndent(suite, "", "  ")
	if err != nil {
		return "", err
	}
	return xml.Header + string(data), nil
}

// GitHubActionsTemplate returns a GitHub Actions workflow template.
func GitHubActionsTemplate() string {
	return `name: PSL CI
on: [push, pull_request]
jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      - run: go install github.com/protospec/psl@latest
      - run: psl ci validate protocols/
      - run: psl ci test protocols/
`
}

// FormatResults formats validation results for CI output.
func FormatResults(results []ValidationResult) string {
	var b strings.Builder
	passed, failed := 0, 0
	for _, r := range results {
		if r.Valid {
			passed++
			b.WriteString(fmt.Sprintf("✓ %s\n", r.File))
		} else {
			failed++
			b.WriteString(fmt.Sprintf("✗ %s — %s\n", r.File, r.Error))
		}
	}
	b.WriteString(fmt.Sprintf("\n%d passed, %d failed, %d total\n", passed, failed, len(results)))
	return b.String()
}
