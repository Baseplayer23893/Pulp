package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateSkillZipCreatesOutputDirectory(t *testing.T) {
	base := t.TempDir()
	outDir := filepath.Join(base, "nested", "output")

	zipPath, err := CreateSkillZip("My Skill", "# Hello\n", nil, outDir)
	if err != nil {
		t.Fatalf("CreateSkillZip returned error: %v", err)
	}

	if _, err := os.Stat(outDir); err != nil {
		t.Fatalf("expected output directory to exist: %v", err)
	}
	if _, err := os.Stat(zipPath); err != nil {
		t.Fatalf("expected zip file to exist: %v", err)
	}
}
