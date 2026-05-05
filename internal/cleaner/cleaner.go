package cleaner

import (
	"net/url"
	"regexp"
	"strings"
)

var (
	// Match multiple consecutive blank lines
	multiBlankLines = regexp.MustCompile(`\n{3,}`)
	// Match tracking parameters in URLs
	trackingParams = []string{
		"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content",
		"fbclid", "gclid", "ref", "source", "mc_cid", "mc_eid",
	}
	// Match empty markdown links like [](...)
	emptyLinks = regexp.MustCompile(`\[\s*\]\([^)]*\)`)
)

// Clean performs post-processing on extracted markdown content
func Clean(markdown string) string {
	result := markdown

	// Normalize line endings
	result = strings.ReplaceAll(result, "\r\n", "\n")
	result = strings.ReplaceAll(result, "\r", "\n")

	// Remove empty links
	result = emptyLinks.ReplaceAllString(result, "")

	// Clean tracking parameters from URLs in markdown
	result = cleanTrackingURLs(result)

	// Collapse multiple blank lines to maximum 2
	result = multiBlankLines.ReplaceAllString(result, "\n\n")

	// Trim leading/trailing whitespace
	result = strings.TrimSpace(result)

	// Ensure file ends with newline
	if !strings.HasSuffix(result, "\n") {
		result += "\n"
	}

	return result
}

// cleanTrackingURLs removes tracking parameters from URLs in markdown
func cleanTrackingURLs(content string) string {
	// Match markdown links: [text](url)
	linkRegex := regexp.MustCompile(`\[([^\]]*)\]\(([^)]+)\)`)

	return linkRegex.ReplaceAllStringFunc(content, func(match string) string {
		parts := linkRegex.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}

		text := parts[1]
		rawURL := parts[2]

		parsed, err := url.Parse(rawURL)
		if err != nil {
			return match
		}

		query := parsed.Query()
		changed := false
		for _, param := range trackingParams {
			if query.Has(param) {
				query.Del(param)
				changed = true
			}
		}

		if !changed {
			return match
		}

		parsed.RawQuery = query.Encode()
		cleanURL := parsed.String()

		// Remove trailing ? if no query params remain
		cleanURL = strings.TrimSuffix(cleanURL, "?")

		return "[" + text + "](" + cleanURL + ")"
	})
}

// AddFrontmatter prepends YAML frontmatter to markdown content
func AddFrontmatter(content string, meta map[string]string) string {
	var sb strings.Builder
	sb.WriteString("---\n")
	for key, value := range meta {
		sb.WriteString(key)
		sb.WriteString(": ")
		sb.WriteString(value)
		sb.WriteString("\n")
	}
	sb.WriteString("---\n\n")
	sb.WriteString(content)
	return sb.String()
}
