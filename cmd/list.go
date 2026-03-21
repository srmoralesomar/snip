package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/srmoralesomar/snip/internal/store"
)

var (
	listCount int
	listJSON  bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Show recent clipboard history",
	Long: `Show the most recent clipboard entries.

Displays index, relative timestamp, and a content preview for each entry.
Long content is truncated to 80 characters.`,
	RunE: runList,
}

func init() {
	listCmd.Flags().IntVarP(&listCount, "count", "n", 20, "Number of clips to show")
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output as JSON")
	rootCmd.AddCommand(listCmd)
}

// listClipJSON is the JSON representation of a clip for list/search output.
type listClipJSON struct {
	Index     int    `json:"index"`
	Timestamp string `json:"timestamp"`
	Content   string `json:"content"`
}

func runList(cmd *cobra.Command, args []string) error {
	dbPath, err := store.DefaultPath()
	if err != nil {
		return fmt.Errorf("resolve db path: %w", err)
	}

	s, err := store.New(dbPath)
	if err != nil {
		return fmt.Errorf("open history database: %w\nHint: run 'snip start' to start recording clipboard history", err)
	}
	defer s.Close()

	clips, err := s.List(listCount)
	if err != nil {
		return fmt.Errorf("list clips: %w", err)
	}

	if len(clips) == 0 {
		color.Yellow("No clipboard history yet. Run snip start and copy something!")
		return nil
	}

	if listJSON {
		out := make([]listClipJSON, len(clips))
		for i, clip := range clips {
			out[i] = listClipJSON{
				Index:     i + 1,
				Timestamp: clip.Timestamp.UTC().Format(time.RFC3339),
				Content:   clip.Content,
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

	for i, clip := range clips {
		preview := truncate(clip.Content, 80)
		relTime := relativeTime(clip.Timestamp)
		fmt.Printf("%s  %s  %s\n",
			indexColor.Sprintf("%-5d", i+1),
			timeColor.Sprintf("%-12s", relTime),
			preview,
		)
	}

	return nil
}

// truncate shortens s to maxChars runes, appending "…" if truncated.
func truncate(s string, maxChars int) string {
	// Replace newlines and tabs with spaces for single-line preview
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\t", " ")

	if utf8.RuneCountInString(s) <= maxChars {
		return s
	}
	runes := []rune(s)
	return string(runes[:maxChars-1]) + "…"
}

// relativeTime returns a human-readable relative time string.
func relativeTime(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		secs := int(d.Seconds())
		if secs <= 1 {
			return "just now"
		}
		return fmt.Sprintf("%ds ago", secs)
	case d < time.Hour:
		mins := int(d.Minutes())
		if mins == 1 {
			return "1m ago"
		}
		return fmt.Sprintf("%dm ago", mins)
	case d < 24*time.Hour:
		hrs := int(d.Hours())
		if hrs == 1 {
			return "1h ago"
		}
		return fmt.Sprintf("%dh ago", hrs)
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1d ago"
		}
		return fmt.Sprintf("%dd ago", days)
	}
}
