package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"sync"
)

// Bundle holds loaded translations for a specific domain (e.g. "cli").
type Bundle struct {
	mu   sync.RWMutex
	data map[string]map[string]string // lang -> key -> value
}

// NewBundle creates an empty Bundle.
func NewBundle() *Bundle {
	return &Bundle{data: make(map[string]map[string]string)}
}

// LoadFS loads all JSON files from an embed.FS directory.
// Each file name (without .json) is treated as a language code.
func (b *Bundle) LoadFS(fs embed.FS, dir string) error {
	entries, err := fs.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("i18n: read dir %s: %w", dir, err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if len(name) < 6 || name[len(name)-5:] != ".json" {
			continue
		}
		lang := name[:len(name)-5]

		raw, err := fs.ReadFile(dir + "/" + name)
		if err != nil {
			return fmt.Errorf("i18n: read %s: %w", name, err)
		}

		var msgs map[string]string
		if err := json.Unmarshal(raw, &msgs); err != nil {
			return fmt.Errorf("i18n: parse %s: %w", name, err)
		}

		b.mu.Lock()
		b.data[lang] = msgs
		b.mu.Unlock()
	}
	return nil
}

// Get returns the message for the given language and key.
// Falls back to English, then returns the key itself.
func (b *Bundle) Get(lang, key string) string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if msgs, ok := b.data[lang]; ok {
		if s, ok := msgs[key]; ok {
			return s
		}
	}
	if msgs, ok := b.data["en"]; ok {
		if s, ok := msgs[key]; ok {
			return s
		}
	}
	return key
}
