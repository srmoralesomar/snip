package cmd

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/omarmorales/snip/internal/store"
	"github.com/spf13/cobra"
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
		return fmt.Errorf("open store: %w", err)
	}
	defer s.Close()

	if !clearForce {
		fmt.Fprint(cmd.ErrOrStderr(), "This will delete all clipboard history. Continue? [y/N] ")
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

	fmt.Fprintln(cmd.ErrOrStderr(), "Clipboard history cleared.")
	return nil
}
