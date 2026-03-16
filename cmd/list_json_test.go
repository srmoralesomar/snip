package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestRunList_Empty(t *testing.T) {
	makeTestStore(t) // no clips

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)
	rootCmd.SetArgs([]string{"list"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("list (empty): %v", err)
	}
}

func TestRunList_WithClips(t *testing.T) {
	makeTestStore(t, "hello world", "second clip", "third clip")

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)
	rootCmd.SetArgs([]string{"list"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("list: %v", err)
	}
}

func TestRunList_CountFlag(t *testing.T) {
	makeTestStore(t, "a", "b", "c", "d", "e")

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)
	listCount = 2
	rootCmd.SetArgs([]string{"list", "--count", "2"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("list --count: %v", err)
	}
	listCount = 20 // reset
}

func TestRunList_JSON(t *testing.T) {
	makeTestStore(t, "clip one", "clip two")

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	listJSON = true
	rootCmd.SetArgs([]string{"list", "--json"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("list --json: %v", err)
	}
	listJSON = false

	var clips []listClipJSON
	if err := json.Unmarshal(buf.Bytes(), &clips); err != nil {
		// buf may have color codes prefix — find the '[' start
		s := buf.String()
		idx := strings.Index(s, "[")
		if idx < 0 {
			t.Fatalf("no JSON array in output: %q", s)
		}
		if err2 := json.Unmarshal([]byte(s[idx:]), &clips); err2 != nil {
			t.Fatalf("unmarshal JSON: %v (raw: %q)", err2, s)
		}
	}

	if len(clips) != 2 {
		t.Errorf("expected 2 clips, got %d", len(clips))
	}
	if clips[0].Index != 1 {
		t.Errorf("expected index 1, got %d", clips[0].Index)
	}
	// Timestamp should be RFC3339
	if _, err := time.Parse(time.RFC3339, clips[0].Timestamp); err != nil {
		t.Errorf("timestamp not RFC3339: %q", clips[0].Timestamp)
	}
}
