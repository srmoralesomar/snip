package cmd

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/omarmorales/snip/internal/store"
	"github.com/spf13/cobra"
)

var listCount int

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
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	dbPath, err := store.DefaultPath()
	if err != nil {
		return fmt.Errorf("resolve db path: %w", err)
	}

	s, err := store.New(dbPath)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer s.Close()

	clips, err := s.List(listCount)
	if err != nil {
		return fmt.Errorf("list clips: %w", err)
	}

	if len(clips) == 0 {
		fmt.Println("No clipboard history yet. Start the daemon and copy something!")
		return nil
	}

	fmt.Printf("%-5s  %-12s  %s\n", "INDEX", "TIME", "PREVIEW")
	fmt.Printf("%-5s  %-12s  %s\n", "-----", "------------", strings.Repeat("-", 40))

	for i, clip := range clips {
		preview := truncate(clip.Content, 80)
		relTime := relativeTime(clip.Timestamp)
		fmt.Printf("%-5d  %-12s  %s\n", i+1, relTime, preview)
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
