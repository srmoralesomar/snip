package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/srmoralesomar/snip/internal/store"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <index>",
	Short: "Remove a clip from history",
	Long:  `Remove a single clip from history by its index (as shown by 'snip list').`,
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	var index int
	if _, err := fmt.Sscanf(args[0], "%d", &index); err != nil || index < 1 {
		return fmt.Errorf("invalid index %q: must be a positive integer", args[0])
	}

	dbPath, err := store.DefaultPath()
	if err != nil {
		return fmt.Errorf("resolve db path: %w", err)
	}

	s, err := store.New(dbPath)
	if err != nil {
		return fmt.Errorf("open history database: %w\nHint: run 'snip start' to start recording clipboard history", err)
	}
	defer s.Close()

	// List enough clips to reach the requested index.
	clips, err := s.List(index)
	if err != nil {
		return fmt.Errorf("list clips: %w", err)
	}

	if len(clips) == 0 {
		return fmt.Errorf("no clipboard history")
	}

	if index > len(clips) {
		return fmt.Errorf("index %d out of range: only %d clip(s) in history", index, len(clips))
	}

	clip := clips[index-1]

	if err := s.Delete(clip.ID); err != nil {
		return fmt.Errorf("delete clip: %w", err)
	}

	preview := truncate(clip.Content, 60)
	color.New(color.FgGreen).Fprintf(os.Stderr, "Deleted clip #%d: %s\n", index, preview)
	return nil
}
