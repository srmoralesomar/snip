package pidfile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// DefaultPath returns the default path for the daemon PID file (~/.snip/daemon.pid).
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".snip", "daemon.pid"), nil
}

// Write writes pid to the file at path, creating the parent directory if needed.
func Write(path string, pid int) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create pid dir: %w", err)
	}
	return os.WriteFile(path, []byte(strconv.Itoa(pid)+"\n"), 0o600)
}

// Read reads and parses the PID stored in the file at path.
// Returns 0, os.ErrNotExist if the file does not exist.
func Read(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("invalid pid in %s: %w", path, err)
	}
	return pid, nil
}

// Remove deletes the PID file. Returns nil if the file does not exist.
func Remove(path string) error {
	err := os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

// IsRunning returns true if the process with the given PID is currently alive.
// On Unix, this sends signal 0 to the process to check its existence.
func IsRunning(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}

// StartTime returns the modification time of the PID file, used as a proxy
// for when the daemon was started.
func StartTime(path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}
