package sdk

// CExport provides C-compatible API functions for building shared libraries.
// Build with: go build -buildmode=c-shared -o libpsl.so ./src/sdk/cexport/

import (
	"encoding/json"
	"unsafe"
)

// --- C-compatible result types ---

// CResult holds a result string and error for C callers.
type CResult struct {
	Data  *byte
	Error *byte
	Len   int
}

// AllocCString allocates a null-terminated C string from a Go string.
func AllocCString(s string) *byte {
	buf := make([]byte, len(s)+1)
	copy(buf, s)
	buf[len(s)] = 0
	return &buf[0]
}

// FreeCString frees a C string allocated by AllocCString.
// In practice, Go's GC handles this, but we provide the API for symmetry.
func FreeCString(p *byte) {
	// no-op: Go GC manages memory
	_ = p
}

// --- C API wrappers ---

// CPSLNew creates a new library and returns the handle.
func CPSLNew() (Handle, *byte) {
	h, err := NewLibrary()
	if err != nil {
		return 0, AllocCString(err.Error())
	}
	return h, nil
}

// CPSLFree releases a library handle.
func CPSLFree(h Handle) {
	FreeLibrary(h)
}

// CPSLDecode decodes binary data and returns JSON result.
func CPSLDecode(h Handle, proto string, data []byte) (string, error) {
	return Decode(h, proto, data)
}

// CPSLEncode encodes JSON fields to binary.
func CPSLEncode(h Handle, proto string, fieldsJSON string) ([]byte, error) {
	return Encode(h, proto, fieldsJSON)
}

// CPSLList returns protocol list as JSON.
func CPSLList(h Handle) (string, error) {
	return ListProtocols(h)
}

// CPSLLoad loads a PSL file.
func CPSLLoad(h Handle, path string) error {
	return LoadPSL(h, path)
}

// --- Header generation ---

// GenerateCHeader generates a C header file for the PSL shared library.
func GenerateCHeader() string {
	return `/* PSL C API — Auto-generated header */
#ifndef PSL_H
#define PSL_H

#include <stdint.h>
#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef uintptr_t psl_handle_t;

/* Create a new PSL library instance. Returns 0 on error. */
psl_handle_t psl_new(char** err_out);

/* Free a PSL library instance. */
void psl_free(psl_handle_t h);

/* Load a PSL file. Returns 0 on success, -1 on error. */
int psl_load(psl_handle_t h, const char* path, char** err_out);

/* Decode binary data. Result is JSON string (caller must free). */
char* psl_decode(psl_handle_t h, const char* protocol,
                 const uint8_t* data, size_t len, char** err_out);

/* Encode from JSON fields. Result buffer (caller must free). */
uint8_t* psl_encode(psl_handle_t h, const char* protocol,
                    const char* fields_json, size_t* out_len, char** err_out);

/* List all protocols. Returns JSON array string (caller must free). */
char* psl_list(psl_handle_t h, char** err_out);

/* Free a string returned by PSL functions. */
void psl_free_string(char* s);

/* Free a buffer returned by PSL functions. */
void psl_free_buffer(uint8_t* buf);

#ifdef __cplusplus
}
#endif

#endif /* PSL_H */
`
}

// GeneratePythonBinding generates a Python ctypes binding module.
func GeneratePythonBinding() string {
	return `"""PSL Python SDK — ctypes binding to libpsl shared library."""

import ctypes
import json
import os
import sys
from pathlib import Path
from typing import Any, Dict, List, Optional

def _find_lib():
    """Find the libpsl shared library."""
    names = {
        "linux": "libpsl.so",
        "darwin": "libpsl.dylib",
        "win32": "psl.dll",
    }
    libname = names.get(sys.platform, "libpsl.so")
    
    # Search paths
    search = [
        Path(__file__).parent / libname,
        Path.cwd() / libname,
        Path(os.environ.get("PSL_LIB_PATH", "")) / libname,
    ]
    for p in search:
        if p.exists():
            return str(p)
    return libname  # Fall back to system search

_lib = ctypes.CDLL(_find_lib())

# Function signatures
_lib.psl_new.restype = ctypes.c_void_p
_lib.psl_new.argtypes = [ctypes.POINTER(ctypes.c_char_p)]

_lib.psl_free.restype = None
_lib.psl_free.argtypes = [ctypes.c_void_p]

_lib.psl_decode.restype = ctypes.c_char_p
_lib.psl_decode.argtypes = [
    ctypes.c_void_p, ctypes.c_char_p,
    ctypes.POINTER(ctypes.c_uint8), ctypes.c_size_t,
    ctypes.POINTER(ctypes.c_char_p),
]

_lib.psl_encode.restype = ctypes.POINTER(ctypes.c_uint8)
_lib.psl_encode.argtypes = [
    ctypes.c_void_p, ctypes.c_char_p, ctypes.c_char_p,
    ctypes.POINTER(ctypes.c_size_t),
    ctypes.POINTER(ctypes.c_char_p),
]

_lib.psl_list.restype = ctypes.c_char_p
_lib.psl_list.argtypes = [ctypes.c_void_p, ctypes.POINTER(ctypes.c_char_p)]

_lib.psl_load.restype = ctypes.c_int
_lib.psl_load.argtypes = [ctypes.c_void_p, ctypes.c_char_p, ctypes.POINTER(ctypes.c_char_p)]

_lib.psl_free_string.restype = None
_lib.psl_free_string.argtypes = [ctypes.c_char_p]


class PSLError(Exception):
    """PSL SDK error."""
    pass


class Protocol:
    """Represents a PSL library instance with Pythonic API."""

    def __init__(self):
        err = ctypes.c_char_p()
        self._handle = _lib.psl_new(ctypes.byref(err))
        if not self._handle:
            raise PSLError(err.value.decode() if err.value else "unknown error")

    def __del__(self):
        if hasattr(self, "_handle") and self._handle:
            _lib.psl_free(self._handle)

    def load(self, path: str) -> None:
        """Load a PSL file."""
        err = ctypes.c_char_p()
        rc = _lib.psl_load(self._handle, path.encode(), ctypes.byref(err))
        if rc != 0:
            raise PSLError(err.value.decode() if err.value else "load failed")

    def decode(self, protocol: str, data: bytes) -> Dict[str, Any]:
        """Decode binary data using the named protocol."""
        err = ctypes.c_char_p()
        buf = (ctypes.c_uint8 * len(data))(*data)
        result = _lib.psl_decode(
            self._handle, protocol.encode(),
            buf, len(data), ctypes.byref(err),
        )
        if not result:
            raise PSLError(err.value.decode() if err.value else "decode failed")
        return json.loads(result.decode())

    def encode(self, protocol: str, fields: Dict[str, Any]) -> bytes:
        """Encode fields into binary using the named protocol."""
        err = ctypes.c_char_p()
        out_len = ctypes.c_size_t()
        fields_json = json.dumps(fields).encode()
        result = _lib.psl_encode(
            self._handle, protocol.encode(), fields_json,
            ctypes.byref(out_len), ctypes.byref(err),
        )
        if not result:
            raise PSLError(err.value.decode() if err.value else "encode failed")
        return bytes(result[:out_len.value])

    def list_protocols(self) -> List[str]:
        """List all available protocol names."""
        err = ctypes.c_char_p()
        result = _lib.psl_list(self._handle, ctypes.byref(err))
        if not result:
            raise PSLError(err.value.decode() if err.value else "list failed")
        return json.loads(result.decode())

    def __repr__(self):
        try:
            protos = self.list_protocols()
            return f"<PSL Library: {len(protos)} protocols>"
        except Exception:
            return "<PSL Library>"
`
}

// GenerateRustBinding generates a Rust FFI binding module.
func GenerateRustBinding() string {
	return `//! PSL Rust SDK — FFI binding to libpsl shared library.

use std::ffi::{CStr, CString};
use std::os::raw::c_char;
use std::ptr;

#[link(name = "psl")]
extern "C" {
    fn psl_new(err_out: *mut *mut c_char) -> usize;
    fn psl_free(h: usize);
    fn psl_load(h: usize, path: *const c_char, err_out: *mut *mut c_char) -> i32;
    fn psl_decode(
        h: usize, protocol: *const c_char,
        data: *const u8, len: usize,
        err_out: *mut *mut c_char,
    ) -> *mut c_char;
    fn psl_encode(
        h: usize, protocol: *const c_char,
        fields_json: *const c_char,
        out_len: *mut usize,
        err_out: *mut *mut c_char,
    ) -> *mut u8;
    fn psl_list(h: usize, err_out: *mut *mut c_char) -> *mut c_char;
    fn psl_free_string(s: *mut c_char);
    fn psl_free_buffer(buf: *mut u8);
}

/// PSL library error.
#[derive(Debug)]
pub struct PslError(String);

impl std::fmt::Display for PslError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.0)
    }
}

impl std::error::Error for PslError {}

fn read_err(err: *mut c_char) -> PslError {
    if err.is_null() {
        return PslError("unknown error".into());
    }
    let msg = unsafe { CStr::from_ptr(err).to_string_lossy().into_owned() };
    unsafe { psl_free_string(err) };
    PslError(msg)
}

/// PSL library handle with RAII.
pub struct Library {
    handle: usize,
}

impl Library {
    /// Create a new PSL library instance.
    pub fn new() -> Result<Self, PslError> {
        let mut err: *mut c_char = ptr::null_mut();
        let h = unsafe { psl_new(&mut err) };
        if h == 0 {
            return Err(read_err(err));
        }
        Ok(Library { handle: h })
    }

    /// Load a PSL file.
    pub fn load(&self, path: &str) -> Result<(), PslError> {
        let c_path = CString::new(path).map_err(|e| PslError(e.to_string()))?;
        let mut err: *mut c_char = ptr::null_mut();
        let rc = unsafe { psl_load(self.handle, c_path.as_ptr(), &mut err) };
        if rc != 0 {
            return Err(read_err(err));
        }
        Ok(())
    }

    /// Decode binary data and return JSON string.
    pub fn decode(&self, protocol: &str, data: &[u8]) -> Result<String, PslError> {
        let c_proto = CString::new(protocol).map_err(|e| PslError(e.to_string()))?;
        let mut err: *mut c_char = ptr::null_mut();
        let result = unsafe {
            psl_decode(self.handle, c_proto.as_ptr(), data.as_ptr(), data.len(), &mut err)
        };
        if result.is_null() {
            return Err(read_err(err));
        }
        let s = unsafe { CStr::from_ptr(result).to_string_lossy().into_owned() };
        unsafe { psl_free_string(result) };
        Ok(s)
    }

    /// Encode from JSON fields and return binary data.
    pub fn encode(&self, protocol: &str, fields_json: &str) -> Result<Vec<u8>, PslError> {
        let c_proto = CString::new(protocol).map_err(|e| PslError(e.to_string()))?;
        let c_fields = CString::new(fields_json).map_err(|e| PslError(e.to_string()))?;
        let mut out_len: usize = 0;
        let mut err: *mut c_char = ptr::null_mut();
        let result = unsafe {
            psl_encode(self.handle, c_proto.as_ptr(), c_fields.as_ptr(), &mut out_len, &mut err)
        };
        if result.is_null() {
            return Err(read_err(err));
        }
        let data = unsafe { std::slice::from_raw_parts(result, out_len).to_vec() };
        unsafe { psl_free_buffer(result) };
        Ok(data)
    }

    /// List all protocol names.
    pub fn list_protocols(&self) -> Result<Vec<String>, PslError> {
        let mut err: *mut c_char = ptr::null_mut();
        let result = unsafe { psl_list(self.handle, &mut err) };
        if result.is_null() {
            return Err(read_err(err));
        }
        let s = unsafe { CStr::from_ptr(result).to_string_lossy().into_owned() };
        unsafe { psl_free_string(result) };
        serde_json::from_str(&s).map_err(|e| PslError(e.to_string()))
    }
}

impl Drop for Library {
    fn drop(&mut self) {
        unsafe { psl_free(self.handle) };
    }
}

unsafe impl Send for Library {}
unsafe impl Sync for Library {}
`
}

// GenerateTypeScriptBinding generates a TypeScript/Node.js binding module.
func GenerateTypeScriptBinding() string {
	return `/**
 * PSL TypeScript SDK — Node.js FFI binding to libpsl shared library.
 * Also supports WASM for browser environments.
 */

import { createRequire } from 'module';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';

// Type definitions
export interface DecodeResult {
  [field: string]: any;
}

export interface PSLLibrary {
  decode(protocol: string, data: Buffer | Uint8Array): DecodeResult;
  encode(protocol: string, fields: Record<string, any>): Buffer;
  listProtocols(): string[];
  load(path: string): void;
  close(): void;
}

/**
 * Create a new PSL library instance.
 */
export function createLibrary(): PSLLibrary {
  let ffi: any;
  try {
    const require = createRequire(import.meta.url);
    ffi = require('ffi-napi');
  } catch {
    throw new Error('ffi-napi is required: npm install ffi-napi');
  }

  const libPath = process.env.PSL_LIB_PATH || 'libpsl';
  const lib = ffi.Library(libPath, {
    psl_new: ['pointer', ['pointer']],
    psl_free: ['void', ['pointer']],
    psl_load: ['int', ['pointer', 'string', 'pointer']],
    psl_decode: ['string', ['pointer', 'string', 'pointer', 'size_t', 'pointer']],
    psl_encode: ['pointer', ['pointer', 'string', 'string', 'pointer', 'pointer']],
    psl_list: ['string', ['pointer', 'pointer']],
    psl_free_string: ['void', ['string']],
  });

  const errBuf = Buffer.alloc(8);
  const handle = lib.psl_new(errBuf);
  if (!handle || handle.isNull()) {
    throw new Error('Failed to create PSL library');
  }

  return {
    decode(protocol: string, data: Buffer | Uint8Array): DecodeResult {
      const buf = Buffer.from(data);
      const errPtr = Buffer.alloc(8);
      const result = lib.psl_decode(handle, protocol, buf, buf.length, errPtr);
      if (!result) throw new Error('decode failed');
      return JSON.parse(result);
    },

    encode(protocol: string, fields: Record<string, any>): Buffer {
      const json = JSON.stringify(fields);
      const lenBuf = Buffer.alloc(8);
      const errPtr = Buffer.alloc(8);
      const result = lib.psl_encode(handle, protocol, json, lenBuf, errPtr);
      if (!result) throw new Error('encode failed');
      const len = lenBuf.readUIntLE(0, 8);
      return Buffer.from(result.reinterpret(len));
    },

    listProtocols(): string[] {
      const errPtr = Buffer.alloc(8);
      const result = lib.psl_list(handle, errPtr);
      if (!result) throw new Error('list failed');
      return JSON.parse(result);
    },

    load(path: string): void {
      const errPtr = Buffer.alloc(8);
      const rc = lib.psl_load(handle, path, errPtr);
      if (rc !== 0) throw new Error('load failed');
    },

    close(): void {
      lib.psl_free(handle);
    },
  };
}

// TypeScript type definitions for protocol fields
export type ProtocolFields = Record<string, string | number | boolean | Buffer | ProtocolFields>;
`
}

// GenerateCMakeExample generates a CMake integration example.
func GenerateCMakeExample() string {
	return `# CMakeLists.txt — PSL C/C++ SDK integration example
cmake_minimum_required(VERSION 3.14)
project(psl_example C)

# Find the PSL shared library
find_library(PSL_LIB psl HINTS ${PSL_LIB_PATH} /usr/local/lib)
find_path(PSL_INCLUDE psl.h HINTS ${PSL_INCLUDE_PATH} /usr/local/include)

if(NOT PSL_LIB)
    message(FATAL_ERROR "libpsl not found. Set PSL_LIB_PATH.")
endif()

add_executable(example main.c)
target_include_directories(example PRIVATE ${PSL_INCLUDE})
target_link_libraries(example ${PSL_LIB})
`
}

// --- Scapy interop ---

// GenerateScapyInterop generates Python code for Scapy integration.
func GenerateScapyInterop() string {
	return `"""PSL ↔ Scapy interoperability layer."""

from scapy.all import Packet, raw
from protospec import Protocol

_psl = Protocol()

def psl_to_scapy(protocol: str, data: bytes) -> dict:
    """Decode binary data using PSL and return fields compatible with Scapy."""
    return _psl.decode(protocol, data)

def scapy_to_psl(protocol: str, pkt: Packet) -> dict:
    """Convert a Scapy packet to PSL-decoded fields."""
    return _psl.decode(protocol, raw(pkt))

def psl_encode_for_scapy(protocol: str, fields: dict) -> bytes:
    """Encode fields using PSL, suitable for Scapy injection."""
    return _psl.encode(protocol, fields)
`
}

// unused import guard
var _ = unsafe.Pointer(nil)
var _ = json.Marshal
