// Package snapshot captures a point-in-time state of parsed log entries
// across services and can serialize/deserialize them for later comparison.
package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/yourorg/logdrift/internal/parser"
)

// Snapshot holds a labeled collection of log entries from one or more services.
type Snapshot struct {
	Label     string                    `json:"label"`
	CreatedAt time.Time                 `json:"created_at"`
	Entries   map[string][]parser.Entry `json:"entries"`
}

// New creates an empty Snapshot with the given label.
func New(label string) *Snapshot {
	return &Snapshot{
		Label:     label,
		CreatedAt: time.Now().UTC(),
		Entries:   make(map[string][]parser.Entry),
	}
}

// Add appends an entry under the given service name.
func (s *Snapshot) Add(service string, e parser.Entry) {
	s.Entries[service] = append(s.Entries[service], e)
}

// Services returns the sorted list of service names present in the snapshot.
func (s *Snapshot) Services() []string {
	keys := make([]string, 0, len(s.Entries))
	for k := range s.Entries {
		keys = append(keys, k)
	}
	return keys
}

// Save writes the snapshot as JSON to the given file path.
func (s *Snapshot) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("snapshot: create file: %w", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(s); err != nil {
		return fmt.Errorf("snapshot: encode: %w", err)
	}
	return nil
}

// Load reads a snapshot from a JSON file at the given path.
func Load(path string) (*Snapshot, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("snapshot: open file: %w", err)
	}
	defer f.Close()
	var s Snapshot
	if err := json.NewDecoder(f).Decode(&s); err != nil {
		return nil, fmt.Errorf("snapshot: decode: %w", err)
	}
	return &s, nil
}
