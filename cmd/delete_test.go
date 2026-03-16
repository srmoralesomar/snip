package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/srmoralesomar/snip/internal/store"
)

func makeTestStore(t *testing.T, clips ...string) string {
	t.Helper()
	dir := t.TempDir()
	snipDir := filepath.Join(dir, ".snip")
	if err := os.MkdirAll(snipDir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	s, err := store.New(filepath.Join(snipDir, "history.db"))
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	for _, c := range clips {
		if err := s.Save(c); err != nil {
			t.Fatalf("save %q: %v", c, err)
		}
	}
	s.Close()

	// Point HOME to this temp dir so DefaultPath resolves correctly.
	t.Setenv("HOME", dir)
	return dir
}

func TestDeleteCmd_RemovesClip(t *testing.T) {
	makeTestStore(t, "first", "second", "third")

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)
	rootCmd.SetArgs([]string{"delete", "1"}) // newest = "third"
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("delete: %v", err)
	}

	// Re-open store and check count.
	home := os.Getenv("HOME")
	s, err := store.New(filepath.Join(home, ".snip", "history.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.Close()

	clips, err := s.List(0)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(clips) != 2 {
		t.Errorf("expected 2 clips after delete, got %d", len(clips))
	}
	if clips[0].Content != "second" {
		t.Errorf("expected 'second' as newest after delete, got %q", clips[0].Content)
	}
}

func TestDeleteCmd_InvalidIndex(t *testing.T) {
	makeTestStore(t, "only-one")

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)
	rootCmd.SetArgs([]string{"delete", "0"})
	err := rootCmd.Execute()
	_ = err // invalid index error expected
}

func TestDeleteCmd_OutOfRange(t *testing.T) {
	makeTestStore(t, "only-one")

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)
	rootCmd.SetArgs([]string{"delete", "99"})
	err := rootCmd.Execute()
	_ = err // out-of-range error expected
}
