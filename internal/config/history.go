package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// HistoryEntry tracks a single extraction
type HistoryEntry struct {
	URL        string `json:"url"`
	Source     string `json:"source"` // extract, youtube, instagram, reddit, pdf
	Title      string `json:"title"`
	OutputFile string `json:"output_file"`
	WordCount  int    `json:"word_count"`
	Timestamp  string `json:"timestamp"`
}

// History manages extraction history
type History struct {
	Entries []HistoryEntry `json:"entries"`
	mu      sync.Mutex
}

// LoadHistory reads history from disk.
// Returns an empty History if the file does not exist yet.
func LoadHistory() *History {
	h := &History{}

	data, err := os.ReadFile(HistoryPath())
	if err != nil {
		// File doesn't exist yet — that's fine.
		return h
	}

	if err := json.Unmarshal(data, h); err != nil {
		return &History{}
	}
	return h
}

// Add appends a new entry (newest first) and immediately persists to disk.
func (h *History) Add(entry HistoryEntry) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if entry.Timestamp == "" {
		entry.Timestamp = time.Now().Format(time.RFC3339)
	}

	// Prepend (newest first)
	h.Entries = append([]HistoryEntry{entry}, h.Entries...)

	// Trim to max
	cfg := Load()
	if len(h.Entries) > cfg.MaxHistory {
		h.Entries = h.Entries[:cfg.MaxHistory]
	}

	return h.save()
}

// Delete removes the entry at the given index (0-based, into the Recent slice)
// and immediately persists the updated list to disk.
func (h *History) Delete(index int) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if index < 0 || index >= len(h.Entries) {
		return fmt.Errorf("index %d out of range", index)
	}

	h.Entries = append(h.Entries[:index], h.Entries[index+1:]...)
	return h.save()
}

// Recent returns the N most recent entries.
func (h *History) Recent(n int) []HistoryEntry {
	h.mu.Lock()
	defer h.mu.Unlock()

	if n <= 0 {
		return nil
	}
	if n > len(h.Entries) {
		n = len(h.Entries)
	}
	recent := make([]HistoryEntry, n)
	copy(recent, h.Entries[:n])
	return recent
}

// Stats returns summary statistics.
func (h *History) Stats() (total int, bySource map[string]int, totalWords int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	bySource = make(map[string]int)
	for _, e := range h.Entries {
		bySource[e.Source]++
		totalWords += e.WordCount
	}
	return len(h.Entries), bySource, totalWords
}

// save writes the history to history.json. Caller must hold h.mu.
func (h *History) save() error {
	if err := EnsureConfigDir(); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	return os.WriteFile(HistoryPath(), data, 0644)
}
