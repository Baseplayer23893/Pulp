package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Baseplayer23893/Pulp/internal/cleaner"
	"github.com/Baseplayer23893/Pulp/internal/storage"
	"github.com/Baseplayer23893/Pulp/internal/urlutil"
	"github.com/spf13/cobra"
)

var (
	hnTopN    int
	hnMaxDepth int
)

var hnCmd = &cobra.Command{
	Use:     "hn <url|id>",
	Aliases: []string{"hackernews"},
	Short:   "Extract Hacker News post with top comments",
	Long: `Extract a Hacker News post and its top comments as clean markdown.
Uses the official HN Firebase API (no authentication required).

Supports:
- HN item URLs (news.ycombinator.com/item?id=123)
- HN item IDs directly
- Shows title, URL, score, author, and top comments`,
	Args: cobra.ExactArgs(1),
	RunE: runHN,
}

func init() {
	hnCmd.Flags().IntVarP(&hnTopN, "comments", "n", 10, "Number of top comments to include")
	hnCmd.Flags().IntVarP(&hnMaxDepth, "depth", "d", 2, "Maximum comment reply depth")
	rootCmd.AddCommand(hnCmd)
}

const hnAPIBase = "https://hacker-news.firebaseio.com/v0"

type HNItem struct {
	ID        int    `json:"id"`
	Type      string `json:"type"`
	By        string `json:"by"`
	Time      int    `json:"time"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	Text      string `json:"text"`
	Score     int    `json:"score"`
	Descendants int  `json:"descendants"`
	Kids      []int  `json:"kids"`
}

func runHN(cmd *cobra.Command, args []string) error {
	url, err := urlutil.NormalizeURL(args[0])
	if err != nil {
		return fmt.Errorf("invalid URL: %s", err)
	}
	targetOutput := resolveOutputPath(outputFlag, url, ".md")
	if !quietFlag {
		fmt.Fprintf(os.Stderr, "📝 Extracting Hacker News: %s\n", url)
	}

	start := time.Now()

	// Extract HN item ID from URL or use as-is
	itemID, err := extractHNItemID(url)
	if err != nil {
		return fmt.Errorf("failed to extract HN item ID: %w", err)
	}

	// Fetch the item
	item, err := fetchHNItem(itemID)
	if err != nil {
		return fmt.Errorf("failed to fetch HN item: %w", err)
	}

	if item.Type != "story" {
		return fmt.Errorf("HN item is not a story (type: %s)", item.Type)
	}

	// Fetch top comments
	comments, err := fetchHNComments(item.Kids, hnTopN, 0)
	if err != nil {
		return fmt.Errorf("failed to fetch comments: %w", err)
	}

	// Build markdown output
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n", item.Title))

	sb.WriteString(fmt.Sprintf("**Author:** %s\n", item.By))
	sb.WriteString(fmt.Sprintf("**Score:** %d points\n", item.Score))
	if item.Descendants > 0 {
		sb.WriteString(fmt.Sprintf("**Comments:** %d\n", item.Descendants))
	}
	if item.Time > 0 {
		t := time.Unix(int64(item.Time), 0)
		sb.WriteString(fmt.Sprintf("**Posted:** %s\n", t.Format("2006-01-02 15:04")))
	}
	if item.URL != "" {
		sb.WriteString(fmt.Sprintf("**Link:** %s\n", item.URL))
	}
	sb.WriteString(fmt.Sprintf("**HN:** https://news.ycombinator.com/item?id=%d\n", item.ID))

	sb.WriteString("\n---\n\n")

	// Add link post content if present
	if item.Text != "" {
		sb.WriteString("## Post\n\n")
		sb.WriteString(cleanHNText(item.Text))
		sb.WriteString("\n\n---\n\n")
	}

	// Add comments
	if len(comments) > 0 {
		sb.WriteString("## Top Comments\n\n")
		for _, comment := range comments {
			renderHNComment(&sb, comment, 0)
		}
	}

	markdown := sb.String()

	// Add frontmatter
	meta := map[string]string{
		"source":    fmt.Sprintf("https://news.ycombinator.com/item?id=%d", item.ID),
		"created":   time.Now().Format("2006-01-02"),
		"title":     item.Title,
		"author":    item.By,
		"type":      "hn-post",
	}
	output := cleaner.AddFrontmatter(markdown, meta)

	// Dry-run
	if dryRun {
		wordCount := len(strings.Fields(markdown))
		outPath := "stdout"
		if targetOutput != "" {
			outPath = targetOutput
		}
		fmt.Printf("title: %s\n", item.Title)
		fmt.Printf("wordCount: %d\n", wordCount)
		fmt.Printf("sourceType: hn\n")
		fmt.Printf("outputPath: %s\n", outPath)
		return nil
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

func extractHNItemID(url string) (int, error) {
	// Direct ID
	if strings.TrimSpace(url) != "" {
		var id int
		if _, err := fmt.Sscanf(url, "%d", &id); err == nil {
			return id, nil
		}
	}

	// news.ycombinator.com/item?id=123
	idRe := regexp.MustCompile(`id=(\d+)`)
	matches := idRe.FindStringSubmatch(url)
	if len(matches) > 1 {
		var id int
		if _, err := fmt.Sscanf(matches[1], "%d", &id); err == nil {
			return id, nil
		}
	}

	return 0, fmt.Errorf("could not extract HN item ID from %s", url)
}

func fetchHNItem(id int) (*HNItem, error) {
	url := fmt.Sprintf("%s/item/%d.json", hnAPIBase, id)

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Pulp/0.4 (HN extractor)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HN API returned status %d", resp.StatusCode)
	}

	var item HNItem
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, fmt.Errorf("failed to parse HN response: %w", err)
	}

	if item.ID == 0 {
		return nil, fmt.Errorf("HN item not found")
	}

	return &item, nil
}

func fetchHNComments(kids []int, limit int, depth int) ([]HNItem, error) {
	if depth > hnMaxDepth || len(kids) == 0 {
		return nil, nil
	}

	// Fetch top-level comments in parallel
	type result struct {
		item  *HNItem
		index int
	}

	results := make([]result, len(kids))
	var wg sync.WaitGroup
	var mu sync.Mutex
	completed := 0

	for i, kidID := range kids {
		if i >= limit {
			break
		}
		wg.Add(1)
		go func(idx int, id int) {
			defer wg.Done()
			item, err := fetchHNItem(id)
			mu.Lock()
			if err == nil && item != nil {
				results[idx] = result{item: item, index: idx}
				completed++
			}
			mu.Unlock()
		}(i, kidID)
	}
	wg.Wait()

	// Collect valid comments
	var comments []HNItem
	for _, r := range results {
		if r.item != nil && r.item.Type == "comment" && r.item.Text != "" {
			comments = append(comments, *r.item)

			// Fetch replies recursively (still sequential to maintain order)
			if len(r.item.Kids) > 0 && depth < hnMaxDepth && len(comments) < limit {
				replies, _ := fetchHNComments(r.item.Kids, limit-len(comments), depth+1)
				comments = append(comments, replies...)
			}
		}
	}

	return comments, nil
}

func renderHNComment(sb *strings.Builder, comment HNItem, depth int) {
	indent := strings.Repeat("> ", depth)

	// Author line
	scoreStr := ""
	if comment.Score > 0 {
		scoreStr = fmt.Sprintf(" (%d pts)", comment.Score)
	}
	t := time.Unix(int64(comment.Time), 0)

	if depth == 0 {
		sb.WriteString(fmt.Sprintf("### %s%s — %s\n\n", comment.By, scoreStr, t.Format("2006-01-02")))
	} else {
		sb.WriteString(fmt.Sprintf("%s**%s**%s — %s\n", indent, comment.By, scoreStr, t.Format("2006-01-02")))
	}

	// Comment body - strip HTML
	text := cleanHNText(comment.Text)
	sb.WriteString(indent + strings.ReplaceAll(text, "\n", "\n"+indent))
	sb.WriteString("\n\n")

	// Render nested replies
	if len(comment.Kids) > 0 && depth < hnMaxDepth {
		replies, _ := fetchHNComments(comment.Kids, 3, depth+1)
		for _, reply := range replies {
			renderHNComment(sb, reply, depth+1)
		}
	}
}

func cleanHNText(text string) string {
	// Strip HTML tags that HN uses
	text = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(text, "")
	// Decode common HTML entities
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	// Clean excessive whitespace
	text = regexp.MustCompile(`\n{3,}`).ReplaceAllString(text, "\n\n")
	return strings.TrimSpace(text)
}