package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Baseplayer23893/skillforge/internal/storage"
	"github.com/spf13/cobra"
)

var packageCmd = &cobra.Command{
	Use:   "package <name>",
	Short: "Create a skill.zip from extracted content",
	Long: `Package extracted markdown content into a skill.zip archive.
The archive contains SKILL.md with frontmatter and a references/ directory.

Usage:
  cat content.md | pulp package my-skill
  pulp package my-skill -o ./output/`,
	Args: cobra.ExactArgs(1),
	RunE: runPackage,
}

var (
	packageRefs   []string
	packageSource string
)

func init() {
	packageCmd.Flags().StringSliceVarP(&packageRefs, "references", "r", nil, "Reference files to include")
	packageCmd.Flags().StringVarP(&packageSource, "source", "s", "", "Source file to package (alternative to stdin pipe)")
	rootCmd.AddCommand(packageCmd)
}

func runPackage(cmd *cobra.Command, args []string) error {
	name := args[0]
	if !quietFlag {
		fmt.Fprintf(os.Stderr, "📦 Packaging skill: %s\n", name)
	}

	// Read content from --source file, stdin pipe, or error
	var content string
	if packageSource != "" {
		data, err := os.ReadFile(packageSource)
		if err != nil {
			return fmt.Errorf("failed to read source file %s: %w", packageSource, err)
		}
		content = string(data)
	} else {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			data, err := readStdin()
			if err != nil {
				return fmt.Errorf("failed to read stdin: %w", err)
			}
			content = string(data)
		} else {
			return fmt.Errorf("no content provided — use --source or pipe:\n  pulp package %s --source content.md\n  pulp extract <url> | pulp package %s", name, name)
		}
	}

	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("no content to package")
	}

	// Add frontmatter if not already present
	if !strings.HasPrefix(content, "---") {
		frontmatter := storage.GenerateFrontmatter(name, "", "", nil)
		content = frontmatter + content
	}

	outDir := "."
	if outputFlag != "" {
		outDir = outputFlag
	}

	zipPath, err := storage.CreateSkillZip(name, content, packageRefs, outDir)
	if err != nil {
		return fmt.Errorf("failed to create package: %w", err)
	}

	absPath, _ := filepath.Abs(zipPath)
	if !quietFlag {
		fmt.Fprintf(os.Stderr, "✅ Created: %s\n", absPath)
	}
	return nil
}

func readStdin() ([]byte, error) {
	var data []byte
	buf := make([]byte, 4096)
	for {
		n, err := os.Stdin.Read(buf)
		if n > 0 {
			data = append(data, buf[:n]...)
		}
		if err != nil {
			break
		}
	}
	return data, nil
}
