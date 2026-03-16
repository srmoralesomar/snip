package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/srmoralesomar/snip/internal/pidfile"
	"github.com/srmoralesomar/snip/internal/store"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show daemon status",
	Long:  `Show whether the snip daemon is running, along with its PID, uptime, and clip count.`,
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	pidPath, err := pidfile.DefaultPath()
	if err != nil {
		return err
	}

	pid, err := pidfile.Read(pidPath)
	if errors.Is(err, os.ErrNotExist) || (err == nil && !pidfile.IsRunning(pid)) {
		color.New(color.FgYellow).Println("daemon: not running")
		if err == nil {
			// Stale PID file — clean it up silently.
			_ = pidfile.Remove(pidPath)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("read pid file: %w", err)
	}

	// Uptime via PID file modification time.
	uptime := "unknown"
	if startTime, stErr := pidfile.StartTime(pidPath); stErr == nil {
		uptime = formatUptime(time.Since(startTime))
	}

	bold := color.New(color.Bold)
	bold.Print("daemon: ")
	color.New(color.FgGreen).Println("running")
	fmt.Printf("  PID:    %d\n", pid)
	fmt.Printf("  uptime: %s\n", uptime)

	// Clip count — best-effort, don't fail status if store is unavailable.
	cfg, cfgErr := loadConfig()
	if cfgErr == nil {
		s, storeErr := store.New(cfg.StoragePath)
		if storeErr == nil {
			if count, countErr := s.Count(); countErr == nil {
				fmt.Printf("  clips:  %d\n", count)
			}
			s.Close()
		}
	}

	return nil
}

// formatUptime formats a duration into a human-readable uptime string.
func formatUptime(d time.Duration) string {
	d = d.Truncate(time.Second)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
