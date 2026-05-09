package defuddle

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// mockDefuddleSource is a minimal Go program that acts as a mock defuddle CLI.
// It responds based on the URL argument — see the switch statement below.
const mockDefuddleSource = `package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func main() {
	args := os.Args[1:]
	if len(args) < 2 {
		os.Exit(1)
	}
	url := args[len(args)-1]

	switch {
	case strings.HasPrefix(url, "valid://"):
		json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"title":      "Test Title",
			"description": "A test page",
			"domain":     "example.com",
			"url":        url,
			"author":     "Test Author",
			"published":  "2024-01-01",
			"content":    "Hello world, this is some test content.",
			"markdown":   "# Test Title\n\nHello world, this is some test content.",
			"wordCount":  8,
		})
	case strings.HasPrefix(url, "emptymarkup://"):
		json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"markdown":  "",
			"content":   "",
			"title":     "",
			"wordCount": 0,
		})
	case strings.HasPrefix(url, "negativewords://"):
		json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"markdown":  "some content",
			"wordCount": -1,
		})
	case strings.HasPrefix(url, "plaintext://"):
		fmt.Fprint(os.Stdout, "# Just plain markdown\n\nNo JSON here.")
	case strings.HasPrefix(url, "error://"):
		fmt.Fprint(os.Stderr, "defuddle error: something went wrong")
		os.Exit(1)
	case strings.HasPrefix(url, "malformed://"):
		fmt.Fprint(os.Stdout, "{ this is not json }")
	default:
		json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"markdown":  "default response",
			"wordCount": 1,
		})
	}
}
`

// mockBin caches the compiled mock defuddle binary path per test binary run.
var mockBin struct {
	path string
	err  error
	done bool
}

func ensureMockBin(t *testing.T) string {
	if mockBin.done {
		if mockBin.err != nil {
			t.Fatalf("failed to build mock defuddle: %v", mockBin.err)
		}
		return mockBin.path
	}
	mockBin.done = true

	tmpDir, err := os.MkdirTemp("", "pulp-defuddle-mock")
	if err != nil {
		mockBin.err = fmt.Errorf("create temp dir: %w", err)
		t.Fatalf("failed to create temp dir: %v", mockBin.err)
	}

	src := filepath.Join(tmpDir, "mock_defuddle.go")
	if err := os.WriteFile(src, []byte(mockDefuddleSource), 0644); err != nil {
		mockBin.err = fmt.Errorf("write source: %w", err)
		t.Fatalf("failed to write mock source: %v", mockBin.err)
	}

	out := filepath.Join(tmpDir, "defuddle")
	cmd := exec.Command("go", "build", "-o", out, src)
	if err := cmd.Run(); err != nil {
		mockBin.err = fmt.Errorf("go build: %w", err)
		t.Fatalf("failed to build mock defuddle: %v", mockBin.err)
	}

	// Ensure the binary is executable.
	if err := os.Chmod(out, 0755); err != nil {
		mockBin.err = fmt.Errorf("chmod: %w", err)
		t.Fatalf("failed to chmod mock defuddle: %v", mockBin.err)
	}

	mockBin.path = out
	return mockBin.path
}

// withMockBin patches PATH to prefer the mock binary and calls fn.
// It restores the original PATH after fn returns.
func withMockBin(t *testing.T, fn func()) {
	mockPath := ensureMockBin(t)

	origPath := os.Getenv("PATH")
	mockDir := filepath.Dir(mockPath)
	newPath := mockDir + string(os.PathListSeparator) + origPath
	if err := os.Setenv("PATH", newPath); err != nil {
		t.Fatalf("set PATH: %v", err)
	}

	defer os.Setenv("PATH", origPath)

	fn()
}

// TestParseURL_ValidJSON tests that ParseURL correctly parses valid defuddle JSON output.
func TestParseURL_ValidJSON(t *testing.T) {
	withMockBin(t, func() {
		result, err := ParseURL("valid://example.com/page")
		if err != nil {
			t.Fatalf("ParseURL unexpected error: %v", err)
		}
		if result.Title != "Test Title" {
			t.Errorf("Title = %q, want %q", result.Title, "Test Title")
		}
		if result.Markdown == "" {
			t.Error("Markdown is empty")
		}
		if result.WordCount != 8 {
			t.Errorf("WordCount = %d, want %d", result.WordCount, 8)
		}
	})
}

// TestParseURL_EmptyContent tests that ParseURL returns an error when defuddle
// returns a structually valid JSON but with no usable content.
func TestParseURL_EmptyContent(t *testing.T) {
	withMockBin(t, func() {
		_, err := ParseURL("emptymarkup://example.com/empty")
		if err == nil {
			t.Fatal("ParseURL expected error for empty content, got nil")
		}
	})
}

// TestParseURL_PlaintextFallback tests that when defuddle outputs plain markdown
// (not JSON), ParseURL still returns a valid result with the markdown populated.
func TestParseURL_PlaintextFallback(t *testing.T) {
	withMockBin(t, func() {
		result, err := ParseURL("plaintext://example.com/md")
		if err != nil {
			t.Fatalf("ParseURL unexpected error: %v", err)
		}
		if result.Markdown == "" {
			t.Error("Markdown is empty")
		}
		if result.WordCount != 0 {
			t.Errorf("WordCount = %d, want 0 for plaintext fallback", result.WordCount)
		}
	})
}

// TestParseURL_CLIError tests that when defuddle exits with an error, ParseURL
// returns an error that includes the stderr output.
func TestParseURL_CLIError(t *testing.T) {
	withMockBin(t, func() {
		_, err := ParseURL("error://example.com/fail")
		if err == nil {
			t.Fatal("ParseURL expected error, got nil")
		}
		if err.Error() == "" {
			t.Error("error message is empty")
		}
	})
}

// TestParseURL_NegativeWordCount tests that a negative word count in the
// defuddle result is rejected by Validate().
func TestParseURL_NegativeWordCount(t *testing.T) {
	withMockBin(t, func() {
		_, err := ParseURL("negativewords://example.com/neg")
		if err == nil {
			t.Fatal("ParseURL expected error for negative word count, got nil")
		}
	})
}

// TestParseURL_MalformedJSON tests that when defuddle outputs malformed JSON,
// ParseURL falls back to treating it as plain markdown.
func TestParseURL_MalformedJSON(t *testing.T) {
	withMockBin(t, func() {
		result, err := ParseURL("malformed://example.com/bad")
		if err != nil {
			t.Fatalf("ParseURL unexpected error for malformed JSON: %v", err)
		}
		if result == nil {
			t.Fatal("result is nil")
		}
		if result.Markdown == "" {
			t.Error("Markdown should be populated from plain-text fallback")
		}
	})
}

// TestParseMarkdown tests the markdown-only parsing path.
func TestParseMarkdown(t *testing.T) {
	withMockBin(t, func() {
		md, err := ParseMarkdown("valid://example.com/page")
		if err != nil {
			t.Fatalf("ParseMarkdown unexpected error: %v", err)
		}
		if md == "" {
			t.Error("Markdown is empty")
		}
	})
}

// TestIsInstalled_Mock verifies IsInstalled returns true when the mock binary is in PATH.
func TestIsInstalled_Mock(t *testing.T) {
	mockPath := ensureMockBin(t)
	origPath := os.Getenv("PATH")
	mockDir := filepath.Dir(mockPath)
	os.Setenv("PATH", mockDir) // only mock in PATH — no fallback
	defer os.Setenv("PATH", origPath)

	if !IsInstalled() {
		t.Error("IsInstalled() = false, want true (mock binary in PATH)")
	}
}

// TestIsInstalled_NotPresent verifies IsInstalled returns false when nothing is in PATH.
func TestIsInstalled_NotPresent(t *testing.T) {
	ResetBinaryCache()
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", "/nonexistent")
	defer os.Setenv("PATH", origPath)

	if IsInstalled() {
		t.Error("IsInstalled() = true, want false (no binary in PATH)")
	}
}

// TestParseURL_NotInstalled verifies ParseURL returns a descriptive error when
// no defuddle binary is found.
func TestParseURL_NotInstalled(t *testing.T) {
	ResetBinaryCache()
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", "/nonexistent")
	defer os.Setenv("PATH", origPath)

	_, err := ParseURL("https://example.com/page")
	if err == nil {
		t.Fatal("ParseURL expected error when defuddle not installed, got nil")
	}
	if err.Error() == "" {
		t.Error("error message is empty")
	}
}