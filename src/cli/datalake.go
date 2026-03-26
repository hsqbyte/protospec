package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/platform/datalake"
)

func runDatalake(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl datalake <schema|sql|query> <protocol>")
	}
	proto := "IPv4"
	if len(args) >= 2 {
		proto = args[1]
	}
	fields := map[string]string{"version": "uint4", "ihl": "uint4", "total_length": "uint16", "src_ip": "bytes", "dst_ip": "bytes"}
	switch args[0] {
	case "schema":
		s := datalake.DeriveSchema(proto, fields)
		fmt.Print(s.Describe())
	case "sql":
		s := datalake.DeriveSchema(proto, fields)
		fmt.Print(s.GenerateSQL(proto + "_packets"))
	case "query":
		fmt.Println(datalake.GenerateDuckDBQuery(proto+"_packets", proto))
	default:
		return fmt.Errorf("unknown datalake subcommand: %s", args[0])
	}
	return nil
}
