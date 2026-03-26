// Package sdk provides the multi-language SDK export layer.
// It exposes core PSL functionality as C-compatible functions
// that can be called from Python, Rust, JavaScript, and C/C++.
package sdk

import (
	"encoding/json"
	"sync"
	"unsafe"

	"github.com/hsqbyte/protospec/src/protocol"
)

// Handle is an opaque pointer to a Library instance.
type Handle uintptr

var (
	mu      sync.Mutex
	handles        = map[Handle]*protocol.Library{}
	nextID  Handle = 1
)

// NewLibrary creates a new PSL library instance and returns a handle.
func NewLibrary() (Handle, error) {
	lib, err := protocol.NewLibrary()
	if err != nil {
		return 0, err
	}
	mu.Lock()
	h := nextID
	nextID++
	handles[h] = lib
	mu.Unlock()
	return h, nil
}

// GetLibrary retrieves the library for a handle.
func GetLibrary(h Handle) *protocol.Library {
	mu.Lock()
	defer mu.Unlock()
	return handles[h]
}

// FreeLibrary releases a library handle.
func FreeLibrary(h Handle) {
	mu.Lock()
	delete(handles, h)
	mu.Unlock()
}

// Decode decodes binary data using the named protocol and returns JSON.
func Decode(h Handle, protoName string, data []byte) (string, error) {
	lib := GetLibrary(h)
	if lib == nil {
		return "", errInvalidHandle
	}
	result, err := lib.Decode(protoName, data)
	if err != nil {
		return "", err
	}
	out, err := json.Marshal(result.Packet)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// Encode encodes JSON fields into binary using the named protocol.
func Encode(h Handle, protoName string, fieldsJSON string) ([]byte, error) {
	lib := GetLibrary(h)
	if lib == nil {
		return nil, errInvalidHandle
	}
	var fields map[string]any
	if err := json.Unmarshal([]byte(fieldsJSON), &fields); err != nil {
		return nil, err
	}
	return lib.Encode(protoName, fields)
}

// ListProtocols returns all protocol names as JSON array.
func ListProtocols(h Handle) (string, error) {
	lib := GetLibrary(h)
	if lib == nil {
		return "", errInvalidHandle
	}
	names := lib.AllNames()
	out, err := json.Marshal(names)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// LoadPSL loads a PSL file into the library.
func LoadPSL(h Handle, filePath string) error {
	lib := GetLibrary(h)
	if lib == nil {
		return errInvalidHandle
	}
	return lib.LoadPSL(filePath)
}

// CString helpers for FFI
func GoString(p *byte) string {
	if p == nil {
		return ""
	}
	var buf []byte
	for ptr := unsafe.Pointer(p); *(*byte)(ptr) != 0; ptr = unsafe.Add(ptr, 1) {
		buf = append(buf, *(*byte)(ptr))
	}
	return string(buf)
}

var errInvalidHandle = &SDKError{msg: "invalid library handle"}

// SDKError represents an SDK error.
type SDKError struct {
	msg string
}

func (e *SDKError) Error() string { return e.msg }
