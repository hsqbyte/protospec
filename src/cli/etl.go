package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/etl"
)

func runETL(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl etl <run|describe> [options]")
	}
	switch args[0] {
	case "run":
		p := etl.NewPipeline("default").
			SetSource(etl.SourcePCAP, "input.pcap").
			AddTransform(etl.TransformOp{Type: "decode", Config: "auto"}).
			SetSink(etl.SinkFile, "output.json")
		return p.Run()
	case "describe":
		p := etl.NewPipeline("example").
			SetSource(etl.SourceKafka, "localhost:9092/packets").
			AddTransform(etl.TransformOp{Type: "decode", Config: "auto"}).
			AddTransform(etl.TransformOp{Type: "filter", Config: "protocol=tcp"}).
			SetSink(etl.SinkElasticsearch, "http://localhost:9200/packets")
		fmt.Print(p.Describe())
	default:
		return fmt.Errorf("unknown etl subcommand: %s", args[0])
	}
	return nil
}
