package cmd

import (
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Baseplayer23893/Pulp/internal/config"
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
	return filepath.Join(outDir, sanitizedOutputName(source)+ext)
}

func sanitizedOutputName(source string) string {
	raw := strings.TrimSpace(source)
	if raw == "" {
		return "output"
	}

	if parsed, err := url.Parse(raw); err == nil && parsed.Host != "" {
		p := strings.Trim(parsed.Path, "/")
		if p == "" {
			raw = parsed.Host
		} else {
			raw = filepath.Base(p)
		}
	}

	re := regexp.MustCompile(`[^a-zA-Z0-9._-]+`)
	raw = re.ReplaceAllString(raw, "-")
	raw = strings.Trim(raw, "-._")
	if raw == "" {
		return "output"
	}
	return raw
}
