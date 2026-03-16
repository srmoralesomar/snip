package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/sahilm/fuzzy"
	"github.com/spf13/cobra"
	"github.com/srmoralesomar/snip/internal/store"
)

const (
	ansiReset  = "\033[0m"
	ansiBold   = "\033[1m"
	ansiYellow = "\033[33m"
)

var (
	searchLimit int
	searchJSON  bool
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Fuzzy-search clipboard history",
	Long: `Fuzzy-search all stored clipboard entries and display ranked results.

Matched characters are highlighted in the output.`,
	Args: cobra.ExactArgs(1),
	RunE: runSearch,
}

func init() {
	searchCmd.Flags().IntVarP(&searchLimit, "limit", "l", 10, "Maximum number of results to show")
	searchCmd.Flags().BoolVar(&searchJSON, "json", false, "Output as JSON")
	rootCmd.AddCommand(searchCmd)
}

// searchClipJSON is the JSON representation of a search result.
type searchClipJSON struct {
	Index     int    `json:"index"`
	Timestamp string `json:"timestamp"`
	Content   string `json:"content"`
	Score     int    `json:"score"`
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := args[0]

	dbPath, err := store.DefaultPath()
	if err != nil {
		return fmt.Errorf("resolve db path: %w", err)
	}

	s, err := store.New(dbPath)
	if err != nil {
		return fmt.Errorf("open history database: %w\nHint: run 'snip daemon' to start recording clipboard history", err)
	}
	defer s.Close()

	// Load all clips for searching.
	clips, err := s.List(0)
	if err != nil {
		return fmt.Errorf("list clips: %w", err)
	}

	if len(clips) == 0 {
		color.Yellow("No clipboard history yet. Start the daemon and copy something!")
		return nil
	}

	// Build a flat string slice for fuzzy matching (single-line previews).
	contents := make([]string, len(clips))
	for i, c := range clips {
		contents[i] = singleLine(c.Content)
	}

	matches := fuzzy.Find(query, contents)

	if len(matches) == 0 {
		color.Yellow("No results for %q\n", query)
		return nil
	}

	// Cap results to --limit.
	if searchLimit > 0 && len(matches) > searchLimit {
		matches = matches[:searchLimit]
	}

	if searchJSON {
		out := make([]searchClipJSON, len(matches))
		for i, m := range matches {
			clip := clips[m.Index]
			out[i] = searchClipJSON{
				Index:     m.Index + 1,
				Timestamp: clip.Timestamp.UTC().Format(time.RFC3339),
				Content:   clip.Content,
				Score:     m.Score,
			}
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	}

	header := color.New(color.Bold)
	indexColor := color.New(color.FgCyan, color.Bold)
	timeColor := color.New(color.FgYellow)

	header.Printf("%-5s  %-12s  %s\n", "INDEX", "TIME", "PREVIEW")
	fmt.Printf("%-5s  %-12s  %s\n", "-----", "------------", strings.Repeat("-", 40))

	for _, m := range matches {
		clip := clips[m.Index]
		relTime := relativeTime(clip.Timestamp)
		preview := highlightMatches(truncate(singleLine(clip.Content), 80), m.MatchedIndexes)
		fmt.Printf("%s  %s  %s\n",
			indexColor.Sprintf("%-5d", m.Index+1),
			timeColor.Sprintf("%-12s", relTime),
			preview,
		)
	}

	return nil
}

// singleLine replaces newlines/tabs with spaces for display.
func singleLine(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	return s
}

// highlightMatches wraps matched character positions with ANSI color codes.
// matchedIndexes must be in ascending order (as returned by the fuzzy library).
// When color output is disabled (--no-color or NO_COLOR env), returns s unchanged.
func highlightMatches(s string, matchedIndexes []int) string {
	if len(matchedIndexes) == 0 || color.NoColor {
		return s
	}

	runes := []rune(s)
	matched := make(map[int]bool, len(matchedIndexes))
	for _, idx := range matchedIndexes {
		if idx < len(runes) {
			matched[idx] = true
		}
	}

	var b strings.Builder
	inHighlight := false
	for i, r := range runes {
		if matched[i] && !inHighlight {
			b.WriteString(ansiBold + ansiYellow)
			inHighlight = true
		} else if !matched[i] && inHighlight {
			b.WriteString(ansiReset)
			inHighlight = false
		}
		b.WriteRune(r)
	}
	if inHighlight {
		b.WriteString(ansiReset)
	}
	return b.String()
}
