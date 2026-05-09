package defuddle

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

var (
	binaryPath   string
	binaryErr    error
	binaryOnce   sync.Once
)

// ResetBinaryCache clears the cached binary path (for testing)
func ResetBinaryCache() {
	binaryOnce = sync.Once{}
	binaryPath = ""
	binaryErr = nil
}

// Result holds the parsed output from defuddle
type Result struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Domain      string `json:"domain"`
	URL         string `json:"url"`
	Author      string `json:"author"`
	Published   string `json:"published"`
	Content     string `json:"content"`
	Markdown    string `json:"markdown"`
	WordCount   int    `json:"wordCount"`
}

// Validate checks that the result contains usable content. It returns an error
// if defuddle returned a structurally valid but empty response, which can happen
// when the defuddle version changes its output format or the extraction fails
// server-side without reporting an error.
func (r *Result) Validate() error {
	if r.Markdown == "" && r.Content == "" && r.Title == "" {
		return fmt.Errorf("defuddle returned empty content — the output format may have changed or the page could not be extracted")
	}
	if r.WordCount < 0 {
		return fmt.Errorf("defuddle returned negative word count (%d) — output format may have changed", r.WordCount)
	}
	return nil
}

// findBinary locates the defuddle CLI binary (cached)
func findBinary() (string, error) {
	binaryOnce.Do(func() {
		binaryPath, binaryErr = locateBinary()
	})
	return binaryPath, binaryErr
}

// locateBinary does the actual filesystem lookup
func locateBinary() (string, error) {
	// Check standard PATH first
	path, err := exec.LookPath("defuddle")
	if err == nil {
		return path, nil
	}

	// Check common install locations
	commonPaths := []string{
		"/usr/local/bin/defuddle",
		"/usr/bin/defuddle",
	}

	// Check npm global install paths
	out, err := exec.Command("npm", "root", "-g").Output()
	if err == nil {
		npmRoot := strings.TrimSpace(string(out))
		commonPaths = append(commonPaths, npmRoot+"/defuddle/dist/cli.js")
	}

	for _, p := range commonPaths {
		if _, err := exec.LookPath(p); err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("defuddle not found — install with: npm install -g defuddle")
}

// ParseURL extracts clean markdown from a URL using defuddle CLI
func ParseURL(url string) (*Result, error) {
	bin, err := findBinary()
	if err != nil {
		return nil, err
	}

	// Run defuddle parse with JSON output to get metadata + content
	cmd := exec.Command(bin, "parse", "--json", "--markdown", url)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("defuddle parse failed: %w\noutput: %s", err, string(out))
	}

	var result Result
	if err := json.Unmarshal(out, &result); err != nil {
		// If JSON parsing fails, the output is likely raw markdown.
		// Still validate that we got something.
		markdown := strings.TrimSpace(string(out))
		if markdown == "" {
			return nil, fmt.Errorf("defuddle produced no usable output")
		}
		return &Result{
			URL:      url,
			Markdown: markdown,
		}, nil
	}

	if err := result.Validate(); err != nil {
		return nil, fmt.Errorf("defuddle result invalid: %w", err)
	}

	return &result, nil
}

// ParseMarkdown extracts just the markdown content from a URL
func ParseMarkdown(url string) (string, error) {
	bin, err := findBinary()
	if err != nil {
		return "", err
	}

	cmd := exec.Command(bin, "parse", "--markdown", url)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("defuddle parse failed: %w\noutput: %s", err, string(out))
	}

	return strings.TrimSpace(string(out)), nil
}

// IsInstalled checks whether defuddle is available
func IsInstalled() bool {
	_, err := findBinary()
	return err == nil
}
