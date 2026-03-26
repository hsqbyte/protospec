package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hsqbyte/protospec/src/platform/convert"
)

func runConvert(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl convert <protocol> --to <format> [-o file]\n  formats: protobuf, flatbuffers, asn1, json-schema, openapi")
	}

	name := args[0]
	var toFmt, outFile string

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--to":
			i++
			if i < len(args) {
				toFmt = args[i]
			}
		case "-o":
			i++
			if i < len(args) {
				outFile = args[i]
			}
		}
	}

	if toFmt == "" {
		return fmt.Errorf("--to format is required")
	}

	conv := convert.NewConverter(ctx.Lib)
	var result string
	var err error

	switch toFmt {
	case "protobuf", "proto":
		result, err = conv.ToProtobuf(name)
	case "flatbuffers", "fbs":
		result, err = conv.ToFlatBuffers(name)
	case "asn1":
		result, err = conv.ToASN1(name)
	case "json-schema":
		result, err = conv.ToJSONSchema(name)
	case "openapi":
		result, err = conv.ToOpenAPI(name)
	default:
		return fmt.Errorf("unsupported format: %s", toFmt)
	}

	if err != nil {
		return err
	}

	if outFile != "" {
		if dir := filepath.Dir(outFile); dir != "." {
			os.MkdirAll(dir, 0o755)
		}
		if err := os.WriteFile(outFile, []byte(result), 0o644); err != nil {
			return err
		}
		fmt.Printf("converted %s → %s\n", name, outFile)
	} else {
		fmt.Print(result)
	}
	return nil
}

func runImport(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl import --from <format> <file>\n  formats: protobuf, json-schema")
	}

	var fromFmt, inputFile string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--from":
			i++
			if i < len(args) {
				fromFmt = args[i]
			}
		default:
			if inputFile == "" {
				inputFile = args[i]
			}
		}
	}

	if fromFmt == "" {
		return fmt.Errorf("--from format is required")
	}
	if inputFile == "" {
		return fmt.Errorf("input file is required")
	}

	data, err := os.ReadFile(inputFile)
	if err != nil {
		return err
	}

	var result string
	switch fromFmt {
	case "protobuf", "proto":
		result, err = convert.ImportFromProtobuf(string(data))
	case "json-schema":
		result, err = convert.ImportFromJSONSchema(string(data))
	default:
		return fmt.Errorf("unsupported import format: %s", fromFmt)
	}

	if err != nil {
		return err
	}

	fmt.Print(result)
	return nil
}
