package tui

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
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
		{name: "hn", url: "https://news.ycombinator.com/item?id=123", want: 4},
		{name: "pdf", url: "https://example.com/file.PDF", want: 5},
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

func TestSettingsSaveCreatesConfiguredOutputDir(t *testing.T) {
	cfgHome := t.TempDir()
	outDir := filepath.Join(t.TempDir(), "new-output")
	t.Setenv("XDG_CONFIG_HOME", cfgHome)

	m := initialModel()
	m.state = stateSettings
	m.settingsValues = [2]string{outDir, "md"}

	next, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	got := next.(Model)

	if _, err := os.Stat(outDir); err != nil {
		t.Fatalf("expected settings save to create output dir %q: %v", outDir, err)
	}
	cfg := config.Load()
	if cfg.OutputDir != outDir {
		t.Fatalf("saved output dir = %q, want %q", cfg.OutputDir, outDir)
	}
	if got.state != stateHome {
		t.Fatalf("expected settings save to return home, got state %v", got.state)
	}
	if !strings.Contains(got.noticeMsg, "Settings saved") {
		t.Fatalf("expected saved notice, got %q", got.noticeMsg)
	}
}

func TestCopyFallbackWritesCacheFileWhenClipboardUnavailable(t *testing.T) {
	cacheHome := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cacheHome)

	origWriteClipboard := writeClipboard
	origReadClipboard := readClipboard
	writeClipboard = func(string) error { return errors.New("clipboard unavailable") }
	readClipboard = func() (string, error) { return "", errors.New("clipboard unavailable") }
	t.Cleanup(func() {
		writeClipboard = origWriteClipboard
		readClipboard = origReadClipboard
	})

	msg, err := copyResultOutput("hello world")
	if err != nil {
		t.Fatalf("copyResultOutput returned error: %v", err)
	}
	want := filepath.Join(cacheHome, "pulp", "last-copy.md")
	if !strings.Contains(msg, want) {
		t.Fatalf("expected fallback path in status %q", msg)
	}
	data, err := os.ReadFile(want)
	if err != nil {
		t.Fatalf("expected fallback copy file: %v", err)
	}
	if string(data) != "hello world" {
		t.Fatalf("fallback file = %q, want copied content", data)
	}
}

func TestCopyFallbackWhenClipboardDoesNotRetainContent(t *testing.T) {
	cacheHome := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cacheHome)

	origWriteClipboard := writeClipboard
	origReadClipboard := readClipboard
	writeClipboard = func(string) error { return nil }
	readClipboard = func() (string, error) { return "", errors.New("nothing copied") }
	t.Cleanup(func() {
		writeClipboard = origWriteClipboard
		readClipboard = origReadClipboard
	})

	msg, err := copyResultOutput("hello arch")
	if err != nil {
		t.Fatalf("copyResultOutput returned error: %v", err)
	}
	want := filepath.Join(cacheHome, "pulp", "last-copy.md")
	if !strings.Contains(msg, want) {
		t.Fatalf("expected fallback path in status %q", msg)
	}
}

func TestCopySuccessWhenClipboardDropsTrailingNewline(t *testing.T) {
	origWriteClipboard := writeClipboard
	origReadClipboard := readClipboard
	writeClipboard = func(string) error { return nil }
	readClipboard = func() (string, error) { return "hello world", nil }
	t.Cleanup(func() {
		writeClipboard = origWriteClipboard
		readClipboard = origReadClipboard
	})

	msg, err := copyResultOutput("hello world\n")
	if err != nil {
		t.Fatalf("copyResultOutput returned error: %v", err)
	}
	if !strings.Contains(msg, "Copied to clipboard") {
		t.Fatalf("expected clipboard success message, got %q", msg)
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

func TestHistoryNavigationMovesBeyondVisibleWindow(t *testing.T) {
	m := initialModel()
	m.state = stateHistory
	m.height = 12 // maxShow = 2 entries
	m.history = &config.History{Entries: make([]config.HistoryEntry, 20)}
	for i := range m.history.Entries {
		m.history.Entries[i] = config.HistoryEntry{Title: "entry-" + string(rune('A'+i))}
	}

	for i := 0; i < 10; i++ {
		next, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		m = next.(Model)
	}

	if m.cursor != 10 {
		t.Fatalf("expected cursor at absolute index 10, got %d", m.cursor)
	}
}

func TestHistoryViewShowsWindowAroundCursor(t *testing.T) {
	m := initialModel()
	m.state = stateHistory
	m.height = 14 // maxShow = 4 entries
	m.cursor = 8
	m.history = &config.History{Entries: make([]config.HistoryEntry, 12)}
	for i := range m.history.Entries {
		m.history.Entries[i] = config.HistoryEntry{
			Title: "entry-" + twoDigits(i),
			URL:   "https://example.com/item-" + strconv.Itoa(i),
		}
	}

	view := m.viewHistory()
	if !strings.Contains(view, "entry-08") {
		t.Fatalf("expected selected window to contain cursor item, view=%q", view)
	}
	if strings.Contains(view, "entry-00") {
		t.Fatalf("expected window to scroll away from first entries, view=%q", view)
	}
}

func twoDigits(n int) string {
	if n < 10 {
		return "0" + strconv.Itoa(n)
	}
	return strconv.Itoa(n)
}
