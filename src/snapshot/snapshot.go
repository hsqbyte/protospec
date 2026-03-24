// Package snapshot provides protocol library snapshot and session recording.
package snapshot

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Snapshot represents a saved protocol library state.
type Snapshot struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	Hash      string    `json:"hash"`
	Protocols []string  `json:"protocols"`
	Size      int64     `json:"size"`
}

// Manager manages protocol snapshots.
type Manager struct {
	dir string
}

// NewManager creates a new snapshot manager.
func NewManager(dir string) *Manager {
	os.MkdirAll(dir, 0755)
	return &Manager{dir: dir}
}

// Create creates a new snapshot.
func (m *Manager) Create(name string, protocols []string) (*Snapshot, error) {
	data, _ := json.Marshal(protocols)
	hash := sha256.Sum256(data)

	snap := &Snapshot{
		Name:      name,
		CreatedAt: time.Now(),
		Hash:      fmt.Sprintf("%x", hash[:8]),
		Protocols: protocols,
		Size:      int64(len(data)),
	}

	snapData, _ := json.MarshalIndent(snap, "", "  ")
	path := filepath.Join(m.dir, name+".json")
	if err := os.WriteFile(path, snapData, 0644); err != nil {
		return nil, fmt.Errorf("save snapshot: %w", err)
	}
	return snap, nil
}

// List lists all snapshots.
func (m *Manager) List() ([]*Snapshot, error) {
	entries, err := os.ReadDir(m.dir)
	if err != nil {
		return nil, err
	}

	var snapshots []*Snapshot
	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(m.dir, e.Name()))
		if err != nil {
			continue
		}
		var snap Snapshot
		if json.Unmarshal(data, &snap) == nil {
			snapshots = append(snapshots, &snap)
		}
	}
	return snapshots, nil
}

// Load loads a snapshot by name.
func (m *Manager) Load(name string) (*Snapshot, error) {
	path := filepath.Join(m.dir, name+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("snapshot %q not found", name)
	}
	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, err
	}
	return &snap, nil
}

// SessionRecord represents a recorded protocol session.
type SessionRecord struct {
	ID        string    `json:"id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Packets   int       `json:"packets"`
	Protocol  string    `json:"protocol"`
}
