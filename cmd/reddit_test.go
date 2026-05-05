package cmd

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNormalizeRedditURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "www.reddit.com standard",
			input:    "https://www.reddit.com/r/golang/comments/abc123/my_post/",
			expected: "https://old.reddit.com/r/golang/comments/abc123/my_post.json",
		},
		{
			name:     "new.reddit.com",
			input:    "https://new.reddit.com/r/golang/comments/abc123/my_post",
			expected: "https://old.reddit.com/r/golang/comments/abc123/my_post.json",
		},
		{
			name:     "old.reddit.com already",
			input:    "https://old.reddit.com/r/golang/comments/abc123/my_post",
			expected: "https://old.reddit.com/r/golang/comments/abc123/my_post.json",
		},
		{
			name:     "bare reddit.com",
			input:    "https://reddit.com/r/golang/comments/abc123/my_post/",
			expected: "https://old.reddit.com/r/golang/comments/abc123/my_post.json",
		},
		{
			name:     "with query params stripped",
			input:    "https://www.reddit.com/r/golang/comments/abc123/my_post?utm_source=share&context=3",
			expected: "https://old.reddit.com/r/golang/comments/abc123/my_post.json",
		},
		{
			name:     "already has .json",
			input:    "https://old.reddit.com/r/golang/comments/abc123/my_post.json",
			expected: "https://old.reddit.com/r/golang/comments/abc123/my_post.json",
		},
		{
			name:     "http scheme",
			input:    "http://www.reddit.com/r/golang/comments/abc123/my_post",
			expected: "https://old.reddit.com/r/golang/comments/abc123/my_post.json",
		},
		{
			name:     "with trailing whitespace",
			input:    "  https://www.reddit.com/r/golang/comments/abc123/my_post  ",
			expected: "https://old.reddit.com/r/golang/comments/abc123/my_post.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeRedditURL(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeRedditURL(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCleanRedditMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "HTML entities",
			input:    "this &amp; that &lt;script&gt;",
			expected: "this & that <script>",
		},
		{
			name:     "schemeless links",
			input:    "[example](//example.com/page)",
			expected: "[example](https://example.com/page)",
		},
		{
			name:     "excessive newlines",
			input:    "line 1\n\n\n\n\nline 2",
			expected: "line 1\n\nline 2",
		},
		{
			name:     "nbsp entities",
			input:    "hello&nbsp;world",
			expected: "hello world",
		},
		{
			name:     "clean text unchanged",
			input:    "This is normal markdown with **bold** and *italic*.",
			expected: "This is normal markdown with **bold** and *italic*.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanRedditMarkdown(tt.input)
			if result != tt.expected {
				t.Errorf("cleanRedditMarkdown(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRedditRepliesUnmarshal(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantEmpty bool
	}{
		{
			name:      "empty string replies",
			input:     `""`,
			wantEmpty: true,
		},
		{
			name:      "null replies",
			input:     `null`,
			wantEmpty: true,
		},
		{
			name:      "object replies with children",
			input:     `{"kind":"Listing","data":{"children":[{"kind":"t1","data":{"body":"hello","author":"testuser","score":5}}]}}`,
			wantEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var replies RedditReplies
			err := json.Unmarshal([]byte(tt.input), &replies)
			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}
			isEmpty := len(replies.Data.Children) == 0
			if isEmpty != tt.wantEmpty {
				t.Errorf("isEmpty=%v, wantEmpty=%v, children=%d", isEmpty, tt.wantEmpty, len(replies.Data.Children))
			}
		})
	}
}

func TestRenderComment(t *testing.T) {
	// Test basic comment rendering
	comment := RedditPost{
		Author: "testuser",
		Body:   "This is a test comment.",
		Score:  42,
	}

	var sb stringBuilderHelper
	// Set depth to 0 for top-level
	oldDepth := redditDepth
	redditDepth = 1
	defer func() { redditDepth = oldDepth }()

	renderComment(&sb.Builder, comment, 0)
	result := sb.String()

	// Should contain author
	if !containsStr(result, "u/testuser") {
		t.Errorf("renderComment missing author, got: %s", result)
	}
	// Should contain score
	if !containsStr(result, "42 pts") {
		t.Errorf("renderComment missing score, got: %s", result)
	}
	// Should contain body
	if !containsStr(result, "This is a test comment.") {
		t.Errorf("renderComment missing body, got: %s", result)
	}
}

func TestRenderCommentNested(t *testing.T) {
	// Build a comment with a reply
	reply := RedditPost{
		Author: "replier",
		Body:   "Nice point!",
		Score:  10,
	}

	replyChild := RedditChild{Kind: "t1", Data: reply}

	parent := RedditPost{
		Author: "parent_user",
		Body:   "Original comment.",
		Score:  50,
		Replies: RedditReplies{
			Data: struct {
				Children []RedditChild `json:"children"`
			}{
				Children: []RedditChild{replyChild},
			},
		},
	}

	var sb stringBuilderHelper
	oldDepth := redditDepth
	redditDepth = 2
	defer func() { redditDepth = oldDepth }()

	renderComment(&sb.Builder, parent, 0)
	result := sb.String()

	// Should contain parent
	if !containsStr(result, "u/parent_user") {
		t.Errorf("missing parent author, got: %s", result)
	}
	// Should contain nested reply with blockquote
	if !containsStr(result, "> ") {
		t.Errorf("missing blockquote for nested reply, got: %s", result)
	}
	if !containsStr(result, "replier") {
		t.Errorf("missing reply author, got: %s", result)
	}
}

// stringBuilderHelper wraps strings.Builder for test use
type stringBuilderHelper struct {
	Builder strings.Builder
}

func (s *stringBuilderHelper) String() string {
	return s.Builder.String()
}

func containsStr(haystack, needle string) bool {
	return len(haystack) > 0 && len(needle) > 0 && indexStr(haystack, needle) >= 0
}

func indexStr(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
