package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/omarmorales/snip/internal/clipboard"
	"github.com/omarmorales/snip/internal/pidfile"
	"github.com/omarmorales/snip/internal/store"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Start the clipboard watcher daemon",
	Long: `Start the clipboard watcher daemon.

The daemon polls the system clipboard, saves each new clip to the local
history database, and runs silently in the background.
Press Ctrl+C to stop.`,
	RunE: runDaemon,
}

func init() {
	daemonCmd.Flags().Int("max-history", 0, "Maximum number of clips to keep in history (0 = use config/default)")
	daemonCmd.Flags().Int("poll-interval-ms", 0, "Clipboard poll interval in milliseconds (0 = use config/default)")
	daemonCmd.Flags().String("storage-path", "", "Path to the history database (empty = use config/default)")

	// Bind flags to viper keys so config file values are overridden by flags.
	_ = viper.BindPFlag("max_history", daemonCmd.Flags().Lookup("max-history"))
	_ = viper.BindPFlag("poll_interval_ms", daemonCmd.Flags().Lookup("poll-interval-ms"))
	_ = viper.BindPFlag("storage_path", daemonCmd.Flags().Lookup("storage-path"))

	rootCmd.AddCommand(daemonCmd)
}

func runDaemon(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// CLI flag values of 0/"" mean "not set by user" — viper already applied
	// the right precedence (flag > config file > default), but zero-value flags
	// registered with cobra would override config-file values.  We handle this
	// by only overriding the config when the flag was explicitly changed.
	maxHistory := cfg.MaxHistory
	if f := cmd.Flags().Lookup("max-history"); f.Changed {
		v, _ := cmd.Flags().GetInt("max-history")
		maxHistory = v
	}

	pollIntervalMs := cfg.PollIntervalMs
	if f := cmd.Flags().Lookup("poll-interval-ms"); f.Changed {
		v, _ := cmd.Flags().GetInt("poll-interval-ms")
		pollIntervalMs = v
	}

	dbPath := cfg.StoragePath
	if f := cmd.Flags().Lookup("storage-path"); f.Changed {
		dbPath, _ = cmd.Flags().GetString("storage-path")
	}

	// PID file: prevent multiple daemon instances.
	pidPath, err := pidfile.DefaultPath()
	if err != nil {
		return fmt.Errorf("pid file path: %w", err)
	}
	if pid, readErr := pidfile.Read(pidPath); readErr == nil && pidfile.IsRunning(pid) {
		return fmt.Errorf("daemon is already running (PID %d)", pid)
	}
	if err := pidfile.Write(pidPath, os.Getpid()); err != nil {
		return fmt.Errorf("write pid file: %w", err)
	}
	defer pidfile.Remove(pidPath) //nolint:errcheck

	s, err := store.New(dbPath)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer s.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pollInterval := time.Duration(pollIntervalMs) * time.Millisecond
	watcher := clipboard.NewWatcher(clipboard.SystemReader{}, pollInterval)
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
