package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/knowledge"
)

func runKnowledge(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl knowledge <graph|query> [protocol]")
	}
	g := knowledge.NewGraph()
	g.AddNode("tcp", "protocol", "TCP")
	g.AddNode("ip", "protocol", "IPv4")
	g.AddNode("rfc793", "rfc", "RFC 793")
	g.AddNode("ietf", "org", "IETF")
	g.AddEdge("ip", "tcp", knowledge.RelEncapsulates)
	g.AddEdge("tcp", "rfc793", knowledge.RelDefinedBy)
	g.AddEdge("rfc793", "ietf", knowledge.RelDefinedBy)

	switch args[0] {
	case "graph":
		fmt.Print(g.ToMermaid())
	case "query":
		node := "tcp"
		if len(args) >= 2 {
			node = args[1]
		}
		edges := g.Query(node)
		for _, e := range edges {
			fmt.Printf("  %s -[%s]-> %s\n", e.From, e.Relation, e.To)
		}
	case "describe":
		fmt.Print(g.Describe())
	default:
		return fmt.Errorf("unknown knowledge subcommand: %s", args[0])
	}
	return nil
}
