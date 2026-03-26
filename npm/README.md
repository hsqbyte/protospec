# @hsqbyte/protospec-wasm

PSL (Protocol Specification Language) engine compiled to WebAssembly.

Browse, encode, and decode binary protocol packets directly in the browser — no backend required.

## Install

```bash
npm install @hsqbyte/protospec-wasm
```

## Usage

```js
import { init, listProtocols, getProtocol, encode, decode } from "@hsqbyte/protospec-wasm";

// Initialize WASM (call once)
await init();

// List all protocols
const protocols = listProtocols();
console.log(protocols);

// Get protocol schema
const tcp = getProtocol("TCP");

// Encode
const result = encode("UDP", {
  source_port: 12345,
  destination_port: 53,
  length: 20,
  checksum: 0
});
console.log(result.hex);

// Decode
const decoded = decode("UDP", result.hex);
console.log(decoded.fields);
```

## Vite Configuration

For Vite projects, configure `wasm` file handling in `vite.config.ts`:

```ts
export default defineConfig({
  assetsInclude: ["**/*.wasm"],
});
```

## Related

- [protospec](https://github.com/hsqbyte/protospec) — PSL engine and CLI
- [psl_ui](https://github.com/hsqbyte/psl_ui) — Web UI for PSL

## License

GPL-3.0
