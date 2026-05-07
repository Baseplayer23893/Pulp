package urlutil

import (
	"fmt"
	"net/url"
	"strings"
)

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

	if len(raw) > 1 && raw[1] == ':' &&
		((raw[0] >= 'A' && raw[0] <= 'Z') || (raw[0] >= 'a' && raw[0] <= 'z')) {
		return "", fmt.Errorf("looks like a file path, not a web URL")
	}

	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		u, err := url.Parse(raw)
		if err != nil || u.Host == "" {
			return "", fmt.Errorf("invalid URL format")
		}
		return raw, nil
	}

	full := "https://" + raw
	u, err := url.Parse(full)
	if err != nil || u.Host == "" {
		return "", fmt.Errorf("invalid URL format")
	}
	return full, nil
}
