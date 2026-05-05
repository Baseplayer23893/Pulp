package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/Baseplayer23893/Pulp/internal/cleaner"
	"github.com/Baseplayer23893/Pulp/internal/storage"
	"github.com/spf13/cobra"
)

var (
	redditTopN     int
	redditDepth    int
	redditIncFlair bool
)

var redditCmd = &cobra.Command{
	Use:   "reddit <url>",
	Short: "Extract Reddit post with top comments",
	Long: `Extract a Reddit post and its top comments as clean markdown.
Uses Reddit's public JSON API (no authentication required).

Supports:
- Standard post URLs (www.reddit.com, old.reddit.com, new.reddit.com)
- Share links (reddit.com/r/.../s/...)
- Cross-posts
- Link posts and self posts
- Nested comment threads with configurable depth`,
	Args: cobra.ExactArgs(1),
	RunE: runReddit,
}

func init() {
	redditCmd.Flags().IntVarP(&redditTopN, "comments", "n", 10, "Number of top comments to include")
	redditCmd.Flags().IntVarP(&redditDepth, "depth", "d", 2, "Maximum comment reply depth (0=top-level only)")
	redditCmd.Flags().BoolVar(&redditIncFlair, "flair", false, "Include post and user flairs")
	rootCmd.AddCommand(redditCmd)
}

// RedditListing is the top-level Reddit API response
type RedditListing struct {
	Data struct {
		Children []RedditChild `json:"children"`
	} `json:"data"`
}

// RedditChild wraps the kind + data pattern
type RedditChild struct {
	Kind string     `json:"kind"`
	Data RedditPost `json:"data"`
}

// RedditPost represents a post or comment
type RedditPost struct {
	Title         string        `json:"title"`
	Author        string        `json:"author"`
	AuthorFlair   string        `json:"author_flair_text"`
	Selftext      string        `json:"selftext"`
	Score         int           `json:"score"`
	URL           string        `json:"url"`
	Permalink     string        `json:"permalink"`
	Subreddit     string        `json:"subreddit"`
	Body          string        `json:"body"`
	NumComments   int           `json:"num_comments"`
	Created       float64       `json:"created_utc"`
	LinkFlairText string        `json:"link_flair_text"`
	IsSelf        bool          `json:"is_self"`
	Domain        string        `json:"domain"`
	Ups           int           `json:"ups"`
	Downs         int           `json:"downs"`
	Gilded        int           `json:"gilded"`
	Stickied      bool          `json:"stickied"`
	Distinguished string        `json:"distinguished"`
	Replies       RedditReplies `json:"replies"`
}

// RedditReplies handles the polymorphic "replies" field (string or object)
type RedditReplies struct {
	Data struct {
		Children []RedditChild `json:"children"`
	} `json:"data"`
}

func (r *RedditReplies) UnmarshalJSON(data []byte) error {
	// Reddit returns "" when there are no replies, or an object when there are
	if len(data) == 0 || string(data) == `""` || string(data) == `null` {
		return nil
	}
	// Try to unmarshal as a listing object
	type Alias RedditReplies
	var alias Alias
	if err := json.Unmarshal(data, &alias); err != nil {
		return nil // Silently ignore — it's just an empty string
	}
	r.Data = alias.Data
	return nil
}

func runReddit(cmd *cobra.Command, args []string) error {
	url := args[0]
	targetOutput := resolveOutputPath(outputFlag, url, ".md")
	if !quietFlag {
		fmt.Fprintf(os.Stderr, "🔗 Extracting Reddit post: %s\n", url)
	}

	start := time.Now()

	// Normalize URL and convert to JSON endpoint
	jsonURL := normalizeRedditURL(url)
	if !quietFlag {
		fmt.Fprintf(os.Stderr, "   Fetching: %s\n", jsonURL)
	}

	// Fetch the JSON data
	client := &http.Client{
		Timeout: 15 * time.Second,
		// Follow redirects (for /s/ share links)
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}
	req, err := http.NewRequest("GET", jsonURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "Pulp/0.2 (CLI content extractor)")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch Reddit post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Reddit API returned status %d — check if the URL is valid", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Reddit returns an array of two listings: [post, comments]
	var listings []RedditListing
	if err := json.Unmarshal(body, &listings); err != nil {
		return fmt.Errorf("failed to parse Reddit response: %w", err)
	}

	if len(listings) < 1 || len(listings[0].Data.Children) < 1 {
		return fmt.Errorf("no post found at %s", url)
	}

	post := listings[0].Data.Children[0].Data

	// Build markdown output
	var sb strings.Builder

	// Title with optional flair
	titleLine := post.Title
	if redditIncFlair && post.LinkFlairText != "" {
		titleLine = fmt.Sprintf("[%s] %s", post.LinkFlairText, post.Title)
	}
	sb.WriteString(fmt.Sprintf("# %s\n\n", titleLine))

	// Metadata block
	sb.WriteString(fmt.Sprintf("**Subreddit:** r/%s\n", post.Subreddit))
	authorLine := fmt.Sprintf("u/%s", post.Author)
	if redditIncFlair && post.AuthorFlair != "" {
		authorLine += fmt.Sprintf(" (%s)", post.AuthorFlair)
	}
	if post.Distinguished == "moderator" {
		authorLine += " [MOD]"
	} else if post.Distinguished == "admin" {
		authorLine += " [ADMIN]"
	}
	sb.WriteString(fmt.Sprintf("**Author:** %s\n", authorLine))
	sb.WriteString(fmt.Sprintf("**Score:** %d points\n", post.Score))
	sb.WriteString(fmt.Sprintf("**Comments:** %d\n", post.NumComments))
	if post.Gilded > 0 {
		sb.WriteString(fmt.Sprintf("**Awards:** %d\n", post.Gilded))
	}
	if post.Created > 0 {
		t := time.Unix(int64(post.Created), 0)
		sb.WriteString(fmt.Sprintf("**Posted:** %s\n", t.Format("2006-01-02 15:04 UTC")))
	}
	sb.WriteString(fmt.Sprintf("**Source:** https://reddit.com%s\n", post.Permalink))

	// Link post — show the linked URL
	if !post.IsSelf && post.URL != "" {
		sb.WriteString(fmt.Sprintf("\n**Linked URL:** [%s](%s)\n", post.Domain, post.URL))
	}
	sb.WriteString("\n---\n\n")

	// Post body
	if post.Selftext != "" {
		sb.WriteString("## Post\n\n")
		sb.WriteString(cleanRedditMarkdown(post.Selftext))
		sb.WriteString("\n\n")
	}

	// Comments
	if len(listings) > 1 && len(listings[1].Data.Children) > 0 {
		sb.WriteString("---\n\n## Top Comments\n\n")
		count := 0
		for _, child := range listings[1].Data.Children {
			if count >= redditTopN {
				break
			}
			if child.Kind != "t1" || child.Data.Body == "" {
				continue
			}
			if child.Data.Stickied {
				continue // Skip pinned bot comments
			}
			renderComment(&sb, child.Data, 0)
			count++
		}
	}

	markdown := sb.String()

	// Add frontmatter
	meta := map[string]string{
		"source":    url,
		"created":   time.Now().Format("2006-01-02"),
		"title":     post.Title,
		"subreddit": "r/" + post.Subreddit,
		"author":    "u/" + post.Author,
		"type":      "reddit-post",
	}
	output := cleaner.AddFrontmatter(markdown, meta)

	// Write output
	if err := storage.WriteOutput(output, targetOutput); err != nil {
		return err
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

// renderComment writes a single comment and its nested replies
func renderComment(sb *strings.Builder, comment RedditPost, depth int) {
	if depth > redditDepth {
		return
	}

	indent := strings.Repeat("> ", depth)
	prefix := indent

	// Author line
	authorInfo := fmt.Sprintf("u/%s", comment.Author)
	if comment.Distinguished == "moderator" {
		authorInfo += " [MOD]"
	}
	scoreStr := fmt.Sprintf("%d pts", comment.Score)

	if depth == 0 {
		sb.WriteString(fmt.Sprintf("### %s (%s)\n\n", authorInfo, scoreStr))
	} else {
		sb.WriteString(fmt.Sprintf("%s**%s** (%s)\n%s\n", prefix, authorInfo, scoreStr, prefix))
	}

	// Comment body
	bodyLines := strings.Split(cleanRedditMarkdown(comment.Body), "\n")
	for _, line := range bodyLines {
		sb.WriteString(prefix + line + "\n")
	}
	sb.WriteString("\n")

	// Render nested replies
	if len(comment.Replies.Data.Children) > 0 {
		replyCount := 0
		for _, reply := range comment.Replies.Data.Children {
			if reply.Kind != "t1" || reply.Data.Body == "" {
				continue
			}
			if replyCount >= 3 {
				// Limit nested replies to avoid huge output
				remaining := len(comment.Replies.Data.Children) - replyCount
				if remaining > 0 {
					sb.WriteString(fmt.Sprintf("%s*... %d more replies*\n\n", indent, remaining))
				}
				break
			}
			renderComment(sb, reply.Data, depth+1)
			replyCount++
		}
	}
}

// normalizeRedditURL handles all Reddit URL formats
func normalizeRedditURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)

	// Remove trailing slash
	rawURL = strings.TrimSuffix(rawURL, "/")

	// Handle various Reddit domains
	rawURL = strings.Replace(rawURL, "https://new.reddit.com", "https://old.reddit.com", 1)
	rawURL = strings.Replace(rawURL, "https://www.reddit.com", "https://old.reddit.com", 1)
	rawURL = strings.Replace(rawURL, "http://www.reddit.com", "https://old.reddit.com", 1)
	rawURL = strings.Replace(rawURL, "http://reddit.com", "https://old.reddit.com", 1)
	rawURL = strings.Replace(rawURL, "https://reddit.com", "https://old.reddit.com", 1)

	// Handle short share links: reddit.com/r/.../s/...
	// These redirect, so we just need to ensure they have .json
	// The HTTP client will follow redirects

	// Remove query params except those needed
	if idx := strings.Index(rawURL, "?"); idx != -1 {
		rawURL = rawURL[:idx]
	}

	// Add .json if not present
	if !strings.HasSuffix(rawURL, ".json") {
		rawURL += ".json"
	}

	return rawURL
}

// cleanRedditMarkdown cleans up Reddit's markdown quirks
func cleanRedditMarkdown(text string) string {
	// Reddit uses &amp; &lt; &gt; entities
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&nbsp;", " ")

	// Clean up Reddit's weird link format: [text](//example.com)
	schemelessLink := regexp.MustCompile(`\]\(//`)
	text = schemelessLink.ReplaceAllString(text, "](https://")

	// Clean excessive newlines
	multiNewlines := regexp.MustCompile(`\n{3,}`)
	text = multiNewlines.ReplaceAllString(text, "\n\n")

	return strings.TrimSpace(text)
}
