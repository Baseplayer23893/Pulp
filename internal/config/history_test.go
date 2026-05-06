package config

import (
	"os"
	"testing"
)

func TestLoadHistoryInvalidJSONReturnsEmptyHistory(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := EnsureConfigDir(); err != nil {
		t.Fatalf("ensure config dir: %v", err)
	}
	if err := os.WriteFile(HistoryPath(), []byte("{invalid json"), 0644); err != nil {
		t.Fatalf("write history: %v", err)
	}

	h := LoadHistory()
	if len(h.Entries) != 0 {
		t.Fatalf("expected empty history for invalid JSON, got %d entries", len(h.Entries))
	}
}

func TestRecentNonPositiveCount(t *testing.T) {
	h := &History{Entries: []HistoryEntry{{URL: "https://example.com"}}}

	if got := h.Recent(0); len(got) != 0 {
		t.Fatalf("Recent(0) returned %d entries, want 0", len(got))
	}
	if got := h.Recent(-1); len(got) != 0 {
		t.Fatalf("Recent(-1) returned %d entries, want 0", len(got))
	}
}
