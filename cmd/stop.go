package cmd

import (
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/srmoralesomar/snip/internal/pidfile"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the running daemon",
	Long:  `Send SIGTERM to the running snip daemon to shut it down cleanly.`,
	RunE:  runStop,
}

func init() {
	rootCmd.AddCommand(stopCmd)
}

func runStop(cmd *cobra.Command, args []string) error {
	pidPath, err := pidfile.DefaultPath()
	if err != nil {
		return err
	}

	pid, err := pidfile.Read(pidPath)
	if errors.Is(err, os.ErrNotExist) {
		color.New(color.FgYellow).Fprintln(os.Stderr, "daemon is not running")
		return nil
	}
	if err != nil {
		return fmt.Errorf("read pid file: %w", err)
	}

	if !pidfile.IsRunning(pid) {
		color.New(color.FgYellow).Fprintln(os.Stderr, "daemon is not running (stale pid file removed)")
		_ = pidfile.Remove(pidPath)
		return nil
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find process %d: %w", pid, err)
	}
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("send SIGTERM to pid %d: %w", pid, err)
	}

	color.New(color.FgGreen).Fprintf(os.Stderr, "sent SIGTERM to daemon (PID %d)\n", pid)
	return nil
}
