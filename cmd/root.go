package cmd

import (
	"fmt"

	"github.com/Baseplayer23893/skillforge/cmd/tui"
	"github.com/Baseplayer23893/skillforge/internal/config"
	"github.com/spf13/cobra"
)

const version = "0.2.0"

var (
	outputFlag string
	formatFlag string
	quietFlag  bool
)

var rootCmd = &cobra.Command{
	Use:   "pulp",
	Short: "Pulp — squeeze the web into clean markdown",
	Long: `Pulp is an open-source tool that extracts clean markdown from web content
and packages it for AI workflows, custom agents, and local LLM pipelines.

Supported sources: web pages, YouTube, Instagram, Reddit, PDFs.
Uses defuddle under the hood for high-quality content extraction.

Run 'pulp tui' for the interactive terminal UI.`,
	Version: version,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

// TUI command
var tuiCmd = &cobra.Command{
	Use:     "tui",
	Aliases: []string{"ui", "menu"},
	Short:   "Launch interactive terminal UI",
	Long: `Launch the Pulp interactive terminal UI.
Select sources, enter URLs, preview results, and manage your squeezes
all from a beautiful terminal dashboard.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.ShowMenu()
	},
}

// Dashboard command
var dashboardCmd = &cobra.Command{
	Use:     "dashboard",
	Aliases: []string{"dash", "history"},
	Short:   "View extraction history and stats",
	RunE: func(cmd *cobra.Command, args []string) error {
		return tui.ShowDashboard()
	},
}

// Config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show or modify Pulp configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Load()
		fmt.Printf("Config file:    %s\n", config.ConfigPath())
		fmt.Printf("Output dir:     %s\n", cfg.OutputDir)
		fmt.Printf("Default format: %s\n", cfg.DefaultFormat)
		fmt.Printf("History file:   %s\n", cfg.HistoryFile)
		fmt.Printf("Max history:    %d\n", cfg.MaxHistory)
		return nil
	},
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create default config file at ~/.pulp.yaml",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.DefaultConfig()
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to create config: %w", err)
		}
		fmt.Printf("✅ Config created: %s\n", config.ConfigPath())
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a config value (output_dir, default_format)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key, value := args[0], args[1]
		cfg := config.Load()

		switch key {
		case "output_dir":
			cfg.OutputDir = value
		case "default_format":
			if value != "md" && value != "skillzip" && value != "single" {
				return fmt.Errorf("invalid format %q — use md, skillzip, or single", value)
			}
			cfg.DefaultFormat = value
		default:
			return fmt.Errorf("unknown config key %q — use output_dir or default_format", key)
		}

		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Printf("✅ Set %s = %s\n", key, value)
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Load config defaults for format flag
	cfg := config.Load()

	rootCmd.PersistentFlags().StringVarP(&outputFlag, "output", "o", "", "Output file location")
	rootCmd.PersistentFlags().StringVarP(&formatFlag, "format", "f", cfg.DefaultFormat, "Output format: md, skillzip, single")
	rootCmd.PersistentFlags().BoolVarP(&quietFlag, "quiet", "q", false, "Suppress verbose output")

	rootCmd.SetVersionTemplate(fmt.Sprintf("Pulp v%s\n", version))

	// Register TUI commands
	rootCmd.AddCommand(tuiCmd)
	rootCmd.AddCommand(dashboardCmd)

	// Register config commands
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configCmd)
}
