package urlutil

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var slugCleaner = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

// NormalizeURL normalizes a raw URL string by prepending https:// if no scheme is present.
// It returns a user-friendly error if the input is clearly not a valid web URL.
func NormalizeURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("URL is empty")
	}

	if strings.ContainsAny(raw, " \t") {
		return "", fmt.Errorf("URL contains spaces")
	}

	if strings.HasPrefix(raw, "/") || strings.HasPrefix(raw, "\\") ||
		strings.HasPrefix(raw, "./") || strings.HasPrefix(raw, "..") {
		return "", fmt.Errorf("looks like a file path, not a web URL")
	}

	// Check for Windows-style paths like C:\path or c:/path
	if len(raw) > 2 && raw[1] == ':' && (raw[2] == '\\' || raw[2] == '/') {
		return "", fmt.Errorf("looks like a file path, not a web URL")
	}

	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		u, err := url.Parse(raw)
		if err != nil || u.Host == "" {
			return "", fmt.Errorf("invalid URL: missing host %q", raw)
		}
		return raw, nil
	}

	full := "https://" + raw
	u, err := url.Parse(full)
	if err != nil || u.Host == "" {
		return "", fmt.Errorf("invalid URL: missing host %q", raw)
	}
	return full, nil
}

// SlugFromURL converts a URL or raw string into a filename-safe slug.
// If the input is empty, returns "output".
// For URLs, extracts the meaningful path component or falls back to host.
// Strips non-alphanumeric characters and trims trailing punctuation.
func SlugFromURL(raw string) string {
	candidate := strings.TrimSpace(raw)
	if candidate == "" {
		return "output"
	}

	// Try to parse as URL
	if parsed, err := url.Parse(candidate); err == nil && parsed.Host != "" {
		p := strings.Trim(parsed.Path, "/")
		if p != "" {
			last := strings.Split(p, "/")
			lastPart := last[len(last)-1]
			lastPart = strings.Split(lastPart, "?")[0]
			lastPart = strings.Split(lastPart, "#")[0]
			if lastPart != "" && lastPart != "." && lastPart != "/" {
				candidate = lastPart
			} else {
				candidate = parsed.Host
			}
		} else {
			candidate = parsed.Host
		}
	}

	candidate = slugCleaner.ReplaceAllString(candidate, "-")
	candidate = strings.Trim(candidate, "-._")
	if candidate == "" {
		return "output"
	}
	return candidate
}
