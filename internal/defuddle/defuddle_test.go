package defuddle

import (
	"strings"
	"testing"
)

func TestResultValidate(t *testing.T) {
	tests := []struct {
		name   string
		result Result
		wantOk bool
	}{
		{name: "valid with markdown", result: Result{Markdown: "hello world", WordCount: 2}, wantOk: true},
		{name: "valid with content only", result: Result{Content: "hello world", WordCount: 2}, wantOk: true},
		{name: "valid with title only", result: Result{Title: "Hello World"}, wantOk: true},
		{name: "valid with all fields", result: Result{Title: "t", Markdown: "m", Content: "c", WordCount: 10}, wantOk: true},
		{name: "zero result is invalid", result: Result{}, wantOk: false},
		{name: "empty strings are invalid", result: Result{Markdown: "", Content: ""}, wantOk: false},
		{name: "negative word count is invalid", result: Result{Markdown: "hello", WordCount: -1}, wantOk: false},
		{name: "zero word count with content is valid", result: Result{Markdown: "hello", WordCount: 0}, wantOk: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.result.Validate()
			if tc.wantOk && err != nil {
				t.Fatalf("Validate() unexpected error: %v", err)
			}
			if !tc.wantOk && err == nil {
				t.Fatalf("Validate() got nil, want error")
			}
			if !tc.wantOk && tc.result.WordCount < 0 && !strings.Contains(err.Error(), "negative") {
				t.Fatalf("Validate() error %q does not mention negative word count", err)
			}
		})
	}
}