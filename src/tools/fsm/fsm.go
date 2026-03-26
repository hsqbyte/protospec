// Package fsm provides protocol state machine extraction and visualization.
package fsm

import (
	"fmt"
	"strings"
)

// State represents a protocol state.
type State struct {
	Name    string `json:"name"`
	Initial bool   `json:"initial,omitempty"`
	Final   bool   `json:"final,omitempty"`
}

// Transition represents a state transition.
type Transition struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Event     string `json:"event"`
	Condition string `json:"condition,omitempty"`
}

// StateMachine represents a protocol state machine.
type StateMachine struct {
	Protocol    string       `json:"protocol"`
	States      []State      `json:"states"`
	Transitions []Transition `json:"transitions"`
}

// NewStateMachine creates a new state machine.
func NewStateMachine(protocol string) *StateMachine {
	return &StateMachine{Protocol: protocol}
}

// AddState adds a state.
func (sm *StateMachine) AddState(name string, initial, final bool) {
	sm.States = append(sm.States, State{Name: name, Initial: initial, Final: final})
}

// AddTransition adds a transition.
func (sm *StateMachine) AddTransition(from, to, event string) {
	sm.Transitions = append(sm.Transitions, Transition{From: from, To: to, Event: event})
}

// ToMermaid generates a Mermaid state diagram.
func (sm *StateMachine) ToMermaid() string {
	var b strings.Builder
	b.WriteString("stateDiagram-v2\n")
	for _, s := range sm.States {
		if s.Initial {
			b.WriteString(fmt.Sprintf("  [*] --> %s\n", s.Name))
		}
	}
	for _, t := range sm.Transitions {
		b.WriteString(fmt.Sprintf("  %s --> %s : %s\n", t.From, t.To, t.Event))
	}
	for _, s := range sm.States {
		if s.Final {
			b.WriteString(fmt.Sprintf("  %s --> [*]\n", s.Name))
		}
	}
	return b.String()
}

// ToDOT generates a Graphviz DOT diagram.
func (sm *StateMachine) ToDOT() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("digraph %s {\n  rankdir=LR;\n", sm.Protocol))
	for _, s := range sm.States {
		shape := "circle"
		if s.Final {
			shape = "doublecircle"
		}
		b.WriteString(fmt.Sprintf("  %s [shape=%s];\n", s.Name, shape))
	}
	for _, t := range sm.Transitions {
		b.WriteString(fmt.Sprintf("  %s -> %s [label=\"%s\"];\n", t.From, t.To, t.Event))
	}
	b.WriteString("}\n")
	return b.String()
}

// Coverage calculates state coverage from observed transitions.
func (sm *StateMachine) Coverage(observed []string) float64 {
	seen := make(map[string]bool)
	for _, s := range observed {
		seen[s] = true
	}
	if len(sm.States) == 0 {
		return 0
	}
	covered := 0
	for _, s := range sm.States {
		if seen[s.Name] {
			covered++
		}
	}
	return float64(covered) / float64(len(sm.States)) * 100
}
