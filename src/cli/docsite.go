package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/docsite"
)

func runDocsite(ctx *Context, args []string) error {
	outDir := "./docs-site"
	lang := ctx.Lang
	title := "PSL Protocol Documentation"

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-o":
			i++
			if i < len(args) {
				outDir = args[i]
			}
		case "--lang":
			i++
			if i < len(args) {
				lang = args[i]
			}
		case "--title":
			i++
			if i < len(args) {
				title = args[i]
			}
		case "--json":
			return docsite.GenerateJSON(ctx.Lib, outDir)
		}
	}

	cfg := &docsite.Config{
		Title:  title,
		Lang:   lang,
		OutDir: outDir,
		Lib:    ctx.Lib,
	}

	if err := docsite.Generate(cfg); err != nil {
		return err
	}

	fmt.Printf("documentation site generated in %s/\n", outDir)
	return nil
}
