package main

import (
	"os"

	"github.com/Baseplayer23893/skillforge/cmd"
)

func main() {
	// If the user runs "pulp mcp", bypass the Cobra CLI and TUI entirely
	// and start the MCP server for IDE integration.
	if len(os.Args) >= 2 && os.Args[1] == "mcp" {
		RunMCP()
		return
	}

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
