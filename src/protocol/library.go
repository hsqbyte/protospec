package protocol

import (
	"encoding/json"
	"io/fs"
	"path/filepath"
	"github.com/hsqbyte/protospec/psl"
	"github.com/hsqbyte/protospec/src/core/checksum"
	"github.com/hsqbyte/protospec/src/core/codec"
	"github.com/hsqbyte/protospec/src/core/format"
	"github.com/hsqbyte/protospec/src/core/pdl"
	"github.com/hsqbyte/protospec/src/core/registry"
	"github.com/hsqbyte/protospec/src/core/schema"
	"sort"
	"strings"
)

// ProtocolMeta holds metadata loaded from meta.json alongside a protocol.
type ProtocolMeta struct {
	RFC         string                       `json:"rfc"`
	URL         string                       `json:"url"`
	Title       map[string]string            `json:"title"`
	Description map[string]string            `json:"description"`
	Status      string                       `json:"status"`
	SeeAlso     []string                     `json:"see_also"`
	Type        string                       `json:"type"`
	Layer       string                       `json:"layer"`
	DependsOn   []string                     `json:"depends_on"`
	Fields      map[string]map[string]string `json:"fields"`
}

// Library is the top-level API for protocol encoding and decoding.
type Library struct {
	registry         *registry.ProtocolRegistry
	engine           *codec.CodecEngine
	checksumRegistry *checksum.ChecksumRegistry
	formatRegistry   *format.FormatRegistry
	metas            map[string]*ProtocolMeta
	messages         map[string]*schema.MessageSchema
}

// NewLibrary creates a Library with default registries and loads all
// built-in PSL files from the psl/ directory.
func NewLibrary() (*Library, error) {
	cr := checksum.NewDefaultChecksumRegistry()
	fr := format.NewDefaultFormatRegistry()
	lib := &Library{
		registry:         registry.NewProtocolRegistry(cr, fr),
		engine:           codec.NewCodecEngine(cr, fr),
		checksumRegistry: cr,
		formatRegistry:   fr,
		metas:            make(map[string]*ProtocolMeta),
		messages:         make(map[string]*schema.MessageSchema),
	}
	if err := lib.loadBuiltins(); err != nil {
		return nil, err
	}
	return lib, nil
}

// Meta returns the ProtocolMeta for the given protocol name, or nil if not found.
func (l *Library) Meta(name string) *ProtocolMeta {
	return l.metas[name]
}

// Message returns the MessageSchema for the given protocol name, or nil if not found.
func (l *Library) Message(name string) *schema.MessageSchema {
	return l.messages[name]
}

// AllNames returns all registered protocol names (binary + message) in sorted order.
func (l *Library) AllNames() []string {
	names := l.registry.List()
	for name := range l.messages {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Encode encodes a packet by protocol name.
func (l *Library) Encode(protocolName string, packet map[string]any) ([]byte, error) {
	s, err := l.registry.GetSchema(protocolName)
	if err != nil {
		return nil, err
	}
	return l.engine.Encode(s, packet)
}

// DecodeResult re-exports the codec DecodeResult for public use.
type DecodeResult = codec.DecodeResult

// Decode decodes raw bytes by protocol name.
func (l *Library) Decode(protocolName string, data []byte) (*DecodeResult, error) {
	s, err := l.registry.GetSchema(protocolName)
	if err != nil {
		return nil, err
	}
	return l.engine.Decode(s, data)
}

// Codec is a standalone encoder/decoder created from PSL text without
// registering the protocol.
type Codec struct {
	schema *schema.ProtocolSchema
	engine *codec.CodecEngine
}

// Encode encodes a packet using this codec's schema.
func (c *Codec) Encode(packet map[string]any) ([]byte, error) {
	return c.engine.Encode(c.schema, packet)
}

// Decode decodes raw bytes using this codec's schema.
func (c *Codec) Decode(data []byte) (*DecodeResult, error) {
	return c.engine.Decode(c.schema, data)
}

// CreateCodec creates a standalone Codec from PSL text without registering.
func (l *Library) CreateCodec(pslText string) (*Codec, error) {
	parser := pdl.NewPDLParser(l.checksumRegistry, l.formatRegistry)
	s, err := parser.Parse(pslText)
	if err != nil {
		return nil, err
	}
	return &Codec{schema: s, engine: l.engine}, nil
}

// LoadPSL loads a PSL file and registers the protocol.
func (l *Library) LoadPSL(filePath string) error {
	return l.registry.RegisterFromFile(filePath)
}

// RegisterChecksum registers a custom checksum algorithm.
func (l *Library) RegisterChecksum(name string, fn checksum.ChecksumFunc) {
	l.checksumRegistry.Register(name, fn)
}

// RegisterFormat registers a custom display format converter.
func (l *Library) RegisterFormat(name string, formatter format.DisplayFormatter) {
	l.formatRegistry.Register(name, formatter)
}

// Registry returns the underlying protocol registry.
func (l *Library) Registry() *registry.ProtocolRegistry {
	return l.registry
}

// loadBuiltins loads all protocols from the embedded psl/ subdirectories.
// It recursively walks the filesystem to find protocol directories — any
// directory containing at least one .psl file is treated as a protocol.
func (l *Library) loadBuiltins() error {
	parser := pdl.NewPDLParser(l.checksumRegistry, l.formatRegistry)

	// Collect protocol directories: directories that contain .psl files.
	protoDirs := map[string]bool{}
	err := fs.WalkDir(psl.FS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".psl") {
			protoDirs[filepath.Dir(path)] = true
		}
		return nil
	})
	if err != nil {
		return err
	}

	for dirName := range protoDirs {
		if dirName == "." {
			continue
		}
		entries, err := psl.FS.ReadDir(dirName)
		if err != nil {
			return err
		}

		// Check meta.json first to determine protocol type
		var meta ProtocolMeta
		metaData, metaErr := psl.FS.ReadFile(dirName + "/meta.json")
		if metaErr == nil {
			json.Unmarshal(metaData, &meta)
		}

		// Load .psl files
		var protoName string
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".psl") {
				continue
			}
			data, err := psl.FS.ReadFile(dirName + "/" + entry.Name())
			if err != nil {
				return err
			}

			source := string(data)
			// Auto-detect protocol type from content: "message" keyword = message protocol
			isMessage := meta.Type == "message"
			if !isMessage {
				trimmed := strings.TrimSpace(source)
				if strings.HasPrefix(trimmed, "message ") || strings.HasPrefix(trimmed, "message\t") {
					isMessage = true
				}
				// Skip lines starting with comments
				for strings.HasPrefix(trimmed, "//") {
					idx := strings.Index(trimmed, "\n")
					if idx < 0 {
						break
					}
					trimmed = strings.TrimSpace(trimmed[idx+1:])
				}
				if strings.HasPrefix(trimmed, "message ") || strings.HasPrefix(trimmed, "message\t") {
					isMessage = true
				}
			}

			if isMessage {
				// Check if this is a transport definition (no "transport" declaration inside)
				// Transport definitions define message types like command/event/data
				// and are loaded on-demand by TransportLoader.
				if !strings.Contains(source, "transport ") {
					// This is a transport definition file — parse it as transport
					// and also register it as a message protocol for lookups.
					td, tdErr := pdl.ParseTransportPSL(source)
					if tdErr == nil {
						// Convert transport def to a MessageSchema with Messages populated
						ms := &schema.MessageSchema{
							Name:         td.Name,
							Version:      td.Version,
							TransportDef: td,
						}
						// Convert MessageTypeDefs to MessageDefs for codegen compatibility
						for _, mt := range td.MessageTypes {
							msg := schema.MessageDef{
								Name: mt.Name,
								Kind: mt.Name,
							}
							for _, f := range mt.Fields {
								msg.Fields = append(msg.Fields, schema.MessageFieldDef{
									Name:     f.Name,
									Type:     f.Type,
									Optional: f.Optional,
								})
							}
							ms.Messages = append(ms.Messages, msg)
						}
						l.messages[ms.Name] = ms
						protoName = ms.Name
					}
					continue
				}
				// Parse as message protocol (upper-layer with transport reference)
				loader := pdl.NewTransportLoader(psl.FS, nil)
				parser.SetTransportLoader(loader)
				ms, err := parser.ParseMessage(source)
				if err != nil {
					return err
				}
				l.messages[ms.Name] = ms
				protoName = ms.Name
			} else {
				// Parse as binary protocol
				name, err := l.registry.RegisterFromPSLReturnName(source)
				if err != nil {
					return err
				}
				protoName = name
			}
		}

		// Store meta
		if protoName != "" && metaErr == nil {
			l.metas[protoName] = &meta
		}
	}
	return nil
}
