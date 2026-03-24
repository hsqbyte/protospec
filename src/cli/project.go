package cli

import (
	"fmt"
	"os"

	"github.com/hsqbyte/protospec/src/project"
)

func runProjectInit(ctx *Context, args []string) error {
	name := "my-protocols"
	dir := "."
	if len(args) > 0 {
		name = args[0]
		dir = args[0]
	}

	if err := project.Init(dir, name); err != nil {
		return err
	}
	fmt.Printf("Initialized PSL project: %s\n", name)
	return nil
}

func runGitHook(ctx *Context, args []string) error {
	dir := "."
	if len(args) > 0 {
		dir = args[0]
	}

	if err := project.InstallGitHook(dir); err != nil {
		return err
	}
	fmt.Println("Installed PSL pre-commit hook.")
	return nil
}

func runSign(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl sign <file.psl>")
	}

	sig, err := project.Sign(args[0])
	if err != nil {
		return err
	}

	fmt.Printf("%s  %s\n", sig, args[0])

	if ctx.OutputFile != "" {
		return os.WriteFile(ctx.OutputFile+".sig", []byte(sig+"\n"), 0644)
	}
	return nil
}
