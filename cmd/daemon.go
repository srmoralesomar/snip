package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/omarmorales/snip/internal/clipboard"
	"github.com/omarmorales/snip/internal/store"
	"github.com/spf13/cobra"
)

var maxHistory int

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Start the clipboard watcher daemon",
	Long: `Start the clipboard watcher daemon.

The daemon polls the system clipboard every 500ms, saves each new clip to
the local history database, and runs silently in the background.
Press Ctrl+C to stop.`,
	RunE: runDaemon,
}

func init() {
	daemonCmd.Flags().IntVar(&maxHistory, "max-history", 500, "Maximum number of clips to keep in history")
	rootCmd.AddCommand(daemonCmd)
}

func runDaemon(cmd *cobra.Command, args []string) error {
	dbPath, err := store.DefaultPath()
	if err != nil {
		return fmt.Errorf("resolve db path: %w", err)
	}

	s, err := store.New(dbPath)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer s.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	watcher := clipboard.NewWatcher(clipboard.SystemReader{}, 500*time.Millisecond)
	go watcher.Start(ctx)

	fmt.Fprintln(os.Stderr, "snip daemon started")

	for clip := range watcher.Clips() {
		if saveErr := s.Save(clip.Content); saveErr != nil {
			fmt.Fprintf(os.Stderr, "store error: %v\n", saveErr)
			continue
		}
		if pruneErr := s.Prune(maxHistory); pruneErr != nil {
			fmt.Fprintf(os.Stderr, "prune error: %v\n", pruneErr)
		}
	}

	fmt.Fprintln(os.Stderr, "snip daemon stopped")
	return nil
}
