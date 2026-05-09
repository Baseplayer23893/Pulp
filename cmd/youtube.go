package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/Baseplayer23893/Pulp/internal/cache"
	"github.com/Baseplayer23893/Pulp/internal/cleaner"
	"github.com/Baseplayer23893/Pulp/internal/storage"
	"github.com/Baseplayer23893/Pulp/internal/urlutil"
	"github.com/spf13/cobra"
)

var (
	youtubeTimestampRe = regexp.MustCompile(`^(?:\d{2}:)?\d{2}:\d{2}[.,]\d{3}\s*-->`)
	youtubeCueIDRe    = regexp.MustCompile(`^\d+$`)
	youtubeTagRe      = regexp.MustCompile(`<[^>]+>`)
	youtubeSpaceRe    = regexp.MustCompile(`\s+`)
	youtubeSentenceRe = regexp.MustCompile(`([.!?])\s+`)
)

var youtubeCmd = &cobra.Command{
	Use:     "youtube <url>",
	Aliases: []string{"yt"},
	Short:   "Extract YouTube video transcript",
	Long: `Extract the transcript from a YouTube video as clean markdown.
Supports standard videos, Shorts, and playlists.
Requires yt-dlp to be installed (pip install yt-dlp).`,
	Args: cobra.ExactArgs(1),
	RunE: runYoutube,
}

func init() {
	rootCmd.AddCommand(youtubeCmd)
}

// YouTubeInfo holds video metadata from yt-dlp
type YouTubeInfo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Channel     string `json:"channel"`
	UploadDate  string `json:"upload_date"`
	Duration    int    `json:"duration"`
	ViewCount   int    `json:"view_count"`
	Subtitles   map[string][]struct {
		URL string `json:"url"`
		Ext string `json:"ext"`
	} `json:"subtitles"`
	AutomaticCaptions map[string][]struct {
		URL string `json:"url"`
		Ext string `json:"ext"`
	} `json:"automatic_captions"`
}

func runYoutube(cmd *cobra.Command, args []string) error {
	url, err := urlutil.NormalizeURL(args[0])
	if err != nil {
		return fmt.Errorf("invalid URL: %s", err)
	}
	targetOutput := resolveOutputPath(outputFlag, url, ".md")

	if !quietFlag {
		fmt.Fprintf(os.Stderr, "🎬 Extracting YouTube transcript: %s\n", url)
	}

	start := time.Now()

	// Check cache (unless --no-cache is set)
	var transcript string
	var info *YouTubeInfo
	if !noCache {
		if cached, err := cache.Get(url); err == nil {
			if !quietFlag {
				fmt.Fprintf(os.Stderr, "📋 Using cached result\n")
			}
			transcript = cached
		}
	}

	// If no cached content, extract fresh
	if transcript == "" {
		// Check yt-dlp is available
		ytdlp, err := exec.LookPath("yt-dlp")
		if err != nil {
			return fmt.Errorf("yt-dlp not found\nInstall with: pipx install yt-dlp")
		}

		// Get video info as JSON
		infoCmd := exec.Command(ytdlp, "--dump-json", "--no-download", url)
		infoOut, err := infoCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to get video info: %w", err)
		}

		info = &YouTubeInfo{}
		if err := json.Unmarshal(infoOut, info); err != nil {
			return fmt.Errorf("failed to parse video info: %w", err)
		}

		// Extract subtitles/transcript
		transcript, err = extractTranscript(ytdlp, url)
		if err != nil {
			return fmt.Errorf("failed to extract transcript: %w", err)
		}

		if transcript == "" {
			return fmt.Errorf("no transcript available for this video")
		}

		// Clean transcript
		transcript = cleanTranscript(transcript)

		// Cache the cleaned transcript (unless --no-cache)
		if !noCache && !dryRun {
			cache.Set(url, transcript, cache.DefaultTTL)
		}
	}

	// Build markdown output
	var sb strings.Builder

	title := "YouTube Video"
	channel := "Unknown"
	uploadDate := ""
	duration := "Unknown"

	if info != nil {
		title = info.Title
		channel = info.Channel
		uploadDate = info.UploadDate
		duration = formatDuration(info.Duration)
	}

	sb.WriteString(fmt.Sprintf("# %s\n\n", title))
	sb.WriteString(fmt.Sprintf("**Channel:** %s\n", channel))
	if uploadDate != "" {
		sb.WriteString(fmt.Sprintf("**Published:** %s\n", formatDate(uploadDate)))
	}
	sb.WriteString(fmt.Sprintf("**Duration:** %s\n", duration))
	sb.WriteString(fmt.Sprintf("**Source:** %s\n\n", url))
	sb.WriteString("---\n\n")
	sb.WriteString("## Transcript\n\n")
	sb.WriteString(transcript)
	sb.WriteString("\n")

	markdown := sb.String()

	// Dry-run: just print info and exit
	if dryRun {
		wordCount := len(strings.Fields(transcript))
		outPath := "stdout"
		if targetOutput != "" {
			outPath = targetOutput
		}
		fmt.Printf("title: %s\n", title)
		fmt.Printf("wordCount: %d\n", wordCount)
		fmt.Printf("sourceType: youtube\n")
		fmt.Printf("outputPath: %s\n", outPath)
		return nil
	}

	// Add frontmatter
	meta := map[string]string{
		"source":  url,
		"created": time.Now().Format("2006-01-02"),
		"title":   title,
		"channel": channel,
		"type":    "youtube-transcript",
	}
	output := cleaner.AddFrontmatter(markdown, meta)

	// Write output
	if err := storage.WriteOutput(output, targetOutput); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	if !quietFlag {
		elapsed := time.Since(start)
		wordCount := len(strings.Fields(transcript))
		target := "stdout"
		if targetOutput != "" {
			target = targetOutput
		}
		fmt.Fprintf(os.Stderr, "✅ Done: %d words → %s (%s)\n", wordCount, target, elapsed.Round(time.Millisecond))
	}

	return nil
}

func extractTranscript(ytdlp string, url string) (string, error) {
	// Try to get subtitles (manual first, then auto-generated)
	// Write to temp dir
	tmpDir, err := os.MkdirTemp("", "pulp-yt-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)

	// Try manual subtitles first
	subCmd := exec.Command(ytdlp,
		"--write-sub", "--write-auto-sub",
		"--sub-lang", "en",
		"--sub-format", "vtt",
		"--skip-download",
		"-o", tmpDir+"/transcript",
		url,
	)
	subCmd.Run() // Ignore error — we check for files

	// Find and read the subtitle file
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".vtt") {
			data, err := os.ReadFile(tmpDir + "/" + entry.Name())
			if err != nil {
				continue
			}
			return parseVTT(string(data)), nil
		}
	}

	return "", fmt.Errorf("no subtitles found")
}

func parseVTT(vtt string) string {
	lines := strings.Split(vtt, "\n")
	var result []string
	seen := make(map[string]bool)

	// VTT format: timestamp lines followed by text lines
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines, WEBVTT header, timestamps, NOTE lines
		if line == "" || line == "WEBVTT" || strings.HasPrefix(line, "Kind:") ||
			strings.HasPrefix(line, "Language:") || strings.HasPrefix(line, "NOTE") ||
			youtubeTimestampRe.MatchString(line) {
			continue
		}

		// Skip numeric-only lines (cue identifiers)
		if youtubeCueIDRe.MatchString(line) {
			continue
		}

		// Remove HTML-like tags
		line = youtubeTagRe.ReplaceAllString(line, "")
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		// Deduplicate (auto-generated subs often repeat)
		if !seen[line] {
			seen[line] = true
			result = append(result, line)
		}
	}

	return strings.Join(result, " ")
}

func cleanTranscript(transcript string) string {
	// Normalize whitespace
	transcript = youtubeSpaceRe.ReplaceAllString(transcript, " ")
	transcript = strings.TrimSpace(transcript)

	// Break into sentences for readability
	// Add paragraph breaks every ~3-4 sentences
	sentences := splitSentences(transcript)
	var sb strings.Builder
	for i, sentence := range sentences {
		sb.WriteString(sentence)
		if i < len(sentences)-1 {
			if (i+1)%4 == 0 {
				sb.WriteString("\n\n")
			} else {
				sb.WriteString(" ")
			}
		}
	}

	return sb.String()
}

func splitSentences(text string) []string {
	// Simple sentence splitting
	re := youtubeSentenceRe
	parts := re.Split(text, -1)
	delims := re.FindAllStringSubmatch(text, -1)

	var sentences []string
	for i, part := range parts {
		if part == "" {
			continue
		}
		s := part
		if i < len(delims) {
			s += delims[i][1]
		}
		sentences = append(sentences, strings.TrimSpace(s))
	}
	return sentences
}

func formatDuration(seconds int) string {
	h := seconds / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60

	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

func formatDate(yyyymmdd string) string {
	if len(yyyymmdd) == 8 {
		return yyyymmdd[:4] + "-" + yyyymmdd[4:6] + "-" + yyyymmdd[6:]
	}
	return yyyymmdd
}
