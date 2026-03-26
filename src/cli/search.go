package cli

import (
	"fmt"
	"strings"

	"github.com/hsqbyte/protospec/src/tools/search"
)

func runSearch(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl search <query>")
	}
	idx := search.NewIndex()
	idx.Add(search.Document{Protocol: "IPv4", Layer: "network", Tags: []string{"ip", "network"}, Fields: map[string]string{"src_ip": "bytes", "dst_ip": "bytes"}})
	idx.Add(search.Document{Protocol: "TCP", Layer: "transport", Tags: []string{"tcp", "transport"}, Fields: map[string]string{"src_port": "uint16", "dst_port": "uint16"}})
	idx.Add(search.Document{Protocol: "DNS", Layer: "application", Tags: []string{"dns", "name"}, Fields: map[string]string{"qname": "string", "qtype": "uint16"}})
	results := idx.Search(strings.Join(args, " "))
	fmt.Print(search.FormatResults(results))
	return nil
}
