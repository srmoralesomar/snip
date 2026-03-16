package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/omarmorales/snip/internal/clipboard"
	"github.com/spf13/cobra"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Start the clipboard watcher daemon",
	Long: `Start the clipboard watcher daemon.

The daemon polls the system clipboard every 500ms and prints each new clip
to stdout. Press Ctrl+C to stop.`,
	RunE: runDaemon,
}

func init() {
	rootCmd.AddCommand(daemonCmd)
}

func runDaemon(cmd *cobra.Command, args []string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	watcher := clipboard.NewWatcher(clipboard.SystemReader{}, 500*time.Millisecond)
	go watcher.Start(ctx)

	fmt.Fprintln(os.Stderr, "snip daemon started — watching clipboard (Ctrl+C to stop)")

	for clip := range watcher.Clips() {
		fmt.Printf("[%s] %s\n", clip.Timestamp.Format("15:04:05"), clip.Content)
	}

	fmt.Fprintln(os.Stderr, "snip daemon stopped")
	return nil
}
