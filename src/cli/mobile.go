package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/platform/mobile"
)

func runMobile(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl mobile <ios|android|sdk>")
	}
	switch args[0] {
	case "ios":
		cfg := mobile.NewSDKConfig(mobile.PlatformIOS)
		fmt.Print(cfg.Describe())
		fmt.Println("\n--- Swift SDK ---")
		fmt.Print(mobile.GenerateSwiftSDK())
	case "android":
		cfg := mobile.NewSDKConfig(mobile.PlatformAndroid)
		fmt.Print(cfg.Describe())
		fmt.Println("\n--- Kotlin SDK ---")
		fmt.Print(mobile.GenerateKotlinSDK())
	case "sdk":
		platform := mobile.PlatformIOS
		if len(args) >= 2 && args[1] == "android" {
			platform = mobile.PlatformAndroid
		}
		cfg := mobile.NewSDKConfig(platform)
		fmt.Print(cfg.Describe())
	default:
		return fmt.Errorf("unknown mobile subcommand: %s", args[0])
	}
	return nil
}
