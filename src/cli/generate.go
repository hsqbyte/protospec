package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hsqbyte/protospec/src/codegen"
)

func runGenerate(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl generate <lang> [protocol|--all] [-o dir] [--pkg name]")
	}

	lang := args[0]
	rest := args[1:]

	var (
		protocols    []string
		all          bool
		outDir       = "."
		pkgName      = ""
		templateFile = ""
	)

	for i := 0; i < len(rest); i++ {
		switch rest[i] {
		case "--all":
			all = true
		case "-o":
			i++
			if i < len(rest) {
				outDir = rest[i]
			}
		case "--pkg":
			i++
			if i < len(rest) {
				pkgName = rest[i]
			}
		case "--template":
			i++
			if i < len(rest) {
				templateFile = rest[i]
			}
		default:
			protocols = append(protocols, rest[i])
		}
	}

	if !all && len(protocols) == 0 {
		return fmt.Errorf("specify a protocol name or use --all")
	}

	if all {
		protocols = ctx.Lib.AllNames()
	}

	if pkgName == "" {
		pkgName = "proto"
	}

	gen := codegen.NewGenerator(ctx.Lib)

	if templateFile != "" {
		return generateFromTemplate(gen, protocols, outDir, templateFile)
	}

	switch lang {
	case "go":
		return generateLangFiles(gen, protocols, outDir, pkgName, lang)
	case "python":
		return generateLangFiles(gen, protocols, outDir, pkgName, lang)
	case "rust":
		return generateLangFiles(gen, protocols, outDir, pkgName, lang)
	case "c":
		return generateLangFiles(gen, protocols, outDir, pkgName, lang)
	case "typescript", "ts":
		return generateLangFiles(gen, protocols, outDir, pkgName, "typescript")
	case "wireshark", "lua":
		return generateLangFiles(gen, protocols, outDir, pkgName, "wireshark")
	default:
		return fmt.Errorf("unsupported language: %s (supported: go, python, rust, c, typescript)", lang)
	}
}

func generateLangFiles(gen *codegen.Generator, protocols []string, outDir, pkgName, lang string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	extMap := map[string]string{
		"go":         ".go",
		"python":     ".py",
		"rust":       ".rs",
		"c":          ".h",
		"typescript": ".ts",
		"wireshark":  ".lua",
	}
	ext := extMap[lang]

	for _, name := range protocols {
		var code string
		var err error
		switch lang {
		case "go":
			code, err = gen.GenerateGo(name, pkgName)
		case "python":
			code, err = gen.GeneratePython(name)
		case "rust":
			code, err = gen.GenerateRust(name)
		case "c":
			code, err = gen.GenerateC(name)
		case "typescript":
			code, err = gen.GenerateTypeScript(name)
		case "wireshark":
			code, err = gen.GenerateWireshark(name)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: skip %s: %v\n", name, err)
			continue
		}

		filename := strings.ToLower(name) + ext
		path := filepath.Join(outDir, filename)
		if err := os.WriteFile(path, []byte(code), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
		fmt.Printf("generated %s\n", path)
	}

	return nil
}

func generateFromTemplate(gen *codegen.Generator, protocols []string, outDir, templateFile string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	for _, name := range protocols {
		code, err := gen.GenerateFromTemplate(name, templateFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: skip %s: %v\n", name, err)
			continue
		}

		filename := strings.ToLower(name) + ".gen"
		path := filepath.Join(outDir, filename)
		if err := os.WriteFile(path, []byte(code), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
		fmt.Printf("generated %s\n", path)
	}
	return nil
}
