package pidfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteAndRead(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.pid")

	if err := Write(path, 12345); err != nil {
		t.Fatalf("Write: %v", err)
	}

	pid, err := Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if pid != 12345 {
		t.Errorf("got pid %d, want 12345", pid)
	}
}

func TestWriteCreatesParentDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "nested", "test.pid")

	if err := Write(path, 99); err != nil {
		t.Fatalf("Write: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Errorf("pid file not created: %v", err)
	}
}

func TestReadNotExist(t *testing.T) {
	_, err := Read("/nonexistent/path/daemon.pid")
	if err == nil {
		t.Fatal("expected error reading nonexistent file")
	}
	if !os.IsNotExist(err) {
		t.Errorf("expected ErrNotExist, got: %v", err)
	}
}

func TestRemove(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.pid")

	if err := Write(path, 1); err != nil {
		t.Fatalf("Write: %v", err)
	}

	if err := Remove(path); err != nil {
		t.Fatalf("Remove: %v", err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("pid file should be removed")
	}
}

func TestRemoveNotExist(t *testing.T) {
	// Should not return an error when file doesn't exist.
	if err := Remove("/nonexistent/daemon.pid"); err != nil {
		t.Errorf("Remove on nonexistent file: %v", err)
	}
}

func TestIsRunning(t *testing.T) {
	// Our own PID should be running.
	if !IsRunning(os.Getpid()) {
		t.Error("expected current process to be running")
	}

	// A very large PID is very unlikely to be a real process.
	if IsRunning(99999999) {
		t.Skip("PID 99999999 unexpectedly exists, skipping")
	}
}

func TestStartTime(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.pid")

	if err := Write(path, 1); err != nil {
		t.Fatalf("Write: %v", err)
	}

	ts, err := StartTime(path)
	if err != nil {
		t.Fatalf("StartTime: %v", err)
	}
	if ts.IsZero() {
		t.Error("expected non-zero start time")
	}
}
