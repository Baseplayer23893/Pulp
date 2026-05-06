package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds Pulp configuration
type Config struct {
	// OutputDir is the default output directory for extracted content
	OutputDir string `json:"output_dir"`
	// DefaultFormat is the default output format (md, skillzip, single)
	DefaultFormat string `json:"default_format"`
	// MaxHistory is the maximum number of history entries to keep
	MaxHistory int `json:"max_history"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		OutputDir:     ".",
		DefaultFormat: "md",
		MaxHistory:    100,
	}
}

// ConfigDir returns the Pulp config directory using the OS-native config path.
// On Linux: ~/.config/pulp
// On macOS: ~/Library/Application Support/pulp
// On Windows: %AppData%\pulp
func ConfigDir() string {
	base, err := os.UserConfigDir()
	if err != nil {
		// fallback: use ~/.config/pulp manually
		home, herr := os.UserHomeDir()
		if herr != nil {
			return ".pulp"
		}
		return filepath.Join(home, ".config", "pulp")
	}
	return filepath.Join(base, "pulp")
}

// ConfigPath returns the path to the JSON config file inside the config dir.
func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.json")
}

// HistoryPath returns the path to the JSON history file inside the config dir.
func HistoryPath() string {
	return filepath.Join(ConfigDir(), "history.json")
}

// EnsureConfigDir creates the config directory if it doesn't exist.
func EnsureConfigDir() error {
	return os.MkdirAll(ConfigDir(), 0755)
}

// Load reads config from the config directory.
// Falls back to defaults if file doesn't exist or cannot be parsed.
func Load() *Config {
	cfg := DefaultConfig()

	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		// File doesn't exist yet — that's fine, return defaults.
		return cfg
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return DefaultConfig()
	}

	return cfg
}

// Save writes the current config as JSON to the config directory.
func (c *Config) Save() error {
	if err := EnsureConfigDir(); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(ConfigPath(), data, 0644)
}

// ResolveOutputDir returns the output directory with precedence:
// CLI flag > config > current directory
func ResolveOutputDir(cliFlag string) string {
	if cliFlag != "" {
		return NormalizeOutputDir(cliFlag)
	}

	cfg := Load()
	if cfg.OutputDir != "" && cfg.OutputDir != "." {
		return NormalizeOutputDir(cfg.OutputDir)
	}

	return "."
}

// NormalizeOutputDir expands shell-style home and environment references in an
// output directory while preserving "." as the default current-directory value.
func NormalizeOutputDir(dir string) string {
	dir = os.ExpandEnv(strings.TrimSpace(dir))
	if dir == "" || dir == "." {
		return dir
	}
	if dir == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}
	}
	if strings.HasPrefix(dir, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(dir, "~/"))
		}
	}
	return filepath.Clean(dir)
}
