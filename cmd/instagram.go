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
	"github.com/spf13/cobra"
)

import (
	"github.com/Baseplayer23893/Pulp/internal/urlutil"
)

var (
	instagramMentionRe = regexp.MustCompile(`@(\w+)`)
	instagramTagRe     = regexp.MustCompile(`#(\w+)`)
)

var instagramCmd = &cobra.Command{
	Use:     "instagram <url>",
	Aliases: []string{"ig", "insta"},
	Short:   "Extract Instagram Reel content",
	Long: `Extract content from an Instagram Reel including:
- Caption/description text with hashtag extraction
- Audio transcription from voice/speech (via yt-dlp subtitles)
- Metadata (author, likes, comments, duration)

Supports standard Reels, posts, and IGTV.
Requires yt-dlp (pip install yt-dlp).`,
	Args: cobra.ExactArgs(1),
	RunE: runInstagram,
}

func init() {
	rootCmd.AddCommand(instagramCmd)
}

// InstagramInfo holds reel metadata from yt-dlp
type InstagramInfo struct {
	Title             string                     `json:"title"`
	Description       string                     `json:"description"`
	Uploader          string                     `json:"uploader"`
	UploaderID        string                     `json:"uploader_id"`
	Channel           string                     `json:"channel"`
	ChannelID         string                     `json:"channel_id"`
	Timestamp         int64                      `json:"timestamp"`
	LikeCount         int                        `json:"like_count"`
	CommentCount      int                        `json:"comment_count"`
	Duration          int                        `json:"duration"`
	ViewCount         int                        `json:"view_count"`
	Subtitles         map[string][]SubtitleEntry `json:"subtitles"`
	AutomaticCaptions map[string][]SubtitleEntry `json:"automatic_captions"`
}

// SubtitleEntry represents a subtitle track from yt-dlp
type SubtitleEntry struct {
	URL  string `json:"url"`
	Ext  string `json:"ext"`
	Name string `json:"name"`
}

func runInstagram(cmd *cobra.Command, args []string) error {
	url, err := urlutil.NormalizeURL(args[0])
	if err != nil {
		return fmt.Errorf("invalid URL: %s", err)
	}
	targetOutput := resolveOutputPath(outputFlag, url, ".md")

	// Normalize Instagram URL
	url = normalizeInstagramURL(url)

	if !quietFlag {
		fmt.Fprintf(os.Stderr, "📸 Extracting Instagram Reel: %s\n", url)
	}

	start := time.Now()

	// Check cache (unless --no-cache is set)
	var transcript string
	var description string
	var info *InstagramInfo

	if !noCache {
		if cached, err := cache.Get(url); err == nil {
			if !quietFlag {
				fmt.Fprintf(os.Stderr, "📋 Using cached result\n")
			}
			// For cached content, we just have the raw transcript + description
			// Parse from cached format: "transcript|||description"
			parts := strings.SplitN(cached, "|||", 2)
			if len(parts) > 0 {
				transcript = parts[0]
			}
			if len(parts) > 1 {
				description = parts[1]
			}
		}
	}

	// If no cached content, extract fresh
	if description == "" && transcript == "" {
		// Check yt-dlp is available
		ytdlp, err := exec.LookPath("yt-dlp")
		if err != nil {
			return fmt.Errorf("yt-dlp not found\nInstall with: pipx install yt-dlp")
		}

		// Get reel info as JSON
		infoCmd := exec.Command(ytdlp, "--dump-json", "--no-download", url)
		infoOut, err := infoCmd.Output()
		if err != nil {
			return fmt.Errorf("failed to get reel info: %w\nMake sure the URL is a valid Instagram Reel/Post", err)
		}

		info = &InstagramInfo{}
		if err := json.Unmarshal(infoOut, info); err != nil {
			return fmt.Errorf("failed to parse reel info: %w", err)
		}

		// Attempt to extract audio transcription
		transcript = extractInstagramTranscript(ytdlp, url)
		description = info.Description

		// Cache the core content (transcript + description)
		if !noCache && !dryRun {
			cacheContent := transcript + "|||" + description
			cache.Set(url, cacheContent, cache.DefaultTTL)
		}
	}

	// Build markdown output
	var sb strings.Builder

	title := "Instagram Reel"
	uploaderHandle := "unknown"

	if info != nil {
		title = info.Title
		uploaderHandle = coalesce(info.UploaderID, info.ChannelID, info.Channel, info.Uploader)
		if title == "" {
			title = fmt.Sprintf("Instagram Reel by @%s", uploaderHandle)
		}

		sb.WriteString(fmt.Sprintf("# %s\n\n", title))
		sb.WriteString(fmt.Sprintf("**Author:** @%s\n", uploaderHandle))
		if info.Duration > 0 {
			sb.WriteString(fmt.Sprintf("**Duration:** %s\n", formatDuration(info.Duration)))
		}
		if info.LikeCount > 0 {
			sb.WriteString(fmt.Sprintf("**Likes:** %s\n", formatCount(info.LikeCount)))
		}
		if info.CommentCount > 0 {
			sb.WriteString(fmt.Sprintf("**Comments:** %s\n", formatCount(info.CommentCount)))
		}
		if info.ViewCount > 0 {
			sb.WriteString(fmt.Sprintf("**Views:** %s\n", formatCount(info.ViewCount)))
		}
		if info.Timestamp > 0 {
			t := time.Unix(info.Timestamp, 0)
			sb.WriteString(fmt.Sprintf("**Posted:** %s\n", t.Format("2006-01-02")))
		}
	} else {
		// Cached content - limited metadata
		sb.WriteString(fmt.Sprintf("# %s\n\n", title))
		sb.WriteString(fmt.Sprintf("**Author:** @%s\n", uploaderHandle))
	}

	sb.WriteString(fmt.Sprintf("**Source:** %s\n\n", url))
	sb.WriteString("---\n\n")

	// Audio transcription section
	if transcript != "" {
		sb.WriteString("## Audio Transcription\n\n")
		sb.WriteString(transcript)
		sb.WriteString("\n\n---\n\n")
	}

	// Caption section
	if description != "" {
		sb.WriteString("## Caption\n\n")
		caption := formatInstagramCaption(description)
		sb.WriteString(caption)
		sb.WriteString("\n\n")

		// Extract and list hashtags
		hashtags := extractHashtags(description)
		if len(hashtags) > 0 {
			sb.WriteString("## Hashtags\n\n")
			sb.WriteString(strings.Join(hashtags, " "))
			sb.WriteString("\n\n")
		}
	}

	markdown := sb.String()

	// Dry-run: just print info and exit
	if dryRun {
		wordCount := len(strings.Fields(markdown))
		outPath := "stdout"
		if targetOutput != "" {
			outPath = targetOutput
		}
		fmt.Printf("title: %s\n", title)
		fmt.Printf("wordCount: %d\n", wordCount)
		fmt.Printf("sourceType: instagram\n")
		fmt.Printf("outputPath: %s\n", outPath)
		return nil
	}

	// Add frontmatter
	meta := map[string]string{
		"source":  url,
		"created": time.Now().Format("2006-01-02"),
		"title":   title,
		"author":  "@" + uploaderHandle,
		"type":    "instagram-reel",
	}
	if transcript != "" {
		meta["has_transcription"] = "true"
	}
	output := cleaner.AddFrontmatter(markdown, meta)

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
		parts := []string{fmt.Sprintf("%d words", wordCount)}
		if transcript != "" {
			parts = append(parts, "with audio transcription")
		}
		fmt.Fprintf(os.Stderr, "✅ Done: %s → %s (%s)\n", strings.Join(parts, ", "), target, elapsed.Round(time.Millisecond))
	}

	return nil
}

// extractInstagramTranscript attempts to get audio transcription from IG Reels
func extractInstagramTranscript(ytdlp string, url string) string {
	tmpDir, err := os.MkdirTemp("", "pulp-ig-*")
	if err != nil {
		return ""
	}
	defer os.RemoveAll(tmpDir)

	// Try to download subtitles/captions (auto-generated speech-to-text)
	subCmd := exec.Command(ytdlp,
		"--write-sub", "--write-auto-sub",
		"--sub-lang", "en",
		"--sub-format", "vtt",
		"--skip-download",
		"-o", tmpDir+"/reel",
		url,
	)
	subCmd.Run()

	// Check for subtitle files
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return ""
	}

	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".vtt") {
			data, err := os.ReadFile(tmpDir + "/" + entry.Name())
			if err != nil {
				continue
			}
			transcript := parseVTT(string(data))
			if transcript != "" {
				return cleanTranscript(transcript)
			}
		}
	}

	return ""
}

// normalizeInstagramURL cleans up Instagram URLs to a canonical form
func normalizeInstagramURL(rawURL string) string {
	// Remove tracking params
	rawURL = strings.Split(rawURL, "?")[0]
	// Ensure https
	if !strings.HasPrefix(rawURL, "http") {
		rawURL = "https://www.instagram.com/" + strings.TrimPrefix(rawURL, "/")
	}
	return rawURL
}

// formatInstagramCaption cleans up Instagram caption text
func formatInstagramCaption(caption string) string {
	// Convert @mentions to bold
	caption = instagramMentionRe.ReplaceAllString(caption, "**@$1**")

	// Normalize whitespace but preserve intentional line breaks
	lines := strings.Split(caption, "\n")
	var cleaned []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleaned = append(cleaned, line)
		}
	}

	return strings.Join(cleaned, "\n\n")
}

// extractHashtags pulls all #hashtags from text
func extractHashtags(text string) []string {
	matches := instagramTagRe.FindAllString(text, -1)

	// Deduplicate
	seen := make(map[string]bool)
	var unique []string
	for _, tag := range matches {
		lower := strings.ToLower(tag)
		if !seen[lower] {
			seen[lower] = true
			unique = append(unique, tag)
		}
	}
	return unique
}

// coalesce returns the first non-empty string
func coalesce(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return "unknown"
}

// formatCount formats large numbers with K/M suffixes for display
func formatCount(n int) string {
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 10_000 {
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}
