//go:build js && wasm

package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/hsqbyte/protospec/src/protocol"
)

var lib *protocol.Library

func errJSON(msg string) string {
	b, _ := json.Marshal(map[string]string{"error": msg})
	return string(b)
}

func listProtocols(_ js.Value, _ []js.Value) any {
	type item struct {
		Name        string            `json:"name"`
		RFC         string            `json:"rfc,omitempty"`
		Title       map[string]string `json:"title,omitempty"`
		Description map[string]string `json:"description,omitempty"`
		Layer       string            `json:"layer,omitempty"`
		Type        string            `json:"type,omitempty"`
		DependsOn   []string          `json:"depends_on,omitempty"`
		Status      string            `json:"status,omitempty"`
	}
	var list []item
	for _, name := range lib.AllNames() {
		it := item{Name: name}
		if m := lib.Meta(name); m != nil {
			it.RFC = m.RFC
			it.Title = m.Title
			it.Description = m.Description
			it.Layer = m.Layer
			it.Type = m.Type
			it.DependsOn = m.DependsOn
			it.Status = m.Status
		}
		list = append(list, it)
	}
	b, _ := json.Marshal(list)
	return string(b)
}

func getProtocol(_ js.Value, args []js.Value) any {
	if len(args) < 1 {
		return errJSON("missing protocol name")
	}
	name := args[0].String()
	s, err := lib.Registry().GetSchema(name)
	if err == nil {
		b, _ := json.Marshal(s)
		return string(b)
	}
	if ms := lib.Message(name); ms != nil {
		b, _ := json.Marshal(ms)
		return string(b)
	}
	return errJSON(fmt.Sprintf("protocol %q not found", name))
}

func getMeta(_ js.Value, args []js.Value) any {
	if len(args) < 1 {
		return errJSON("missing protocol name")
	}
	m := lib.Meta(args[0].String())
	if m == nil {
		return errJSON(fmt.Sprintf("meta for %q not found", args[0].String()))
	}
	b, _ := json.Marshal(m)
	return string(b)
}

func encode(_ js.Value, args []js.Value) any {
	if len(args) < 2 {
		return errJSON("usage: encode(protocolName, jsonFieldValues)")
	}
	name := args[0].String()
	var packet map[string]any
	if err := json.Unmarshal([]byte(args[1].String()), &packet); err != nil {
		return errJSON("invalid JSON: " + err.Error())
	}
	data, err := lib.Encode(name, packet)
	if err != nil {
		return errJSON(err.Error())
	}
	b, _ := json.Marshal(map[string]string{"hex": hex.EncodeToString(data)})
	return string(b)
}

func decode(_ js.Value, args []js.Value) any {
	if len(args) < 2 {
		return errJSON("usage: decode(protocolName, hexString)")
	}
	name := args[0].String()
	data, err := hex.DecodeString(args[1].String())
	if err != nil {
		return errJSON("invalid hex: " + err.Error())
	}
	result, err := lib.Decode(name, data)
	if err != nil {
		return errJSON(err.Error())
	}
	b, _ := json.Marshal(result)
	return string(b)
}

func main() {
	var err error
	lib, err = protocol.NewLibrary()
	if err != nil {
		fmt.Println("failed to init library:", err)
		return
	}

	js.Global().Set("psl_listProtocols", js.FuncOf(listProtocols))
	js.Global().Set("psl_getProtocol", js.FuncOf(getProtocol))
	js.Global().Set("psl_getMeta", js.FuncOf(getMeta))
	js.Global().Set("psl_encode", js.FuncOf(encode))
	js.Global().Set("psl_decode", js.FuncOf(decode))

	fmt.Println("protospec wasm loaded")

	// Block forever so the Go runtime stays alive
	select {}
}
