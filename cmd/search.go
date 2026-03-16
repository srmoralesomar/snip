package cmd

import (
	"fmt"
	"strings"

	"github.com/omarmorales/snip/internal/store"
	"github.com/sahilm/fuzzy"
	"github.com/spf13/cobra"
)

const (
	ansiReset  = "\033[0m"
	ansiBold   = "\033[1m"
	ansiYellow = "\033[33m"
)

var searchLimit int

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
	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := args[0]

	dbPath, err := store.DefaultPath()
	if err != nil {
		return fmt.Errorf("resolve db path: %w", err)
	}

	s, err := store.New(dbPath)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer s.Close()

	// Load all clips for searching.
	clips, err := s.List(0)
	if err != nil {
		return fmt.Errorf("list clips: %w", err)
	}

	if len(clips) == 0 {
		fmt.Println("No clipboard history yet. Start the daemon and copy something!")
		return nil
	}

	// Build a flat string slice for fuzzy matching (single-line previews).
	contents := make([]string, len(clips))
	for i, c := range clips {
		contents[i] = singleLine(c.Content)
	}

	matches := fuzzy.Find(query, contents)

	if len(matches) == 0 {
		fmt.Printf("No results for %q\n", query)
		return nil
	}

	// Cap results to --limit.
	if searchLimit > 0 && len(matches) > searchLimit {
		matches = matches[:searchLimit]
	}

	fmt.Printf("%-5s  %-12s  %s\n", "INDEX", "TIME", "PREVIEW")
	fmt.Printf("%-5s  %-12s  %s\n", "-----", "------------", strings.Repeat("-", 40))

	for _, m := range matches {
		clip := clips[m.Index]
		relTime := relativeTime(clip.Timestamp)
		preview := highlightMatches(truncate(singleLine(clip.Content), 80), m.MatchedIndexes)
		fmt.Printf("%-5d  %-12s  %s\n", m.Index+1, relTime, preview)
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
func highlightMatches(s string, matchedIndexes []int) string {
	if len(matchedIndexes) == 0 {
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
