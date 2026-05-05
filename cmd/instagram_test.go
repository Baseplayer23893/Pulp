package cmd

import (
	"testing"
)

func TestNormalizeInstagramURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard reel URL",
			input:    "https://www.instagram.com/reel/ABC123/",
			expected: "https://www.instagram.com/reel/ABC123/",
		},
		{
			name:     "URL with tracking params",
			input:    "https://www.instagram.com/reel/ABC123/?igsh=xyz&utm_source=share",
			expected: "https://www.instagram.com/reel/ABC123/",
		},
		{
			name:     "URL without scheme",
			input:    "reel/ABC123/",
			expected: "https://www.instagram.com/reel/ABC123/",
		},
		{
			name:     "post URL",
			input:    "https://www.instagram.com/p/ABC123/",
			expected: "https://www.instagram.com/p/ABC123/",
		},
		{
			name:     "URL with query only",
			input:    "https://www.instagram.com/reel/ABC123?utm_medium=copy_link",
			expected: "https://www.instagram.com/reel/ABC123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeInstagramURL(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeInstagramURL(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatInstagramCaption(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple caption",
			input:    "Hello world",
			expected: "Hello world",
		},
		{
			name:     "caption with mentions",
			input:    "Thanks @johndoe for this!",
			expected: "Thanks **@johndoe** for this!",
		},
		{
			name:     "caption with multiple mentions",
			input:    "@alice and @bob collab",
			expected: "**@alice** and **@bob** collab",
		},
		{
			name:     "multiline caption with empty lines",
			input:    "Line one\n\n\nLine two\n\nLine three",
			expected: "Line one\n\nLine two\n\nLine three",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatInstagramCaption(tt.input)
			if result != tt.expected {
				t.Errorf("formatInstagramCaption(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExtractHashtags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "no hashtags",
			input:    "Hello world",
			expected: nil,
		},
		{
			name:     "single hashtag",
			input:    "Check this out #coding",
			expected: []string{"#coding"},
		},
		{
			name:     "multiple hashtags",
			input:    "#go #rust #python",
			expected: []string{"#go", "#rust", "#python"},
		},
		{
			name:     "duplicate hashtags case-insensitive",
			input:    "#Go #go #GO",
			expected: []string{"#Go"},
		},
		{
			name:     "hashtags mixed with text",
			input:    "Learning #AI and #MachineLearning today!",
			expected: []string{"#AI", "#MachineLearning"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractHashtags(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("extractHashtags(%q) returned %d tags, want %d: %v", tt.input, len(result), len(tt.expected), result)
				return
			}
			for i, tag := range result {
				if tag != tt.expected[i] {
					t.Errorf("extractHashtags(%q)[%d] = %q, want %q", tt.input, i, tag, tt.expected[i])
				}
			}
		})
	}
}

func TestCoalesce(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{"all empty", []string{"", "", ""}, "unknown"},
		{"first non-empty", []string{"hello", "world"}, "hello"},
		{"skip empty", []string{"", "world"}, "world"},
		{"single value", []string{"only"}, "only"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := coalesce(tt.input...)
			if result != tt.expected {
				t.Errorf("coalesce(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatCount(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected string
	}{
		{"small number", 42, "42"},
		{"thousands", 1500, "1500"},
		{"ten thousands", 15000, "15.0K"},
		{"hundred thousands", 150000, "150.0K"},
		{"millions", 1500000, "1.5M"},
		{"zero", 0, "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatCount(tt.input)
			if result != tt.expected {
				t.Errorf("formatCount(%d) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
