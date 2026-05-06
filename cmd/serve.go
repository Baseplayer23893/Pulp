package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

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

	mux.HandleFunc("/", handleDashboard)
	mux.HandleFunc("/api/extract", handleAPIExtract)
	mux.HandleFunc("/health", handleHealth)

	addr := fmt.Sprintf(":%d", servePort)
	fmt.Fprintf(os.Stderr, "🍊 Pulp dashboard running at http://localhost%s\n", addr)
	fmt.Fprintf(os.Stderr, "   Press Ctrl+C to stop\n")

	return http.ListenAndServe(addr, mux)
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, dashboardHTML)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"status":"ok","version":"`+version+`"}`)
}

func handleAPIExtract(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	url := r.FormValue("url")
	if url == "" {
		http.Error(w, `{"error":"url is required"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	resp := map[string]string{
		"status":  "extracting",
		"url":     url,
		"message": "Use CLI for extraction: pulp extract " + url,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, `{"error":"failed to encode response"}`, http.StatusInternalServerError)
		return
	}
}

const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Pulp Dashboard</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:system-ui,-apple-system,sans-serif;background:#0a0a0a;color:#e0e0e0;min-height:100vh;display:flex;flex-direction:column;align-items:center;justify-content:center}
.container{max-width:600px;text-align:center;padding:2rem}
h1{font-size:2.5rem;background:linear-gradient(135deg,#667eea,#764ba2);-webkit-background-clip:text;-webkit-text-fill-color:transparent;margin-bottom:1rem}
p{color:#888;font-size:1.1rem;line-height:1.6;margin-bottom:0.5rem}
code{background:#1a1a2e;padding:0.3rem 0.6rem;border-radius:6px;font-size:0.9rem;color:#667eea}
.status{margin-top:2rem;padding:1rem;background:#111;border:1px solid #222;border-radius:12px}
.badge{display:inline-block;background:#1a3a1a;color:#4ade80;padding:0.25rem 0.75rem;border-radius:999px;font-size:0.8rem}
</style>
</head>
<body>
<div class="container">
<h1>🍊 Pulp</h1>
<p>Extract clean markdown from web content</p>
<div class="status">
<span class="badge">● Running</span>
<p style="margin-top:1rem">Dashboard is running. Use the CLI for extraction:</p>
<p><code>pulp extract &lt;url&gt;</code></p>
</div>
</div>
</body>
</html>`
