// Package hotload provides runtime protocol hot-loading and dynamic registration.
package hotload

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hsqbyte/protospec/src/protocol"
)

// Watcher watches for PSL file changes and hot-reloads protocols.
type Watcher struct {
	lib       *protocol.Library
	dir       string
	interval  time.Duration
	mu        sync.RWMutex
	versions  map[string]time.Time
	callbacks []func(event Event)
	stop      chan struct{}
}

// Event represents a hot-load event.
type Event struct {
	Type     string    `json:"type"` // "loaded", "updated", "removed", "error"
	Protocol string    `json:"protocol"`
	File     string    `json:"file"`
	Time     time.Time `json:"time"`
	Error    string    `json:"error,omitempty"`
}

// NewWatcher creates a new file watcher for hot-loading.
func NewWatcher(lib *protocol.Library, dir string, interval time.Duration) *Watcher {
	if interval <= 0 {
		interval = 2 * time.Second
	}
	return &Watcher{
		lib:      lib,
		dir:      dir,
		interval: interval,
		versions: make(map[string]time.Time),
		stop:     make(chan struct{}),
	}
}

// OnEvent registers a callback for hot-load events.
func (w *Watcher) OnEvent(cb func(Event)) {
	w.callbacks = append(w.callbacks, cb)
}

// Scan scans for changed PSL files and reloads them.
func (w *Watcher) Scan() []Event {
	var events []Event

	entries, err := os.ReadDir(w.dir)
	if err != nil {
		return events
	}

	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".psl" {
			continue
		}
		path := filepath.Join(w.dir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		w.mu.RLock()
		lastMod, known := w.versions[path]
		w.mu.RUnlock()

		if !known || info.ModTime().After(lastMod) {
			err := w.lib.LoadPSL(path)
			event := Event{
				File: path,
				Time: time.Now(),
			}
			if err != nil {
				event.Type = "error"
				event.Error = err.Error()
			} else if known {
				event.Type = "updated"
			} else {
				event.Type = "loaded"
			}

			w.mu.Lock()
			w.versions[path] = info.ModTime()
			w.mu.Unlock()

			events = append(events, event)
			for _, cb := range w.callbacks {
				cb(event)
			}
		}
	}
	return events
}

// Stats returns hot-load statistics.
type Stats struct {
	WatchedFiles int       `json:"watched_files"`
	LastScan     time.Time `json:"last_scan"`
	TotalLoads   int       `json:"total_loads"`
}

// GetStats returns current watcher statistics.
func (w *Watcher) GetStats() Stats {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return Stats{
		WatchedFiles: len(w.versions),
		LastScan:     time.Now(),
	}
}

// FormatEvent formats an event for display.
func FormatEvent(e Event) string {
	icon := "●"
	switch e.Type {
	case "loaded":
		icon = "+"
	case "updated":
		icon = "~"
	case "removed":
		icon = "-"
	case "error":
		icon = "✗"
	}
	msg := fmt.Sprintf("[%s] %s %s", icon, e.Type, e.File)
	if e.Error != "" {
		msg += " — " + e.Error
	}
	return msg
}
