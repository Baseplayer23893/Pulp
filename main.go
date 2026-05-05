package main

import (
	"os"

	"github.com/Baseplayer23893/skillforge/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
