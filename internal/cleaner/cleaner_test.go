package cleaner

import (
	"strings"
	"testing"
)

func TestClean(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normalizes line endings",
			input:    "hello\r\nworld\rfoo",
			expected: "hello\nworld\nfoo\n",
		},
		{
			name:     "collapses multiple blank lines",
			input:    "hello\n\n\n\n\nworld",
			expected: "hello\n\nworld\n",
		},
		{
			name:     "removes empty links",
			input:    "text [](http://example.com) more",
			expected: "text  more\n",
		},
		{
			name:     "trims whitespace",
			input:    "  \n  hello  \n  ",
			expected: "hello\n",
		},
		{
			name:     "ensures trailing newline",
			input:    "hello world",
			expected: "hello world\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Clean(tt.input)
			if result != tt.expected {
				t.Errorf("Clean(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCleanTrackingURLs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "removes utm params",
			input:    "[link](https://example.com?utm_source=twitter&utm_medium=social)",
			expected: "[link](https://example.com)",
		},
		{
			name:     "removes fbclid",
			input:    "[link](https://example.com/page?fbclid=abc123&other=keep)",
			expected: "[link](https://example.com/page?other=keep)",
		},
		{
			name:     "keeps clean URLs unchanged",
			input:    "[link](https://example.com/page)",
			expected: "[link](https://example.com/page)",
		},
		{
			name:     "handles multiple links",
			input:    "[a](https://a.com?utm_source=x) and [b](https://b.com?gclid=y)",
			expected: "[a](https://a.com) and [b](https://b.com)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanTrackingURLs(tt.input)
			if result != tt.expected {
				t.Errorf("cleanTrackingURLs(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestAddFrontmatter(t *testing.T) {
	content := "# Hello\n\nWorld"
	meta := map[string]string{
		"title":  "Test",
		"source": "https://example.com",
	}

	result := AddFrontmatter(content, meta)

	if !strings.HasPrefix(result, "---\n") {
		t.Error("frontmatter should start with ---")
	}
	if !strings.Contains(result, "title: Test") {
		t.Error("frontmatter should contain title")
	}
	if !strings.Contains(result, "source: https://example.com") {
		t.Error("frontmatter should contain source")
	}
	if !strings.Contains(result, "---\n\n# Hello") {
		t.Error("frontmatter should be followed by content")
	}
}
