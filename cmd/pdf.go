package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

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
	targetOutput := resolveOutputPath(outputFlag, filePath, ".md")
	if !quietFlag {
		fmt.Fprintf(os.Stderr, "📄 Extracting PDF: %s\n", filePath)
	}

	start := time.Now()

	f, r, err := pdf.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open PDF: %w", err)
	}
	defer f.Close()

	var sb strings.Builder
	totalPages := r.NumPage()

	for i := 1; i <= totalPages; i++ {
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

	markdown := cleaner.Clean(sb.String())
	if markdown == "" || strings.TrimSpace(markdown) == "" {
		return fmt.Errorf("no text content extracted from PDF")
	}

	meta := map[string]string{
		"source":  filePath,
		"created": time.Now().Format("2006-01-02"),
		"pages":   fmt.Sprintf("%d", totalPages),
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
		fmt.Fprintf(os.Stderr, "✅ Done: %d pages, %d words → %s (%s)\n", totalPages, len(strings.Fields(markdown)), target, elapsed.Round(time.Millisecond))
	}
	return nil
}
