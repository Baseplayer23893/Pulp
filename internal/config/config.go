package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds SkillForge configuration
type Config struct {
	// OutputDir is the default output directory for extracted content
	OutputDir string `yaml:"output_dir"`
	// DefaultFormat is the default output format (md, skillzip, single)
	DefaultFormat string `yaml:"default_format"`
	// HistoryFile tracks extraction history
	HistoryFile string `yaml:"history_file"`
	// MaxHistory is the maximum number of history entries to keep
	MaxHistory int `yaml:"max_history"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		OutputDir:     ".",
		DefaultFormat: "md",
		HistoryFile:   filepath.Join(ConfigDir(), "history.json"),
		MaxHistory:    100,
	}
}

// ConfigDir returns the SkillForge config directory
func ConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".skillforge"
	}
	return filepath.Join(home, ".config", "skillforge")
}

// ConfigPath returns the path to the config file
func ConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".skillforge.yaml"
	}
	return filepath.Join(home, ".skillforge.yaml")
}

// Load reads config from ~/.skillforge.yaml
// Falls back to defaults if file doesn't exist
func Load() *Config {
	cfg := DefaultConfig()

	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		return cfg
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return DefaultConfig()
	}

	return cfg
}

// Save writes the current config to ~/.skillforge.yaml
func (c *Config) Save() error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	header := []byte("# SkillForge Configuration\n# See: https://github.com/Baseplayer23893/skillforge\n\n")
	content := append(header, data...)

	return os.WriteFile(ConfigPath(), content, 0644)
}

// ResolveOutputDir returns the output directory with precedence:
// CLI flag > config > current directory
func ResolveOutputDir(cliFlag string) string {
	if cliFlag != "" {
		return cliFlag
	}

	cfg := Load()
	if cfg.OutputDir != "" && cfg.OutputDir != "." {
		return cfg.OutputDir
	}

	return "."
}

// EnsureConfigDir creates the config directory if it doesn't exist
func EnsureConfigDir() error {
	return os.MkdirAll(ConfigDir(), 0755)
}
