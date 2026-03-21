package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/srmoralesomar/snip/internal/clipboard"
	"github.com/srmoralesomar/snip/internal/pidfile"
	"github.com/srmoralesomar/snip/internal/store"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the clipboard watcher",
	Long: `Start the long-running clipboard watcher.

It polls the system clipboard and saves each new clip to the local history
database. Press Ctrl+C or run "snip stop" to shut down.

To run detached from the terminal, use your shell (e.g. "&", nohup) or a
process supervisor (launchd, systemd --user, tmux).`,
	RunE: runStart,
}

func init() {
	startCmd.Flags().Int("max-history", 0, "Maximum number of clips to keep in history (0 = use config/default)")
	startCmd.Flags().Int("poll-interval-ms", 0, "Clipboard poll interval in milliseconds (0 = use config/default)")
	startCmd.Flags().String("storage-path", "", "Path to the history database (empty = use config/default)")

	// Bind flags to viper keys so config file values are overridden by flags.
	_ = viper.BindPFlag("max_history", startCmd.Flags().Lookup("max-history"))
	_ = viper.BindPFlag("poll_interval_ms", startCmd.Flags().Lookup("poll-interval-ms"))
	_ = viper.BindPFlag("storage_path", startCmd.Flags().Lookup("storage-path"))

	rootCmd.AddCommand(startCmd)
}

func runStart(cmd *cobra.Command, args []string) error {
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

	// PID file: prevent multiple watcher instances.
	pidPath, err := pidfile.DefaultPath()
	if err != nil {
		return fmt.Errorf("pid file path: %w", err)
	}
	if pid, readErr := pidfile.Read(pidPath); readErr == nil && pidfile.IsRunning(pid) {
		return fmt.Errorf("snip is already running (PID %d)", pid)
	}
	if err := pidfile.Write(pidPath, os.Getpid()); err != nil {
		return fmt.Errorf("write pid file: %w", err)
	}
	defer pidfile.Remove(pidPath) //nolint:errcheck

	s, err := store.New(dbPath)
	if err != nil {
		return fmt.Errorf("open history database: %w", err)
	}
	s.Close() // Close immediately; we will open it only when saving a new clip.

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pollInterval := time.Duration(pollIntervalMs) * time.Millisecond
	watcher := clipboard.NewWatcher(clipboard.SystemReader{}, pollInterval)
	go watcher.Start(ctx)

	color.New(color.FgGreen).Fprintln(os.Stderr, "snip start: watching clipboard")

	for clip := range watcher.Clips() {
		saveAndPrune(dbPath, clip.Content, maxHistory)
	}

	color.New(color.FgYellow).Fprintln(os.Stderr, "snip start: stopped")
	return nil
}

func saveAndPrune(dbPath string, content string, maxHistory int) {
	s, err := store.New(dbPath)
	if err != nil {
		color.New(color.FgRed).Fprintf(os.Stderr, "open store error: %v\n", err)
		return
	}
	defer s.Close()

	if saveErr := s.Save(content); saveErr != nil {
		color.New(color.FgRed).Fprintf(os.Stderr, "store error: %v\n", saveErr)
	}
	if pruneErr := s.Prune(maxHistory); pruneErr != nil {
		color.New(color.FgRed).Fprintf(os.Stderr, "prune error: %v\n", pruneErr)
	}
}
