package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Baseplayer23893/Pulp/internal/defuddle"
)

const mockDefuddleSource = `package main

import (
	"encoding/json"
	"os"
	"strings"
)

func main() {
	args := os.Args[1:]
	if len(args) < 2 {
		os.Exit(1)
	}
	url := args[len(args)-1]

	// Handle normalized URLs - strip https:// prefix for matching
	cleanURL := strings.TrimPrefix(url, "https://")
	cleanURL = strings.TrimPrefix(cleanURL, "http://")

	switch {
	case strings.Contains(cleanURL, "valid"):
		json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"title":      "Test Title",
			"description": "A test page",
			"domain":     "example.com",
			"url":        url,
			"author":     "Test Author",
			"content":    "Hello world, this is some test content.",
			"markdown":   "# Test Title\n\nHello world, this is some test content.",
			"wordCount":  8,
		})
	case strings.Contains(cleanURL, "empty"):
		json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"markdown":  "",
			"content":   "",
			"title":     "",
			"wordCount": 0,
		})
	default:
		json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
			"title":      "Default",
			"description": "Default description",
			"domain":     "example.com",
			"url":        url,
			"content":    "Default content",
			"markdown":   "# Default\n\nDefault content",
			"wordCount":  2,
		})
	}
}
`

func TestMCPToolsCall(t *testing.T) {
	tmpDir := t.TempDir()
	mockBin := filepath.Join(tmpDir, "defuddle")

	if err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(mockDefuddleSource), 0644); err != nil {
		t.Fatalf("write mock source: %v", err)
	}

	// Create a minimal go.mod
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module mock\ngo 1.21\n"), 0644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	cmd := exec.Command("go", "build", "-o", mockBin, ".")
	cmd.Dir = tmpDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build mock defuddle: %v\n%s", err, out)
	}

	origPath := os.Getenv("PATH")
	defer os.Setenv("PATH", origPath)
	os.Setenv("PATH", tmpDir+":"+origPath)

	defuddle.ResetBinaryCache()

	tests := []struct {
		name           string
		params         string
		wantErrCode    int
		wantContains   []string
		wantNotContain []string
	}{
		{
			name:   "valid URL returns content",
			params: `{"name":"extract_content","arguments":{"url":"valid-example.com/test"}}`,
			wantContains: []string{
				"content",
				"Test Title",
			},
		},
		{
			name:        "missing url field returns codeInvalidParams",
			params:      `{"name":"extract_content","arguments":{}}`,
			wantErrCode: -32602,
			wantContains: []string{
				"Invalid arguments",
				"url",
			},
		},
		{
			name:        "bare domain gets normalized",
			params:      `{"name":"extract_content","arguments":{"url":"example.com"}}`,
			wantContains: []string{
				"content",
				"Default",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defuddle.ResetBinaryCache()

			req := jsonRPCRequest{
				JSONRPC: "2.0",
				ID:      json.RawMessage(`1`),
				Method:  "tools/call",
				Params:  json.RawMessage(tt.params),
			}

resp := handleCallTool(req)

			if tt.wantErrCode != 0 {
				if resp.Error == nil {
					t.Errorf("expected error code %d, got nil", tt.wantErrCode)
				} else if resp.Error.Code != tt.wantErrCode {
					t.Errorf("error code = %d, want %d", resp.Error.Code, tt.wantErrCode)
				}
			}

			respJSON, _ := json.Marshal(resp)
			respStr := string(respJSON)

			for _, want := range tt.wantContains {
				if !strings.Contains(respStr, want) {
					t.Errorf("response should contain %q, got: %s", want, respStr)
				}
			}

			for _, notWant := range tt.wantNotContain {
				if strings.Contains(respStr, notWant) {
					t.Errorf("response should NOT contain %q, got: %s", notWant, respStr)
				}
			}
		})
	}
}

func TestMCPToolsCall_DefuddleNotInstalled(t *testing.T) {
	origPath := os.Getenv("PATH")
	defer os.Setenv("PATH", origPath)

	os.Setenv("PATH", "/nonexistent")

	defuddle.ResetBinaryCache()

	req := jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      json.RawMessage(`1`),
		Method:  "tools/call",
		Params:  json.RawMessage(`{"name":"extract_content","arguments":{"url":"https://example.com"}}`),
	}

	resp := handleCallTool(req)

	// Check response contains error about defuddle via JSON string
	respJSON, _ := json.Marshal(resp)
	respStr := string(respJSON)

	if !strings.Contains(respStr, `"isError":true`) {
		t.Errorf("expected isError: true, got: %s", respStr)
	}
	if !strings.Contains(respStr, "defuddle") {
		t.Errorf("expected error about defuddle, got: %s", respStr)
	}
}