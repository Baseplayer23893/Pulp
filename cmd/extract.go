package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Baseplayer23893/Pulp/internal/cache"
	"github.com/Baseplayer23893/Pulp/internal/cleaner"
	"github.com/Baseplayer23893/Pulp/internal/defuddle"
	"github.com/Baseplayer23893/Pulp/internal/storage"
	"github.com/Baseplayer23893/Pulp/internal/urlutil"
	"github.com/spf13/cobra"
)

var extractCmd = &cobra.Command{
	Use:   "extract <url>",
	Short: "Extract web page content as clean markdown",
	Long: `Extract clean, token-efficient markdown from any web page.
Uses defuddle under the hood for high-quality content extraction,
then applies post-processing to clean tracking params, normalize
whitespace, and add YAML frontmatter.`,
	Args: cobra.ExactArgs(1),
	RunE: runExtract,
}

func init() {
	rootCmd.AddCommand(extractCmd)
}

func runExtract(cmd *cobra.Command, args []string) error {
	url, err := urlutil.NormalizeURL(args[0])
	if err != nil {
		return fmt.Errorf("invalid URL: %s", err)
	}
	targetOutput := resolveOutputPath(outputFlag, url, ".md")

	if !quietFlag {
		fmt.Fprintf(os.Stderr, "⚡ Extracting: %s\n", url)
	}

	start := time.Now()

	// Check cache (unless --no-cache is set)
	var markdown string
	var result *defuddle.Result
	if !noCache {
		if cached, err := cache.Get(url); err == nil {
			if !quietFlag {
				fmt.Fprintf(os.Stderr, "📋 Using cached result\n")
			}
			markdown = cached
		}
	}

	// If no cached content, extract fresh
	if markdown == "" {
		// Check defuddle is available
		if !defuddle.IsInstalled() {
			return fmt.Errorf("defuddle is not installed\nInstall with: npm install -g defuddle")
		}

		// Extract with defuddle
		result, err = defuddle.ParseURL(url)
		if err != nil {
			return fmt.Errorf("extraction failed: %w", err)
		}

		// Get the markdown content
		markdown = result.Markdown
		if markdown == "" {
			markdown = result.Content
		}
		if markdown == "" {
			return fmt.Errorf("no content extracted from %s", url)
		}

		// Clean the markdown
		markdown = cleaner.Clean(markdown)

		// Cache the cleaned markdown (unless --no-cache)
		if !noCache {
			cache.Set(url, markdown, cache.DefaultTTL)
		}
	}

	// Build output based on format
	var output string
	switch formatFlag {
	case "md":
		meta := map[string]string{
			"source":  url,
			"created": time.Now().Format("2006-01-02"),
		}
		if result != nil {
			if result.Title != "" {
				meta["title"] = result.Title
			}
			if result.Description != "" {
				desc := result.Description
				if len(desc) > 120 {
					desc = desc[:120] + "..."
				}
				meta["description"] = desc
			}
			if result.Author != "" {
				meta["author"] = result.Author
			}
			if result.Domain != "" {
				meta["domain"] = result.Domain
			}
		}

		output = cleaner.AddFrontmatter(markdown, meta)

	case "skillzip":
		name := ""
		description := ""
		if result != nil {
			name = result.Title
			description = result.Description
		}
		if name == "" {
			name = sanitizeURLToName(url)
		}
		frontmatter := storage.GenerateFrontmatter(name, description, url, nil)
		content := frontmatter + "# " + name + "\n\n" + markdown

		zipDir := "."
		if targetOutput != "" {
			zipDir = filepath.Dir(targetOutput)
		}
		zipPath, err := storage.CreateSkillZip(name, content, nil, zipDir)
		if err != nil {
			return fmt.Errorf("failed to create skill.zip: %w", err)
		}

		if !quietFlag {
			elapsed := time.Since(start)
			fmt.Fprintf(os.Stderr, "✅ Created: %s (%s)\n", zipPath, elapsed.Round(time.Millisecond))
		}
		return nil

	default:
		output = markdown
	}

	// Write output
	if err := storage.WriteOutput(output, targetOutput); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	if !quietFlag {
		elapsed := time.Since(start)
		wordCount := len(strings.Fields(markdown))
		target := "stdout"
		if targetOutput != "" {
			target = targetOutput
		}
		fmt.Fprintf(os.Stderr, "✅ Done: %d words → %s (%s)\n", wordCount, target, elapsed.Round(time.Millisecond))
	}

	return nil
}

func sanitizeURLToName(rawURL string) string {
	// Extract meaningful name from URL
	name := rawURL
	name = strings.TrimPrefix(name, "https://")
	name = strings.TrimPrefix(name, "http://")
	name = strings.TrimPrefix(name, "www.")
	name = strings.Split(name, "?")[0]
	name = strings.TrimSuffix(name, "/")
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, ".", "-")
	if len(name) > 60 {
		name = name[:60]
	}
	return name
}
