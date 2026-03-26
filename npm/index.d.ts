export interface ProtocolMeta {
  name: string;
  rfc?: string;
  title?: Record<string, string>;
  description?: Record<string, string>;
  layer?: string;
  type?: string;
  depends_on?: string[];
  status?: string;
}

export interface EncodeResult {
  hex?: string;
  error?: string;
}

export interface DecodeResult {
  fields?: Record<string, any>;
  bytes_read?: number;
  error?: string;
}

/** Initialize the WASM module. Must be called before any other function. */
export function init(wasmUrl?: string | URL): Promise<void>;

/** List all available protocols with metadata */
export function listProtocols(): ProtocolMeta[];

/** Get full protocol schema by name */
export function getProtocol(name: string): any;

/** Get protocol metadata by name */
export function getMeta(name: string): ProtocolMeta | { error: string };

/** Encode JSON field values to hex */
export function encode(protocolName: string, fields: Record<string, any>): EncodeResult;

/** Decode hex string to JSON field values */
export function decode(protocolName: string, hexString: string): DecodeResult;
