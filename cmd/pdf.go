package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Baseplayer23893/Pulp/internal/cache"
	"github.com/Baseplayer23893/Pulp/internal/cleaner"
	"github.com/Baseplayer23893/Pulp/internal/storage"
	"github.com/ledongthuc/pdf"
	"github.com/spf13/cobra"
)

var pdfCmd = &cobra.Command{
	Use:   "pdf <file>",
	Short: "Extract text from a PDF file",
	Args:  cobra.ExactArgs(1),
	RunE:  runPDF,
}

func init() {
	rootCmd.AddCommand(pdfCmd)
}

func runPDF(cmd *cobra.Command, args []string) error {
	filePath := args[0]
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		absPath = filePath
	}
	targetOutput := resolveOutputPath(outputFlag, absPath, ".md")
	if !quietFlag {
		fmt.Fprintf(os.Stderr, "📄 Extracting PDF: %s\n", filePath)
	}

	start := time.Now()

	// Check cache (unless --no-cache is set)
	var markdown string
	pageCount := 0

	if !noCache {
		if cached, err := cache.Get(absPath); err == nil {
			if !quietFlag {
				fmt.Fprintf(os.Stderr, "📋 Using cached result\n")
			}
			markdown = cached
		}
	}

	// If no cached content, extract fresh
	if markdown == "" {
		f, r, err := pdf.Open(filePath)
		if err != nil {
			return fmt.Errorf("failed to open PDF: %w", err)
		}
		defer f.Close()

		var sb strings.Builder
		pageCount = r.NumPage()

		for i := 1; i <= pageCount; i++ {
			page := r.Page(i)
			if page.V.IsNull() {
				continue
			}
			text, err := page.GetPlainText(nil)
			if err != nil {
				continue
			}
			text = strings.TrimSpace(text)
			if text != "" {
				if i > 1 {
					sb.WriteString("\n\n---\n\n")
				}
				sb.WriteString(text)
			}
		}

		markdown = cleaner.Clean(sb.String())
		if markdown == "" || strings.TrimSpace(markdown) == "" {
			return fmt.Errorf("no text content extracted from PDF")
		}

		// Cache the cleaned markdown (unless --no-cache)
		if !noCache {
			cache.Set(absPath, markdown, cache.DefaultTTL)
		}
	}

	meta := map[string]string{
		"source":  filePath,
		"created": time.Now().Format("2006-01-02"),
		"type":    "pdf",
	}

	output := cleaner.AddFrontmatter(markdown, meta)

	if err := storage.WriteOutput(output, targetOutput); err != nil {
		return err
	}

	if !quietFlag {
		elapsed := time.Since(start)
		target := "stdout"
		if targetOutput != "" {
			target = targetOutput
		}
		wordCount := len(strings.Fields(markdown))
		if pageCount > 0 {
			fmt.Fprintf(os.Stderr, "✅ Done: %d pages, %d words → %s (%s)\n", pageCount, wordCount, target, elapsed.Round(time.Millisecond))
		} else {
			fmt.Fprintf(os.Stderr, "✅ Done: %d words → %s (%s)\n", wordCount, target, elapsed.Round(time.Millisecond))
		}
	}
	return nil
}
