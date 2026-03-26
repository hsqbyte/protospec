// Package repl provides an interactive protocol shell.
package repl

import (
	"fmt"
	"strings"
)

// Binding represents a variable binding in the REPL.
type Binding struct {
	Name  string
	Value interface{}
}

// Session holds REPL session state.
type Session struct {
	Bindings map[string]interface{}
	History  []string
}

// NewSession creates a new REPL session.
func NewSession() *Session {
	return &Session{Bindings: make(map[string]interface{})}
}

// Execute executes a REPL command and returns the output.
func (s *Session) Execute(input string) (string, error) {
	s.History = append(s.History, input)
	trimmed := strings.TrimSpace(input)

	if trimmed == "" {
		return "", nil
	}
	if trimmed == "help" {
		return s.help(), nil
	}
	if trimmed == "history" {
		return strings.Join(s.History, "\n"), nil
	}
	if trimmed == "bindings" || trimmed == "vars" {
		return s.listBindings(), nil
	}
	if strings.HasPrefix(trimmed, "let ") {
		return s.handleLet(trimmed[4:])
	}
	if strings.Contains(trimmed, "|") {
		return s.handlePipe(trimmed)
	}
	return fmt.Sprintf("=> %s", trimmed), nil
}

func (s *Session) handleLet(expr string) (string, error) {
	parts := strings.SplitN(expr, "=", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("syntax: let <name> = <value>")
	}
	name := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	s.Bindings[name] = value
	return fmt.Sprintf("%s = %s", name, value), nil
}

func (s *Session) handlePipe(expr string) (string, error) {
	stages := strings.Split(expr, "|")
	result := strings.TrimSpace(stages[0])
	for _, stage := range stages[1:] {
		result = fmt.Sprintf("%s | %s", result, strings.TrimSpace(stage))
	}
	return fmt.Sprintf("=> pipeline: %s", result), nil
}

func (s *Session) listBindings() string {
	if len(s.Bindings) == 0 {
		return "no bindings"
	}
	var b strings.Builder
	for k, v := range s.Bindings {
		b.WriteString(fmt.Sprintf("  %s = %v\n", k, v))
	}
	return b.String()
}

func (s *Session) help() string {
	return `PSL REPL commands:
  let <name> = <expr>   — bind a variable
  <expr> | <expr>       — pipeline operation
  bindings              — list variables
  history               — show command history
  help                  — show this help`
}

// Completions returns auto-complete suggestions.
func (s *Session) Completions(prefix string) []string {
	keywords := []string{"let", "decode", "encode", "filter", "show", "help", "history", "bindings"}
	var matches []string
	for _, kw := range keywords {
		if strings.HasPrefix(kw, prefix) {
			matches = append(matches, kw)
		}
	}
	for name := range s.Bindings {
		if strings.HasPrefix(name, prefix) {
			matches = append(matches, name)
		}
	}
	return matches
}
