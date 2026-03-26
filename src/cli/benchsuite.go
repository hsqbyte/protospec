package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/tools/benchsuite"
)

func runBenchSuite(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl benchsuite <protocol> [iterations]")
	}

	protoName := args[0]
	iterations := 10000

	suite := benchsuite.NewSuite(ctx.Lib, iterations)

	// Create sample data for decode benchmark
	samplePacket := map[string]any{}
	encoded, err := ctx.Lib.Encode(protoName, samplePacket)
	if err != nil {
		fmt.Printf("Skipping encode/decode bench (encode failed: %v)\n", err)
		return nil
	}

	var results []*benchsuite.BenchResult

	encResult, err := suite.RunEncode(protoName, samplePacket)
	if err == nil {
		results = append(results, encResult)
	}

	decResult, err := suite.RunDecode(protoName, encoded)
	if err == nil {
		results = append(results, decResult)
	}

	fmt.Print(benchsuite.FormatResults(results))
	return nil
}
