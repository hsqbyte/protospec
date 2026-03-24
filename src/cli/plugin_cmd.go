package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/plugin"
)

func runInstall(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl install <source>")
	}
	if err := plugin.Install(args[0]); err != nil {
		return err
	}
	fmt.Printf("installed %s\n", args[0])
	return nil
}

func runUninstall(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl uninstall <name>")
	}
	if err := plugin.Uninstall(args[0]); err != nil {
		return err
	}
	fmt.Printf("uninstalled %s\n", args[0])
	return nil
}

func runPluginList(ctx *Context) error {
	names, err := plugin.ListInstalled()
	if err != nil {
		return err
	}
	if len(names) == 0 {
		fmt.Println("no plugins installed")
		return nil
	}
	fmt.Println("Installed plugins:")
	for _, n := range names {
		fmt.Printf("  %s\n", n)
	}
	return nil
}

func runInitPackage(ctx *Context, args []string) error {
	name := "my-protocol"
	dir := "."
	if len(args) > 0 {
		name = args[0]
		dir = args[0]
	}
	if err := plugin.InitPackage(dir, name); err != nil {
		return err
	}
	fmt.Printf("initialized package %q in %s/\n", name, dir)
	return nil
}
