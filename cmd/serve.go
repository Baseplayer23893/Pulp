package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Baseplayer23893/Pulp/internal/config"
	"github.com/Baseplayer23893/Pulp/internal/version"
	"github.com/spf13/cobra"
)

var servePort int

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start local Pulp dashboard",
	Long: `Start a local web dashboard for browsing and managing extracted skills.
The dashboard provides a visual interface for the CLI functionality.`,
	RunE: runServe,
}

func init() {
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 3777, "Port to serve on")
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/extract", handleAPIExtract)
	mux.HandleFunc("/api/history", handleAPIHistory)
	mux.HandleFunc("/api/config", handleAPIConfig)
	mux.HandleFunc("/api/health", handleHealth)

	// Serve static files from the dashboard build
	dashboardDir := filepath.Join(getExeDir(), "..", "..", "dashboard-dist")
	if _, err := os.Stat(dashboardDir); os.IsNotExist(err) {
		// Also try relative to cwd
		if _, err := os.Stat("dashboard-dist"); err == nil {
			dashboardDir = "dashboard-dist"
		}
	}
	fs := http.FileServer(http.Dir(dashboardDir))
	mux.Handle("/", fs)

	addr := fmt.Sprintf(":%d", servePort)
	fmt.Fprintf(os.Stderr, "🍊 Pulp dashboard running at http://localhost%s\n", addr)
	fmt.Fprintf(os.Stderr, "   Press Ctrl+C to stop\n")

	return http.ListenAndServe(addr, mux)
}

func getExeDir() string {
	exe, _ := os.Executable()
	return filepath.Dir(exe)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"ok","version":"%s"}`, version.Version)
}

type ExtractRequest struct {
	URL     string `json:"url"`
	Format  string `json:"format"`
}

type ExtractResponse struct {
	Status    string `json:"status"`
	URL       string `json:"url,omitempty"`
	Title     string `json:"title,omitempty"`
	Markdown  string `json:"markdown,omitempty"`
	WordCount int    `json:"wordCount,omitempty"`
	OutputPath string `json:"outputPath,omitempty"`
	Error     string `json:"error,omitempty"`
}

func handleAPIExtract(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req ExtractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, `{"error":"url is required"}`, http.StatusBadRequest)
		return
	}

	// Find the pulp executable
	pulpBin := findPulpBinary()
	if pulpBin == "" {
		json.NewEncoder(w).Encode(ExtractResponse{
			Status: "error",
			Error:  "pulp binary not found",
		})
		return
	}

	// Determine output path
	cfg := config.Load()
	outDir := cfg.OutputDir
	if outDir == "" {
		outDir = filepath.Join(os.Getenv("HOME"), "pulp-output")
	}
	os.MkdirAll(outDir, 0755)

	outputPath := filepath.Join(outDir, sanitizeURL(req.URL)+".md")

	// Run extraction
	format := req.Format
	if format == "" {
		format = "md"
	}

	cmd := exec.Command(pulpBin, "extract", req.URL, "-o", outputPath, "-f", format, "-q")
	output, err := cmd.CombinedOutput()

	if err != nil {
		json.NewEncoder(w).Encode(ExtractResponse{
			Status:  "error",
			URL:     req.URL,
			Error:   strings.TrimSpace(string(output)),
		})
		return
	}

	// Read the output file
	content, err := os.ReadFile(outputPath)
	if err != nil {
		json.NewEncoder(w).Encode(ExtractResponse{
			Status:  "error",
			URL:     req.URL,
			Error:   "extraction succeeded but could not read output: " + err.Error(),
		})
		return
	}

	markdown := string(content)
	wordCount := len(strings.Fields(markdown))

	// Extract title from frontmatter or first line
	title := extractTitle(markdown)

	json.NewEncoder(w).Encode(ExtractResponse{
		Status:    "success",
		URL:       req.URL,
		Title:     title,
		Markdown:  markdown,
		WordCount: wordCount,
		OutputPath: outputPath,
	})
}

func findPulpBinary() string {
	// Check current directory first
	if _, err := os.Stat("./pulp"); err == nil {
		return "./pulp"
	}
	if _, err := os.Stat("./pulp.exe"); err == nil {
		return "./pulp.exe"
	}

	// Check PATH
	if path, err := exec.LookPath("pulp"); err == nil {
		return path
	}

	// Check dir of current executable
	exe, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(exe)
		if _, err := os.Stat(filepath.Join(dir, "pulp")); err == nil {
			return filepath.Join(dir, "pulp")
		}
	}

	return ""
}

func sanitizeURL(url string) string {
	// Simple sanitization - take last path component or host
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "www.")
	if idx := strings.Index(url, "/"); idx > 0 {
		url = url[:idx]
	}
	url = strings.ReplaceAll(url, "/", "-")
	url = strings.ReplaceAll(url, ".", "-")
	if len(url) > 50 {
		url = url[:50]
	}
	return url
}

func extractTitle(markdown string) string {
	// Try to get title from YAML frontmatter
	lines := strings.Split(markdown, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "title:") {
			return strings.TrimPrefix(line, "title:")
		}
		if strings.HasPrefix(line, "# ") && i < 5 {
			return strings.TrimPrefix(line, "# ")
		}
	}
	return "Untitled"
}

type HistoryItem struct {
	URL       string `json:"url"`
	Title     string `json:"title"`
	WordCount int    `json:"wordCount"`
	Timestamp string `json:"timestamp"`
	Format    string `json:"format"`
}

func handleAPIHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cfg := config.Load()
	historyPath := config.HistoryPath()

	// Try to load from history file if it exists
	if _, err := os.Stat(historyPath); err == nil {
		data, err := os.ReadFile(historyPath)
		if err == nil {
			var history struct {
				Entries []HistoryItem `json:"entries"`
			}
			if json.Unmarshal(data, &history) == nil {
				json.NewEncoder(w).Encode(history.Entries)
				return
			}
		}
	}

	// Fallback: scan output directory
	cfg = config.Load()
	outDir := cfg.OutputDir
	if outDir == "" {
		outDir = filepath.Join(os.Getenv("HOME"), "pulp-output")
	}

	entries, _ := os.ReadDir(outDir)
	var items []HistoryItem
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".md") {
			path := filepath.Join(outDir, e.Name())
			info, _ := os.Stat(path)
			items = append(items, HistoryItem{
				URL:       strings.TrimSuffix(e.Name(), ".md"),
				Title:     strings.TrimSuffix(e.Name(), ".md"),
				WordCount: 0,
				Timestamp: info.ModTime().Format("2006-01-02"),
				Format:    "md",
			})
		}
	}

	json.NewEncoder(w).Encode(items)
}

func handleAPIConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	cfg := config.Load()
	json.NewEncoder(w).Encode(map[string]string{
		"output_dir":    cfg.OutputDir,
		"default_format": cfg.DefaultFormat,
	})
}
