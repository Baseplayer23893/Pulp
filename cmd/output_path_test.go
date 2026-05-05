package cmd

import (
	"path/filepath"
	"testing"

	"github.com/Baseplayer23893/Pulp/internal/config"
)

func TestResolveOutputPathUsesConfigOutputDir(t *testing.T) {
	cfgHome := t.TempDir()
	outDir := filepath.Join(t.TempDir(), "out")
	t.Setenv("XDG_CONFIG_HOME", cfgHome)

	cfg := config.Load()
	cfg.OutputDir = outDir
	if err := cfg.Save(); err != nil {
		t.Fatalf("save config: %v", err)
	}

	got := resolveOutputPath("", "https://example.com/posts/abc?x=1", ".md")
	want := filepath.Join(outDir, "abc.md")
	if got != want {
		t.Fatalf("resolveOutputPath()=%q, want %q", got, want)
	}
}

func TestResolveOutputPathForceStdoutEnv(t *testing.T) {
	t.Setenv(forceStdoutEnv, "1")
	got := resolveOutputPath("", "https://example.com/posts/abc", ".md")
	if got != "" {
		t.Fatalf("resolveOutputPath()=%q, want empty stdout path", got)
	}
}
