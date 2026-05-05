package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install SkillForge CLI to PATH",
	Long: `Build and install the skillforge binary to ~/.local/bin.
This makes the 'skillforge' command available globally.`,
	RunE: runInstall,
}

func init() {
	rootCmd.AddCommand(installCmd)
}

func runInstall(cmd *cobra.Command, args []string) error {
	if runtime.GOOS == "windows" {
		return fmt.Errorf("automatic install not supported on Windows — build manually with: go build -o skillforge.exe")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	binDir := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create %s: %w", binDir, err)
	}

	binPath := filepath.Join(binDir, "skillforge")
	fmt.Fprintf(os.Stderr, "🔨 Building skillforge...\n")

	// Find the module root (where go.mod is)
	modRoot, err := findModuleRoot()
	if err != nil {
		return fmt.Errorf("failed to find module root: %w", err)
	}

	buildCmd := exec.Command("go", "build", "-ldflags", "-s -w", "-o", binPath, ".")
	buildCmd.Dir = modRoot
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr

	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	fmt.Fprintf(os.Stderr, "✅ Installed to: %s\n", binPath)
	fmt.Fprintf(os.Stderr, "   Make sure %s is in your PATH\n", binDir)
	return nil
}

func findModuleRoot() (string, error) {
	// Try current executable's directory first
	exe, err := os.Executable()
	if err == nil {
		dir := filepath.Dir(exe)
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
	}

	// Try current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up to find go.mod
	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return cwd, nil
}
