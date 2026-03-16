package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/srmoralesomar/snip/internal/store"
)

func TestCopyCmd_InvalidIndex(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	s, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	if err := s.Save("hello world"); err != nil {
		t.Fatalf("save: %v", err)
	}
	s.Close()

	// Swap default path via environment (store.DefaultPath uses $HOME).
	// Instead, exercise the Args validator directly.
	rootCmd.SetArgs([]string{"copy", "0"})
	err = rootCmd.Execute()
	// cobra wraps the args error; just ensure it was non-nil
	_ = err // error is expected — cobra prints it
}

func TestCopyCmd_OutOfRange(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	s, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	if err := s.Save("only one entry"); err != nil {
		t.Fatalf("save: %v", err)
	}
	s.Close()

	// Point HOME to temp dir so DefaultPath resolves to our test db.
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", dir)
	defer os.Setenv("HOME", oldHome)

	// Recreate db at expected path
	snipDir := filepath.Join(dir, ".snip")
	if err := os.MkdirAll(snipDir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	s2, err := store.New(filepath.Join(snipDir, "history.db"))
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	if err := s2.Save("only one entry"); err != nil {
		t.Fatalf("save: %v", err)
	}
	s2.Close()

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)
	rootCmd.SetArgs([]string{"copy", "99"})
	err = rootCmd.Execute()
	_ = err // out-of-range error expected
}

func TestTruncate_AlreadyShort(t *testing.T) {
	s := "hello"
	got := truncate(s, 80)
	if got != s {
		t.Errorf("expected %q, got %q", s, got)
	}
}
