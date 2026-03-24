package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hsqbyte/protospec/src/cli"
	"github.com/hsqbyte/protospec/src/i18n"
	"github.com/hsqbyte/protospec/src/lsp"
	"github.com/hsqbyte/protospec/src/protocol"
)

const Version = "1.0.0"

var bundle *i18n.Bundle

func init() {
	bundle = i18n.NewBundle()
	if err := bundle.LoadFS(i18n.CliFS, "cli"); err != nil {
		panic("failed to load i18n: " + err.Error())
	}
}

func main() {
	opts, cmdArgs := parseArgs(os.Args[1:])

	if len(cmdArgs) == 0 {
		fmt.Fprint(os.Stderr, bundle.Get(opts.Lang, "usage"))
		os.Exit(1)
	}

	// LSP server runs standalone without library
	if cmdArgs[0] == "lsp" {
		server := lsp.NewServer()
		if err := server.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "lsp error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	lib, err := protocol.NewLibrary()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	for _, f := range opts.LoadFiles {
		if err := lib.LoadPSL(f); err != nil {
			fmt.Fprintf(os.Stderr, "error loading %s: %v\n", f, err)
			os.Exit(1)
		}
	}

	ctx := &cli.Context{
		Lib:        lib,
		InputFile:  opts.InputFile,
		OutputFile: opts.OutputFile,
		Lang:       opts.Lang,
		Raw:        opts.Raw,
	}

	cmd := cmdArgs[0]
	args := cmdArgs[1:]

	if err := cli.Run(ctx, cmd, args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

type options struct {
	LoadFiles  []string
	InputFile  string
	OutputFile string
	Lang       string
	Raw        bool
}

func detectLang() string {
	// 1. Environment variable
	if lang := os.Getenv("PSL_LANG"); lang != "" {
		return lang
	}

	// 2. Config file ~/.psl/config.json
	home, err := os.UserHomeDir()
	if err == nil {
		data, err := os.ReadFile(home + "/.psl/config.json")
		if err == nil {
			var cfg map[string]string
			if json.Unmarshal(data, &cfg) == nil {
				if lang, ok := cfg["lang"]; ok && lang != "" {
					return lang
				}
			}
		}
	}

	// 3. Default
	return "en"
}

func parseArgs(args []string) (options, []string) {
	var opts options
	opts.Lang = detectLang()
	var rest []string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-l":
			i++
			if i < len(args) {
				opts.LoadFiles = append(opts.LoadFiles, args[i])
			}
		case "-f":
			i++
			if i < len(args) {
				opts.InputFile = args[i]
			}
		case "-o":
			i++
			if i < len(args) {
				opts.OutputFile = args[i]
			}
		case "--lang":
			i++
			if i < len(args) {
				opts.Lang = args[i]
			}
		case "-h", "--help":
			fmt.Print(bundle.Get(opts.Lang, "usage"))
			os.Exit(0)
		case "-v", "--version":
			fmt.Printf("psl %s\n", Version)
			os.Exit(0)
		case "--raw":
			opts.Raw = true
		default:
			rest = append(rest, args[i])
		}
	}
	return opts, rest
}
