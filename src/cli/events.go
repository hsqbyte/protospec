package cli

import (
	"fmt"

	"github.com/hsqbyte/protospec/src/platform/events"
)

func runEvents(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl events <listen|replay|webhook>")
	}
	bus := events.NewEventBus()
	switch args[0] {
	case "listen":
		fmt.Println("listening for protocol events... (Ctrl+C to stop)")
		_ = bus.Subscribe("cli")
	case "replay":
		evts := bus.Replay(nil)
		for _, e := range evts {
			fmt.Println(events.FormatEvent(e))
		}
		if len(evts) == 0 {
			fmt.Println("no events recorded")
		}
	case "webhook":
		if len(args) < 2 {
			return fmt.Errorf("usage: psl events webhook <url>")
		}
		bus.AddWebhook(events.Webhook{URL: args[1]})
		fmt.Printf("webhook registered: %s\n", args[1])
	default:
		return fmt.Errorf("unknown events subcommand: %s", args[0])
	}
	return nil
}
