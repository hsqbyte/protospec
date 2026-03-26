<p align="center">
  <img src="assets/psl.svg" width="80" alt="PSL">
</p>

<h1 align="center">PSL — Protocol Specification Language</h1>

<p align="center">
  Define any binary protocol in text, encode and decode automatically.
</p>

<p align="center">
  <a href="README.md">English</a> | <a href="README_ZH.md">中文</a>
</p>

---

PSL is a universal binary protocol codec engine for Go. You describe a protocol's structure in a `.psl` text file — field names, bit widths, byte order, checksums, conditional fields — and the engine handles encoding and decoding automatically. No per-protocol code needed.

Built-in protocols: IPv4, TCP, UDP, ICMP, ARP, DNS, HTTP/1.1, WebSocket, MQTT, TLS, and 40+ more.

```bash
go build -o psl .
```

## Go Integration

Install via [pkg.go.dev](https://pkg.go.dev/github.com/hsqbyte/protospec):

```bash
go get github.com/hsqbyte/protospec@v1.0.0
```

```go
package main

import (
    "fmt"
    "github.com/hsqbyte/protospec/src/protocol"
)

func main() {
    lib, _ := protocol.NewLibrary()

    // List all protocols
    for _, name := range lib.AllNames() {
        fmt.Println(name)
    }

    // Encode
    data, _ := lib.Encode("UDP", map[string]any{
        "source_port": 12345, "destination_port": 53,
        "length": 20, "checksum": 0,
    })
    fmt.Printf("hex: %x\n", data)

    // Decode
    result, _ := lib.Decode("UDP", data)
    fmt.Println(result.Fields)
}
```

## Web / Browser Integration (WASM)

PSL is also available as a WebAssembly module on [npm](https://www.npmjs.com/package/protospec-wasm), so you can run the full protocol engine directly in the browser — no backend required.

```bash
npm install protospec-wasm
```

Copy `protospec.wasm` and `wasm_exec.js` from `node_modules/protospec-wasm/` to your project's public directory, then:

```js
// 1. Load the WASM engine
const go = new Go();  // from wasm_exec.js
const resp = await fetch("/protospec.wasm");
const result = await WebAssembly.instantiateStreaming(resp, go.importObject);
go.run(result.instance);

// 2. List all protocols (returns JSON string)
const protocols = JSON.parse(psl_listProtocols());

// 3. Encode
const encoded = JSON.parse(psl_encode("UDP", JSON.stringify({
  source_port: 12345, destination_port: 53, length: 20, checksum: 0
})));
console.log(encoded.hex);

// 4. Decode
const decoded = JSON.parse(psl_decode("UDP", encoded.hex));
console.log(decoded);
```

### Available WASM Functions

| Function | Args | Returns |
|---|---|---|
| `psl_listProtocols()` | — | JSON array of all protocols with metadata |
| `psl_getProtocol(name)` | protocol name | JSON protocol schema |
| `psl_getMeta(name)` | protocol name | JSON metadata (RFC, description, layer, etc.) |
| `psl_encode(name, json)` | protocol name, JSON fields | `{"hex": "..."}` |
| `psl_decode(name, hex)` | protocol name, hex string | JSON decoded fields |

### Build WASM from Source

```bash
task wasm          # compile → npm/protospec.wasm
task npm:publish   # compile + publish to npm
```

## Web UI

[PSL UI](https://github.com/hsqbyte/psl_ui) provides a web-based interface for browsing, visualizing, and interacting with PSL protocol definitions — powered by the WASM engine above.

<p align="center">
  <img src="assets/psl_ui_screenshot.png" alt="PSL UI">
</p>

## Project Structure

```
cmd/wasm/          WASM entry point
npm/               npm package (protospec-wasm)
psl/               Built-in protocol definitions (.psl + meta.json)
src/
├── core/          Codec engine, parser, schema, checksum, format
├── cli/           CLI commands
├── protocol/      Protocol library API
├── i18n/          Internationalization
├── codegen/       Multi-language code generation
├── lsp/           Language server
├── tools/         Lint, formatter, coverage, testgen, debug, etc.
├── integrations/  eBPF, DPDK, gateway, crypto, compliance, etc.
├── platform/      Cloud, plugin, SDK, CI, migrate, etc.
└── docs/          Documentation generation
```

## License

[GPL-3.0](LICENSE)
