package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	s, err := New(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestNew_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "a", "b", "c")
	s, err := New(filepath.Join(nested, "history.db"))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer s.Close()

	if _, err := os.Stat(nested); os.IsNotExist(err) {
		t.Fatal("expected directory to be created")
	}
}

func TestSave_BasicRoundtrip(t *testing.T) {
	s := newTestStore(t)

	if err := s.Save("hello world"); err != nil {
		t.Fatalf("Save: %v", err)
	}

	clips, err := s.List(10)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(clips) != 1 {
		t.Fatalf("expected 1 clip, got %d", len(clips))
	}
	if clips[0].Content != "hello world" {
		t.Errorf("content = %q, want %q", clips[0].Content, "hello world")
	}
	if clips[0].Hash == "" {
		t.Error("expected non-empty hash")
	}
	if clips[0].Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestSave_EmptyContentIgnored(t *testing.T) {
	s := newTestStore(t)

	if err := s.Save(""); err != nil {
		t.Fatalf("Save empty: %v", err)
	}

	clips, err := s.List(0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(clips) != 0 {
		t.Errorf("expected 0 clips, got %d", len(clips))
	}
}

func TestSave_NoDuplicateConsecutive(t *testing.T) {
	s := newTestStore(t)

	for i := 0; i < 3; i++ {
		if err := s.Save("duplicate"); err != nil {
			t.Fatalf("Save iteration %d: %v", i, err)
		}
	}

	clips, err := s.List(0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(clips) != 1 {
		t.Errorf("expected 1 clip (dedup), got %d", len(clips))
	}
}

func TestSave_AllowsNonConsecutiveDuplicate(t *testing.T) {
	s := newTestStore(t)

	if err := s.Save("alpha"); err != nil {
		t.Fatalf("Save alpha: %v", err)
	}
	if err := s.Save("beta"); err != nil {
		t.Fatalf("Save beta: %v", err)
	}
	if err := s.Save("alpha"); err != nil {
		t.Fatalf("Save alpha again: %v", err)
	}

	clips, err := s.List(0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(clips) != 3 {
		t.Errorf("expected 3 clips, got %d", len(clips))
	}
}

func TestList_NewestFirst(t *testing.T) {
	s := newTestStore(t)

	contents := []string{"first", "second", "third"}
	for _, c := range contents {
		if err := s.Save(c); err != nil {
			t.Fatalf("Save %q: %v", c, err)
		}
		// small sleep so timestamps differ
		time.Sleep(time.Millisecond)
	}

	clips, err := s.List(0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(clips) != 3 {
		t.Fatalf("expected 3 clips, got %d", len(clips))
	}
	if clips[0].Content != "third" {
		t.Errorf("expected newest first, got %q", clips[0].Content)
	}
	if clips[2].Content != "first" {
		t.Errorf("expected oldest last, got %q", clips[2].Content)
	}
}

func TestList_LimitsResults(t *testing.T) {
	s := newTestStore(t)

	for i := 0; i < 5; i++ {
		if err := s.Save(string(rune('a' + i))); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	clips, err := s.List(3)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(clips) != 3 {
		t.Errorf("expected 3 clips, got %d", len(clips))
	}
}

func TestGet(t *testing.T) {
	s := newTestStore(t)

	if err := s.Save("get-me"); err != nil {
		t.Fatalf("Save: %v", err)
	}

	clips, err := s.List(1)
	if err != nil || len(clips) == 0 {
		t.Fatalf("List: %v / len=%d", err, len(clips))
	}
	id := clips[0].ID

	clip, err := s.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if clip.Content != "get-me" {
		t.Errorf("content = %q, want %q", clip.Content, "get-me")
	}
}

func TestGet_NotFound(t *testing.T) {
	s := newTestStore(t)

	_, err := s.Get(9999)
	if err == nil {
		t.Fatal("expected error for missing clip")
	}
}

func TestDelete(t *testing.T) {
	s := newTestStore(t)

	if err := s.Save("to-delete"); err != nil {
		t.Fatalf("Save: %v", err)
	}

	clips, _ := s.List(1)
	id := clips[0].ID

	if err := s.Delete(id); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := s.Get(id)
	if err == nil {
		t.Fatal("expected error after deletion")
	}
}

func TestDelete_NotFound(t *testing.T) {
	s := newTestStore(t)

	if err := s.Delete(9999); err == nil {
		t.Fatal("expected error deleting missing clip")
	}
}

func TestPrune(t *testing.T) {
	s := newTestStore(t)

	for i := 0; i < 10; i++ {
		if err := s.Save(string(rune('a' + i))); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	if err := s.Prune(5); err != nil {
		t.Fatalf("Prune: %v", err)
	}

	n, err := s.Count()
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if n != 5 {
		t.Errorf("expected 5 clips after prune, got %d", n)
	}
}

func TestPrune_KeepsNewest(t *testing.T) {
	s := newTestStore(t)

	contents := []string{"old1", "old2", "old3", "new1", "new2"}
	for _, c := range contents {
		if err := s.Save(c); err != nil {
			t.Fatalf("Save %q: %v", c, err)
		}
	}

	if err := s.Prune(2); err != nil {
		t.Fatalf("Prune: %v", err)
	}

	clips, err := s.List(0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(clips) != 2 {
		t.Fatalf("expected 2 clips, got %d", len(clips))
	}
	// newest first
	if clips[0].Content != "new2" || clips[1].Content != "new1" {
		t.Errorf("unexpected contents: %v, %v", clips[0].Content, clips[1].Content)
	}
}

func TestPrune_NoopWhenUnderLimit(t *testing.T) {
	s := newTestStore(t)

	if err := s.Save("only-one"); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if err := s.Prune(100); err != nil {
		t.Fatalf("Prune: %v", err)
	}

	n, err := s.Count()
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if n != 1 {
		t.Errorf("expected 1 clip, got %d", n)
	}
}

func TestPrune_ZeroIsNoop(t *testing.T) {
	s := newTestStore(t)

	for i := 0; i < 3; i++ {
		if err := s.Save(string(rune('a' + i))); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	if err := s.Prune(0); err != nil {
		t.Fatalf("Prune(0): %v", err)
	}

	n, _ := s.Count()
	if n != 3 {
		t.Errorf("expected 3 clips, got %d", n)
	}
}

func TestClear(t *testing.T) {
	s := newTestStore(t)

	for i := 0; i < 5; i++ {
		if err := s.Save(string(rune('a' + i))); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	if err := s.Clear(); err != nil {
		t.Fatalf("Clear: %v", err)
	}

	n, err := s.Count()
	if err != nil {
		t.Fatalf("Count after Clear: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 clips after Clear, got %d", n)
	}

	// Saving after Clear should work (dedup hash reset).
	if err := s.Save("after-clear"); err != nil {
		t.Fatalf("Save after Clear: %v", err)
	}
	clips, _ := s.List(0)
	if len(clips) != 1 {
		t.Errorf("expected 1 clip after save post-clear, got %d", len(clips))
	}
}

func TestCount(t *testing.T) {
	s := newTestStore(t)

	n, err := s.Count()
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0, got %d", n)
	}

	for i := 0; i < 4; i++ {
		s.Save(string(rune('a' + i)))
	}

	n, err = s.Count()
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if n != 4 {
		t.Errorf("expected 4, got %d", n)
	}
}
