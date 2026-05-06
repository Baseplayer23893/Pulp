package storage

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Baseplayer23893/Pulp/internal/yamlutil"
)

// WriteFile writes content to a file, creating directories as needed
func WriteFile(path string, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	return os.WriteFile(path, []byte(content), 0644)
}

// WriteOutput handles writing content to file or stdout
func WriteOutput(content string, outputPath string) error {
	if outputPath == "" {
		_, err := fmt.Fprint(os.Stdout, content)
		return err
	}

	return WriteFile(outputPath, content)
}

// CreateSkillZip creates a skill.zip package with SKILL.md and references
func CreateSkillZip(name string, content string, references []string, outputDir string) (string, error) {
	if outputDir == "" {
		outputDir = "."
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory %s: %w", outputDir, err)
	}

	// Sanitize name for filesystem
	safeName := sanitizeName(name)
	zipPath := filepath.Join(outputDir, safeName+".zip")

	file, err := os.Create(zipPath)
	if err != nil {
		return "", fmt.Errorf("failed to create zip: %w", err)
	}

	w := zip.NewWriter(file)

	// Add SKILL.md
	skillFile, err := w.Create(safeName + "/SKILL.md")
	if err != nil {
		_ = file.Close()
		return "", fmt.Errorf("failed to add SKILL.md: %w", err)
	}
	if _, err := skillFile.Write([]byte(content)); err != nil {
		_ = w.Close()
		_ = file.Close()
		return "", err
	}

	// Add references
	for _, ref := range references {
		data, err := os.ReadFile(ref)
		if err != nil {
			continue // Skip unreadable references
		}

		refName := filepath.Base(ref)
		refFile, err := w.Create(safeName + "/references/" + refName)
		if err != nil {
			continue
		}
		if _, err := refFile.Write(data); err != nil {
			_ = w.Close()
			return "", fmt.Errorf("failed to write reference %s: %w", ref, err)
		}
	}

	if err := w.Close(); err != nil {
		return "", fmt.Errorf("failed to finalize zip: %w", err)
	}
	if err := file.Close(); err != nil {
		return "", fmt.Errorf("failed to close zip: %w", err)
	}

	return zipPath, nil
}

// GenerateFrontmatter creates YAML frontmatter for a SKILL.md
func GenerateFrontmatter(name, description, source string, tags []string) string {
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("name: %s\n", yamlutil.QuoteString(name)))
	if description != "" {
		sb.WriteString(fmt.Sprintf("description: %s\n", yamlutil.QuoteString(description)))
	}
	sb.WriteString(fmt.Sprintf("created: %s\n", time.Now().Format("2006-01-02")))
	if source != "" {
		sb.WriteString(fmt.Sprintf("source: %s\n", yamlutil.QuoteString(source)))
	}
	if len(tags) > 0 {
		quotedTags := make([]string, 0, len(tags))
		for _, tag := range tags {
			quotedTags = append(quotedTags, yamlutil.QuoteString(tag))
		}
		sb.WriteString(fmt.Sprintf("tags: [%s]\n", strings.Join(quotedTags, ", ")))
	}
	sb.WriteString("---\n\n")
	return sb.String()
}

func sanitizeName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	// Remove characters that are problematic in filenames
	replacer := strings.NewReplacer(
		"/", "-", "\\", "-", ":", "-", "*", "", "?", "",
		"\"", "", "<", "", ">", "", "|", "",
	)
	return replacer.Replace(name)
}
