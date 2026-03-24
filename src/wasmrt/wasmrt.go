// Package wasmrt provides WASM runtime configuration for PSL engine.
package wasmrt

// Config holds WASM compilation configuration.
type Config struct {
	Target    string   `json:"target"`     // "wasm", "wasi"
	OptLevel  int      `json:"opt_level"`  // 0-3
	MaxMemory int      `json:"max_memory"` // MB
	Features  []string `json:"features"`
}

// DefaultConfig returns default WASM config.
func DefaultConfig() *Config {
	return &Config{
		Target:    "wasm",
		OptLevel:  2,
		MaxMemory: 64,
		Features:  []string{"decode", "encode", "validate"},
	}
}

// ExportedFunctions lists functions exported to WASM.
var ExportedFunctions = []string{
	"psl_decode",
	"psl_encode",
	"psl_validate",
	"psl_list_protocols",
	"psl_get_schema",
	"psl_alloc",
	"psl_free",
}
