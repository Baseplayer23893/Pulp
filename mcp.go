package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Baseplayer23893/Pulp/internal/cleaner"
	"github.com/Baseplayer23893/Pulp/internal/defuddle"
	"github.com/Baseplayer23893/Pulp/internal/urlutil"
	"github.com/Baseplayer23893/Pulp/internal/version"
)

// ---------------------------------------------------------------------------
// JSON-RPC 2.0 types
// ---------------------------------------------------------------------------

// jsonRPCRequest represents an incoming JSON-RPC 2.0 request.
type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// jsonRPCResponse represents an outgoing JSON-RPC 2.0 response.
type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

// rpcError represents a JSON-RPC 2.0 error object.
type rpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Standard JSON-RPC 2.0 error codes.
const (
	codeParseError     = -32700
	codeInvalidRequest = -32600
	codeMethodNotFound = -32601
	codeInvalidParams  = -32602
	codeInternalError  = -32603
)

// ---------------------------------------------------------------------------
// MCP protocol types
// ---------------------------------------------------------------------------

// mcpToolDef describes a tool exposed by this MCP server.
type mcpToolDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema mcpInputSchema `json:"inputSchema"`
}

// mcpInputSchema is a minimal JSON Schema descriptor.
type mcpInputSchema struct {
	Type       string                          `json:"type"`
	Properties map[string]mcpPropertySchema    `json:"properties,omitempty"`
	Required   []string                        `json:"required,omitempty"`
}

// mcpPropertySchema describes a single property in the input schema.
type mcpPropertySchema struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// mcpContent is the content block returned by tools/call.
type mcpContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ---------------------------------------------------------------------------
// Tool registry
// ---------------------------------------------------------------------------

// availableTools lists every tool the MCP server advertises.
var availableTools = []mcpToolDef{
	{
		Name:        "extract_content",
		Description: "Extract clean, token-efficient markdown from any URL. Supports web pages, articles, and YouTube videos (via transcript extraction). Returns structured markdown with metadata.",
		InputSchema: mcpInputSchema{
			Type: "object",
			Properties: map[string]mcpPropertySchema{
				"url": {
					Type:        "string",
					Description: "The URL to extract content from (web page, article, or YouTube video)",
				},
			},
			Required: []string{"url"},
		},
	},
}

// ---------------------------------------------------------------------------
// Core MCP server loop
// ---------------------------------------------------------------------------

// RunMCP starts the MCP server, reading newline-delimited JSON-RPC 2.0
// messages from stdin and writing responses to stdout. This allows IDEs
// (Cursor, VS Code, Windsurf, etc.) to use Pulp as a tool via the Model
// Context Protocol.
func RunMCP() {
	scanner := bufio.NewScanner(os.Stdin)

	// Increase the default buffer to handle large JSON-RPC payloads.
	const maxTokenSize = 1024 * 1024 // 1 MB
	scanner.Buffer(make([]byte, 0, maxTokenSize), maxTokenSize)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req jsonRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			writeResponse(jsonRPCResponse{
				JSONRPC: "2.0",
				Error: &rpcError{
					Code:    codeParseError,
					Message: "Parse error: " + err.Error(),
				},
			})
			continue
		}

		// Validate basic JSON-RPC 2.0 envelope.
		if req.JSONRPC != "2.0" {
			writeResponse(jsonRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error: &rpcError{
					Code:    codeInvalidRequest,
					Message: "Invalid request: jsonrpc field must be \"2.0\"",
				},
			})
			continue
		}

		if req.Method == "" {
			writeResponse(jsonRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error: &rpcError{
					Code:    codeInvalidRequest,
					Message: "Invalid request: method field is required",
				},
			})
			continue
		}

		resp := handleRequest(req)

		// Notifications (no ID) do not receive a response per JSON-RPC 2.0.
		if req.ID == nil && resp.Error == nil {
			continue
		}

		writeResponse(resp)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "mcp: scanner error: %v\n", err)
		os.Exit(1)
	}
}

// ---------------------------------------------------------------------------
// Request dispatcher
// ---------------------------------------------------------------------------

// handleRequest dispatches an incoming JSON-RPC request to the right handler.
func handleRequest(req jsonRPCRequest) jsonRPCResponse {
	switch req.Method {

	// ── MCP lifecycle ────────────────────────────────────────────────────
	case "initialize":
		return handleInitialize(req)

	case "notifications/initialized":
		// Client signals initialization is complete — no response needed.
		return jsonRPCResponse{JSONRPC: "2.0"}

	// ── Tool discovery ───────────────────────────────────────────────────
	case "tools/list":
		return handleToolsList(req)

	// ── Tool execution ───────────────────────────────────────────────────
	case "tools/call":
		return handleCallTool(req)

	// ── Unknown method ───────────────────────────────────────────────────
	default:
		return jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error: &rpcError{
				Code:    codeMethodNotFound,
				Message: fmt.Sprintf("Method not found: %s", req.Method),
			},
		}
	}
}

// ---------------------------------------------------------------------------
// MCP method handlers
// ---------------------------------------------------------------------------

// handleInitialize responds with the server's capabilities and metadata.
func handleInitialize(req jsonRPCRequest) jsonRPCResponse {
	return jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]bool{"listChanged": false},
			},
			"serverInfo": map[string]string{
				"name":    "pulp",
				"version": version.Version,
			},
		},
	}
}

// handleToolsList returns the set of tools this server offers.
func handleToolsList(req jsonRPCRequest) jsonRPCResponse {
	return jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"tools": availableTools,
		},
	}
}

// handleCallTool dispatches a tools/call request to the appropriate tool.
func handleCallTool(req jsonRPCRequest) jsonRPCResponse {
	// Parse the envelope: { name, arguments }.
	var params struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}

	if req.Params == nil {
		return rpcErrorResponse(req.ID, codeInvalidParams, "Invalid params: params field is required")
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		return rpcErrorResponse(req.ID, codeInvalidParams, "Invalid params: "+err.Error())
	}

	switch params.Name {
	case "extract_content":
		return toolExtractContent(req.ID, params.Arguments)
	default:
		return rpcErrorResponse(req.ID, codeInvalidParams,
			fmt.Sprintf("Unknown tool: %q — available tools: extract_content", params.Name))
	}
}

// ---------------------------------------------------------------------------
// Tool implementations
// ---------------------------------------------------------------------------

// toolExtractContent runs Pulp's extraction pipeline on a URL and returns
// the cleaned markdown as an MCP text content block.
func toolExtractContent(id json.RawMessage, rawArgs json.RawMessage) jsonRPCResponse {
	// Parse arguments.
	var args struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(rawArgs, &args); err != nil {
		return rpcErrorResponse(id, codeInvalidParams, "Invalid arguments: "+err.Error())
	}
	if args.URL == "" {
		return rpcErrorResponse(id, codeInvalidParams, "Invalid arguments: \"url\" is required")
	}
	if normalizedURL, err := urlutil.NormalizeURL(args.URL); err != nil {
		return toolErrorResult(id, fmt.Sprintf("Invalid URL: %s", err))
	} else {
		args.URL = normalizedURL
	}

	// Check that defuddle is available.
	if !defuddle.IsInstalled() {
		return toolErrorResult(id, "defuddle is not installed — run: npm install -g defuddle")
	}

	start := time.Now()

	// Run Pulp's extraction pipeline (same path the CLI/TUI uses).
	result, err := defuddle.ParseURL(args.URL)
	if err != nil {
		return toolErrorResult(id, fmt.Sprintf("Extraction failed: %v", err))
	}

	markdown := result.Markdown
	if markdown == "" {
		markdown = result.Content
	}
	if markdown == "" {
		return toolErrorResult(id, fmt.Sprintf("No content extracted from %s", args.URL))
	}

	// Clean the markdown (strip tracking params, normalize whitespace, etc.).
	markdown = cleaner.Clean(markdown)

	// Build a rich header with metadata when available.
	var sb strings.Builder

	if result.Title != "" {
		sb.WriteString(fmt.Sprintf("# %s\n\n", result.Title))
	}

	hasMeta := result.Author != "" || result.Domain != "" || result.Published != ""
	if hasMeta {
		if result.Author != "" {
			sb.WriteString(fmt.Sprintf("**Author:** %s  \n", result.Author))
		}
		if result.Domain != "" {
			sb.WriteString(fmt.Sprintf("**Source:** %s  \n", result.Domain))
		}
		if result.Published != "" {
			sb.WriteString(fmt.Sprintf("**Published:** %s  \n", result.Published))
		}
		sb.WriteString("\n---\n\n")
	}

	sb.WriteString(markdown)

	elapsed := time.Since(start).Round(time.Millisecond)
	wordCount := len(strings.Fields(markdown))
	sb.WriteString(fmt.Sprintf("\n\n---\n*Extracted %d words in %s via Pulp*\n", wordCount, elapsed))

	return jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]interface{}{
			"content": []mcpContent{
				{
					Type: "text",
					Text: sb.String(),
				},
			},
			"isError": false,
		},
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// toolErrorResult returns a valid MCP tools/call response with isError: true.
// This follows the MCP spec — tool-level errors are reported in the content
// block, not as JSON-RPC errors, so the IDE can display them gracefully.
func toolErrorResult(id json.RawMessage, message string) jsonRPCResponse {
	return jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]interface{}{
			"content": []mcpContent{
				{Type: "text", Text: message},
			},
			"isError": true,
		},
	}
}

// rpcErrorResponse builds a JSON-RPC 2.0 error response.
func rpcErrorResponse(id json.RawMessage, code int, message string) jsonRPCResponse {
	return jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &rpcError{
			Code:    code,
			Message: message,
		},
	}
}

// writeResponse marshals a JSON-RPC response and writes it to stdout,
// followed by a newline delimiter.
func writeResponse(resp jsonRPCResponse) {
	data, err := json.Marshal(resp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mcp: failed to marshal response: %v\n", err)
		return
	}
	fmt.Fprintln(os.Stdout, string(data))
}
