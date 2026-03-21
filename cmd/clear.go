package cmd

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/srmoralesomar/snip/internal/store"
)

var clearForce bool

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Delete all clipboard history",
	Long: `Delete all stored clipboard entries.

You will be prompted for confirmation unless --force is provided.`,
	Args: cobra.NoArgs,
	RunE: runClear,
}

func init() {
	clearCmd.Flags().BoolVarP(&clearForce, "force", "f", false, "Skip confirmation prompt")
	rootCmd.AddCommand(clearCmd)
}

func runClear(cmd *cobra.Command, args []string) error {
	dbPath, err := store.DefaultPath()
	if err != nil {
		return fmt.Errorf("resolve db path: %w", err)
	}

	s, err := store.New(dbPath)
	if err != nil {
		return fmt.Errorf("open history database: %w\nHint: run 'snip start' to start recording clipboard history", err)
	}
	defer s.Close()

	if !clearForce {
		color.New(color.FgYellow).Fprint(cmd.ErrOrStderr(), "This will delete all clipboard history. Continue? [y/N] ")
		reader := bufio.NewReader(cmd.InOrStdin())
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Fprintln(cmd.ErrOrStderr(), "Aborted.")
			return nil
		}
	}

	if err := s.Clear(); err != nil {
		return fmt.Errorf("clear history: %w", err)
	}

	color.New(color.FgGreen).Fprintln(cmd.ErrOrStderr(), "Clipboard history cleared.")
	return nil
}
