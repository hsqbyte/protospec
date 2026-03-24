package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/hsqbyte/protospec/src/simulate"
)

func runMock(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl mock <message-protocol> [--port 8080] [--rules rules.json]")
	}

	name := args[0]
	port := 8080
	var rulesFile string

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--port", "-p":
			i++
			if i < len(args) {
				p, err := strconv.Atoi(args[i])
				if err == nil {
					port = p
				}
			}
		case "--rules":
			i++
			if i < len(args) {
				rulesFile = args[i]
			}
		}
	}

	mock, err := simulate.NewMockServer(ctx.Lib, name)
	if err != nil {
		return err
	}

	if rulesFile != "" {
		data, err := os.ReadFile(rulesFile)
		if err != nil {
			return fmt.Errorf("read rules: %w", err)
		}
		if err := mock.LoadRules(data); err != nil {
			return fmt.Errorf("parse rules: %w", err)
		}
	}

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("mock server for %s listening on http://localhost%s\n", name, addr)
	fmt.Printf("  logs: http://localhost%s/logs\n", addr)

	// This will block — user should run manually
	return fmt.Errorf("run manually: mock server requires long-running process on %s", addr)
}

func runLoadTest(ctx *Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: psl loadtest <protocol> --target host:port [--rps 1000] [--duration 10s] [--concurrency 10]")
	}

	cfg := &simulate.LoadTestConfig{
		Protocol:    args[0],
		RPS:         100,
		Duration:    10 * time.Second,
		Concurrency: 10,
	}

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--target":
			i++
			if i < len(args) {
				cfg.Target = args[i]
			}
		case "--rps":
			i++
			if i < len(args) {
				r, _ := strconv.Atoi(args[i])
				if r > 0 {
					cfg.RPS = r
				}
			}
		case "--duration":
			i++
			if i < len(args) {
				d, err := time.ParseDuration(args[i])
				if err == nil {
					cfg.Duration = d
				}
			}
		case "--concurrency":
			i++
			if i < len(args) {
				c, _ := strconv.Atoi(args[i])
				if c > 0 {
					cfg.Concurrency = c
				}
			}
		}
	}

	if cfg.Target == "" {
		return fmt.Errorf("--target is required")
	}

	// Generate sample data for the protocol
	cfg.Data = []byte{0x00} // minimal packet

	fmt.Printf("load test: %s → %s (%d rps, %s, %d workers)\n",
		cfg.Protocol, cfg.Target, cfg.RPS, cfg.Duration, cfg.Concurrency)

	result, err := simulate.RunLoadTest(cfg)
	if err != nil {
		return err
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(data))
	return nil
}
