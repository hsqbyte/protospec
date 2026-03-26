import "./wasm_exec.js";

let _initialized = false;
let _initPromise = null;

/**
 * Initialize the protospec WASM module.
 * Call this once before using any other functions.
 * @param {string | URL | undefined} wasmUrl - URL to protospec.wasm (defaults to bundled file)
 */
export async function init(wasmUrl) {
  if (_initialized) return;
  if (_initPromise) return _initPromise;

  _initPromise = (async () => {
    const go = new globalThis.Go();

    let result;
    if (typeof wasmUrl === "string" || wasmUrl instanceof URL) {
      const resp = await fetch(wasmUrl);
      result = await WebAssembly.instantiateStreaming(resp, go.importObject);
    } else {
      // Node.js or bundler: load from file path
      const url = new URL("protospec.wasm", import.meta.url);
      const resp = await fetch(url);
      result = await WebAssembly.instantiateStreaming(resp, go.importObject);
    }

    go.run(result.instance); // non-blocking, Go blocks in select{}
    _initialized = true;
  })();

  return _initPromise;
}

function call(fn, ...args) {
  if (!_initialized) throw new Error("protospec-wasm not initialized. Call init() first.");
  const raw = globalThis[fn](...args);
  return JSON.parse(raw);
}

/** List all available protocols with metadata */
export function listProtocols() {
  return call("psl_listProtocols");
}

/** Get full protocol schema by name */
export function getProtocol(name) {
  return call("psl_getProtocol", name);
}

/** Get protocol metadata by name */
export function getMeta(name) {
  return call("psl_getMeta", name);
}

/**
 * Encode JSON field values to hex
 * @param {string} protocolName
 * @param {object} fields - field values as JSON object
 * @returns {{ hex: string } | { error: string }}
 */
export function encode(protocolName, fields) {
  return call("psl_encode", protocolName, JSON.stringify(fields));
}

/**
 * Decode hex string to JSON field values
 * @param {string} protocolName
 * @param {string} hexString - hex encoded binary data
 * @returns {object | { error: string }}
 */
export function decode(protocolName, hexString) {
  return call("psl_decode", protocolName, hexString);
}
