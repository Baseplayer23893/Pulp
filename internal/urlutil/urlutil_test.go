package urlutil

import (
	"testing"
)

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"bare domain", "example.com", "https://example.com", false},
		{"with path", "example.com/path/to/page", "https://example.com/path/to/page", false},
		{"http URL", "http://example.com", "http://example.com", false},
		{"https URL", "https://example.com", "https://example.com", false},
		{"empty", "", "", true},
		{"spaces", "example com", "", true},
		{"file path", "/tmp/file.md", "", true},
		{"windows path", "C:\\Users\\test", "", true},
		{"relative path", "./local/file.txt", "", true},
		{"localhost port", "localhost:8080", "https://localhost:8080", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NormalizeURL(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("NormalizeURL(%q) wanted error, got %q", tc.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("NormalizeURL(%q) unexpected error: %v", tc.input, err)
			}
			if got != tc.want {
				t.Fatalf("NormalizeURL(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
