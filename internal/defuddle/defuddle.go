package defuddle

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

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

// findBinary locates the defuddle CLI binary
func findBinary() (string, error) {
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
		// If JSON parsing fails, the output is likely raw markdown
		return &Result{
			URL:      url,
			Markdown: strings.TrimSpace(string(out)),
		}, nil
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
