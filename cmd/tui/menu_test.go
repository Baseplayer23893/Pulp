package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Baseplayer23893/Pulp/internal/config"
)

func TestDetectSource(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want int
	}{
		{name: "youtube", url: "https://youtu.be/abc", want: 1},
		{name: "instagram", url: "https://instagram.com/reel/x", want: 2},
		{name: "reddit", url: "https://reddit.com/r/golang", want: 3},
		{name: "pdf", url: "https://example.com/file.PDF", want: 4},
		{name: "default web", url: "https://example.com", want: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := detectSource(tc.url)
			if got != tc.want {
				t.Fatalf("detectSource(%q)=%d, want %d", tc.url, got, tc.want)
			}
		})
	}
}

func TestRelTimeRecentDuration(t *testing.T) {
	ts := time.Now().Add(-45 * time.Minute).Format(time.RFC3339)
	got := relTime(ts)
	if !strings.HasSuffix(got, "m ago") {
		t.Fatalf("relTime(%q)=%q, want minutes ago format", ts, got)
	}
}

func TestResultSaveUsesConfiguredOutputDir(t *testing.T) {
	cfgHome := t.TempDir()
	outDir := filepath.Join(t.TempDir(), "saved")
	t.Setenv("XDG_CONFIG_HOME", cfgHome)

	cfg := config.Load()
	cfg.OutputDir = outDir
	if err := cfg.Save(); err != nil {
		t.Fatalf("save config: %v", err)
	}

	m := initialModel()
	m.state = stateResult
	m.squeezeURL = "https://example.com/path/article"
	m.squeezeOutput = "hello world"

	next, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	got := next.(Model)
	want := filepath.Join(outDir, "article.md")
	if _, err := os.Stat(want); err != nil {
		t.Fatalf("expected saved file %q: %v", want, err)
	}
	if !strings.Contains(got.statusMsg, "Saved") {
		t.Fatalf("expected save status, got %q", got.statusMsg)
	}
}

func TestResultCopyUppercaseKeyShowsFeedback(t *testing.T) {
	m := initialModel()
	m.state = stateResult
	m.squeezeOutput = ""

	next, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'C'}})
	got := next.(Model)
	if !strings.Contains(got.statusMsg, "Nothing to copy") {
		t.Fatalf("expected copy warning, got %q", got.statusMsg)
	}
}
