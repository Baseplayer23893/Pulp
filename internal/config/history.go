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
	URL       string `json:"url"`
	Source    string `json:"source"`    // extract, youtube, instagram, reddit, pdf
	Title     string `json:"title"`
	OutputFile string `json:"output_file"`
	WordCount int    `json:"word_count"`
	Timestamp string `json:"timestamp"`
}

// History manages extraction history
type History struct {
	Entries []HistoryEntry `json:"entries"`
	mu      sync.Mutex
}

// LoadHistory reads history from disk
func LoadHistory() *History {
	h := &History{}
	cfg := Load()

	data, err := os.ReadFile(cfg.HistoryFile)
	if err != nil {
		return h
	}

	json.Unmarshal(data, h)
	return h
}

// Add appends a new entry and persists to disk
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

// Recent returns the N most recent entries
func (h *History) Recent(n int) []HistoryEntry {
	h.mu.Lock()
	defer h.mu.Unlock()

	if n > len(h.Entries) {
		n = len(h.Entries)
	}
	return h.Entries[:n]
}

// Stats returns summary statistics
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

func (h *History) save() error {
	cfg := Load()

	if err := EnsureConfigDir(); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	return os.WriteFile(cfg.HistoryFile, data, 0644)
}
