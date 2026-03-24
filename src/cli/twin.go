package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/twin"
)

func runTwin(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl twin <model|whatif|visualize>")
	}
	m := twin.NewModel("lab-network")
	m.AddNode("router1", "router", []string{"IPv4", "OSPF"})
	m.AddNode("switch1", "switch", []string{"Ethernet", "VLAN"})
	m.AddNode("host1", "host", []string{"TCP", "HTTP"})
	m.AddLink("router1", "switch1", "10Gbps", "1ms")
	m.AddLink("switch1", "host1", "1Gbps", "0.5ms")

	switch args[0] {
	case "model":
		fmt.Print(m.Describe())
	case "whatif":
		param, value := "latency", "10ms"
		if len(args) >= 3 {
			param, value = args[1], args[2]
		}
		fmt.Print(m.WhatIf(param, value))
	case "visualize":
		fmt.Print(m.ToMermaid())
	default:
		return fmt.Errorf("unknown twin subcommand: %s", args[0])
	}
	return nil
}
