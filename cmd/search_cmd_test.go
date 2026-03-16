package cmd

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

func TestRunSearch_Empty(t *testing.T) {
	makeTestStore(t) // no clips

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)
	rootCmd.SetArgs([]string{"search", "anything"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("search (empty): %v", err)
	}
}

func TestRunSearch_NoMatch(t *testing.T) {
	makeTestStore(t, "hello world", "foo bar")

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)
	rootCmd.SetArgs([]string{"search", "zzzzzzzzz"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("search (no match): %v", err)
	}
}

func TestRunSearch_WithMatches(t *testing.T) {
	makeTestStore(t, "hello world", "help me", "world peace")

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)
	rootCmd.SetArgs([]string{"search", "hel"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("search: %v", err)
	}
}

func TestRunSearch_JSON(t *testing.T) {
	makeTestStore(t, "hello world", "help me")

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	searchJSON = true
	rootCmd.SetArgs([]string{"search", "--json", "hel"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("search --json: %v", err)
	}
	searchJSON = false

	var results []searchClipJSON
	if err := json.Unmarshal(buf.Bytes(), &results); err != nil {
		t.Fatalf("unmarshal search JSON: %v (raw: %q)", err, buf.String())
	}

	if len(results) == 0 {
		t.Error("expected at least one search result")
	}
	// Timestamp should be RFC3339
	if _, err := time.Parse(time.RFC3339, results[0].Timestamp); err != nil {
		t.Errorf("timestamp not RFC3339: %q", results[0].Timestamp)
	}
}

func TestRunSearch_LimitFlag(t *testing.T) {
	makeTestStore(t, "hello one", "hello two", "hello three", "hello four")

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)
	rootCmd.SetArgs([]string{"search", "--limit", "2", "hello"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("search --limit: %v", err)
	}
}
