package cmd

import (
	"fmt"
	"os"

	"github.com/atotto/clipboard"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/srmoralesomar/snip/internal/store"
)

var copyLast bool

var copyCmd = &cobra.Command{
	Use:   "copy [index]",
	Short: "Copy a clip back to the clipboard",
	Long: `Copy a clip from history back to the system clipboard.

Provide an index (as shown by 'snip list'), or use --last to re-copy
the most recent entry.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if copyLast {
			if len(args) > 0 {
				return fmt.Errorf("cannot use --last together with an index argument")
			}
			return nil
		}
		if len(args) != 1 {
			return fmt.Errorf("requires an index argument (or --last)")
		}
		return nil
	},
	RunE: runCopy,
}

func init() {
	copyCmd.Flags().BoolVar(&copyLast, "last", false, "Copy the most recent clip")
	rootCmd.AddCommand(copyCmd)
}

func runCopy(cmd *cobra.Command, args []string) error {
	dbPath, err := store.DefaultPath()
	if err != nil {
		return fmt.Errorf("resolve db path: %w", err)
	}

	s, err := store.New(dbPath)
	if err != nil {
		return fmt.Errorf("open history database: %w\nHint: run 'snip daemon' to start recording clipboard history", err)
	}
	defer s.Close()

	var index int
	if copyLast {
		index = 1
	} else {
		if _, err := fmt.Sscanf(args[0], "%d", &index); err != nil || index < 1 {
			return fmt.Errorf("invalid index %q: must be a positive integer", args[0])
		}
	}

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

	if err := clipboard.WriteAll(clip.Content); err != nil {
		return fmt.Errorf("write to clipboard: %w", err)
	}

	preview := truncate(clip.Content, 60)
	color.New(color.FgGreen).Fprintf(os.Stderr, "Copied clip #%d to clipboard: %s\n", index, preview)
	return nil
}
