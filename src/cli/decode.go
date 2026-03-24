package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func runDecode(ctx *Context, name string) error {
	data, err := readInput(ctx.InputFile)
	if err != nil {
		return err
	}

	result, err := ctx.Lib.Decode(name, data)
	if err != nil {
		return err
	}

	output := map[string]any{
		"fields":     result.Packet,
		"bytes_read": result.BytesRead,
	}

	out, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}

	return writeOutput(ctx.OutputFile, append(out, '\n'))
}

func readInput(file string) ([]byte, error) {
	if file != "" {
		return os.ReadFile(file)
	}
	return io.ReadAll(os.Stdin)
}

func writeOutput(file string, data []byte) error {
	if file != "" {
		return os.WriteFile(file, data, 0644)
	}
	_, err := os.Stdout.Write(data)
	return err
}
