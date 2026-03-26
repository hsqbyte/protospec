package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hsqbyte/protospec/src/platform/sdk"
)

func runSDK(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl sdk <command>\n  commands: header, python, rust, typescript, cmake, scapy")
	}

	outDir := "."
	for i := 0; i < len(args); i++ {
		if args[i] == "-o" && i+1 < len(args) {
			outDir = args[i+1]
			args = append(args[:i], args[i+2:]...)
			break
		}
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	cmd := args[0]
	switch cmd {
	case "header":
		return writeFile(outDir, "psl.h", sdk.GenerateCHeader())
	case "python":
		return writeFile(outDir, "protospec.py", sdk.GeneratePythonBinding())
	case "rust":
		return writeFile(outDir, "psl.rs", sdk.GenerateRustBinding())
	case "typescript", "ts":
		return writeFile(outDir, "psl.ts", sdk.GenerateTypeScriptBinding())
	case "cmake":
		return writeFile(outDir, "CMakeLists.txt", sdk.GenerateCMakeExample())
	case "scapy":
		return writeFile(outDir, "psl_scapy.py", sdk.GenerateScapyInterop())
	case "all":
		files := map[string]string{
			"psl.h":          sdk.GenerateCHeader(),
			"protospec.py":   sdk.GeneratePythonBinding(),
			"psl.rs":         sdk.GenerateRustBinding(),
			"psl.ts":         sdk.GenerateTypeScriptBinding(),
			"CMakeLists.txt": sdk.GenerateCMakeExample(),
			"psl_scapy.py":   sdk.GenerateScapyInterop(),
		}
		for name, content := range files {
			if err := writeFile(outDir, name, content); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("unknown sdk command: %s", cmd)
	}
}

func writeFile(dir, name, content string) error {
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	fmt.Printf("generated %s\n", path)
	return nil
}
