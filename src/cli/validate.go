package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/hsqbyte/protospec/psl"
	"github.com/hsqbyte/protospec/src/core/pdl"
)

func runValidate(ctx *Context, file string) error {
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	source := string(data)

	// Detect message vs binary protocol by checking first keyword
	trimmed := strings.TrimSpace(source)
	if strings.HasPrefix(trimmed, "message ") || strings.HasPrefix(trimmed, "message\t") {
		parser := pdl.NewPDLParser(nil, nil)
		loader := pdl.NewTransportLoader(psl.FS, nil)
		parser.SetTransportLoader(loader)
		_, err = parser.ParseMessage(source)
	} else {
		_, err = ctx.Lib.CreateCodec(source)
	}

	if err != nil {
		return fmt.Errorf("%s: %w", file, err)
	}

	fmt.Printf(msg(ctx, "validate.ok")+"\n", file)
	return nil
}
