package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/omarmorales/snip/internal/store"
)

func TestClearCmd_Force(t *testing.T) {
	makeTestStore(t, "alpha", "beta", "gamma")

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)
	rootCmd.SetArgs([]string{"clear", "--force"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("clear --force: %v", err)
	}

	home := os.Getenv("HOME")
	s, err := store.New(filepath.Join(home, ".snip", "history.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	n, err := s.Count()
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 clips after clear --force, got %d", n)
	}
}

func TestClearCmd_AbortOnNo(t *testing.T) {
	makeTestStore(t, "keep-me")

	// Reset flag state that may have been set by a previous test.
	clearForce = false

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)
	rootCmd.SetArgs([]string{"clear"})

	// Inject 'n' as stdin so the confirmation prompt is answered negatively.
	rootCmd.SetIn(strings.NewReader("n\n"))
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("clear (aborted): %v", err)
	}

	home := os.Getenv("HOME")
	s, err := store.New(filepath.Join(home, ".snip", "history.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	n, _ := s.Count()
	if n != 1 {
		t.Errorf("expected 1 clip after abort, got %d", n)
	}
}
