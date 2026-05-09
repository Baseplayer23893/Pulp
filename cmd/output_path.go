package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Baseplayer23893/Pulp/internal/config"
	"github.com/Baseplayer23893/Pulp/internal/urlutil"
)

const forceStdoutEnv = "PULP_FORCE_STDOUT"

func resolveOutputPath(explicitPath string, source string, ext string) string {
	if explicitPath != "" {
		return explicitPath
	}
	if strings.TrimSpace(os.Getenv(forceStdoutEnv)) == "1" {
		return ""
	}

	outDir := config.ResolveOutputDir("")
	if outDir == "" || outDir == "." {
		return ""
	}
	return filepath.Join(outDir, urlutil.SlugFromURL(source)+ext)
}
