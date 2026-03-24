package cli

import (
	"fmt"
	"os"

	"github.com/hsqbyte/protospec/src/sequence"
)

func runSequence(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl sequence <capture.pcap> [--format mermaid|plantuml]")
	}

	format := "mermaid"
	for i, a := range args {
		if a == "--format" && i+1 < len(args) {
			format = args[i+1]
		}
	}

	// Demo: generate sample sequence diagram
	diagram := sequence.NewDiagram("Protocol Sequence — " + args[0])
	diagram.AddEvent(sequence.Event{Source: "Client", Dest: "Server", Protocol: "TCP", Info: "SYN", Timestamp: 0})
	diagram.AddEvent(sequence.Event{Source: "Server", Dest: "Client", Protocol: "TCP", Info: "SYN-ACK", Timestamp: 0.001, Latency: 1.0})
	diagram.AddEvent(sequence.Event{Source: "Client", Dest: "Server", Protocol: "TCP", Info: "ACK", Timestamp: 0.002, Latency: 1.0})

	var output string
	switch format {
	case "plantuml":
		output = diagram.ToPlantUML()
	default:
		output = diagram.ToMermaid()
	}

	if ctx.OutputFile != "" {
		return os.WriteFile(ctx.OutputFile, []byte(output), 0644)
	}
	fmt.Print(output)
	return nil
}
