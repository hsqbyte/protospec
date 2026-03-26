package cli

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/hsqbyte/protospec/src/core/schema"
)

func runBench(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl bench <protocol> [-n iterations]")
	}

	name := args[0]
	iterations := 10000

	for i := 1; i < len(args); i++ {
		if args[i] == "-n" && i+1 < len(args) {
			fmt.Sscanf(args[i+1], "%d", &iterations)
			i++
		}
	}

	s, err := ctx.Lib.Registry().GetSchema(name)
	if err != nil {
		return err
	}

	// Generate test data
	var fields []schema.FieldDef
	for _, f := range s.Fields {
		if f.IsBitfieldGroup {
			fields = append(fields, f.BitfieldFields...)
		} else {
			fields = append(fields, f)
		}
	}

	testData := generateFuzzData(s, false)

	// Benchmark decode
	start := time.Now()
	for i := 0; i < iterations; i++ {
		ctx.Lib.Decode(name, testData)
	}
	decodeTime := time.Since(start)

	// Benchmark encode
	result, decErr := ctx.Lib.Decode(name, testData)
	var packet map[string]any
	if decErr != nil {
		packet = make(map[string]any)
		for _, f := range fields {
			if f.Type == schema.Uint || f.Type == schema.Int {
				packet[f.Name] = uint64(rand.Intn(256))
			}
		}
	} else {
		packet = result.Packet
	}

	start = time.Now()
	for i := 0; i < iterations; i++ {
		ctx.Lib.Encode(name, packet)
	}
	encodeTime := time.Since(start)

	// Results
	fmt.Printf("%sBenchmark: %s%s (%d iterations)\n\n", cBold, name, cReset, iterations)
	fmt.Printf("  Decode: %s%v%s total, %s%v%s/op\n",
		cGreen, decodeTime, cReset,
		cCyan, decodeTime/time.Duration(iterations), cReset)
	fmt.Printf("  Encode: %s%v%s total, %s%v%s/op\n",
		cGreen, encodeTime, cReset,
		cCyan, encodeTime/time.Duration(iterations), cReset)
	fmt.Printf("  Data:   %d bytes\n", len(testData))

	decodeOps := float64(iterations) / decodeTime.Seconds()
	encodeOps := float64(iterations) / encodeTime.Seconds()
	fmt.Printf("\n  Decode throughput: %s%.0f ops/s%s\n", cGreen, decodeOps, cReset)
	fmt.Printf("  Encode throughput: %s%.0f ops/s%s\n", cGreen, encodeOps, cReset)

	return nil
}
