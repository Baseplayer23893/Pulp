package config

import (
	"path/filepath"
	"testing"
)

func TestNormalizeOutputDirExpandsHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	got := NormalizeOutputDir("~/pulp-out")
	want := filepath.Join(home, "pulp-out")
	if got != want {
		t.Fatalf("NormalizeOutputDir()=%q, want %q", got, want)
	}
}

func TestNormalizeOutputDirExpandsEnvironment(t *testing.T) {
	base := t.TempDir()
	t.Setenv("PULP_TEST_OUT", base)

	got := NormalizeOutputDir("$PULP_TEST_OUT/nested")
	want := filepath.Join(base, "nested")
	if got != want {
		t.Fatalf("NormalizeOutputDir()=%q, want %q", got, want)
	}
}
