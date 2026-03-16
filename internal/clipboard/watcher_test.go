package clipboard

import (
	"context"
	"testing"
	"time"
)

// mockReader is a controllable clipboard reader for tests.
type mockReader struct {
	values []string
	index  int
}

func (m *mockReader) ReadAll() (string, error) {
	if m.index >= len(m.values) {
		return m.values[len(m.values)-1], nil
	}
	v := m.values[m.index]
	m.index++
	return v, nil
}

func TestWatcher_EmitsNewClips(t *testing.T) {
	reader := &mockReader{values: []string{"hello", "hello", "world", "world"}}
	w := NewWatcher(reader, 10*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	go w.Start(ctx)

	var got []string
	for clip := range w.Clips() {
		got = append(got, clip.Content)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 clips, got %d: %v", len(got), got)
	}
	if got[0] != "hello" {
		t.Errorf("first clip: want 'hello', got %q", got[0])
	}
	if got[1] != "world" {
		t.Errorf("second clip: want 'world', got %q", got[1])
	}
}

func TestWatcher_SkipsDuplicates(t *testing.T) {
	reader := &mockReader{values: []string{"same", "same", "same"}}
	w := NewWatcher(reader, 10*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go w.Start(ctx)

	var got []string
	for clip := range w.Clips() {
		got = append(got, clip.Content)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 clip (no duplicates), got %d: %v", len(got), got)
	}
}

func TestWatcher_SkipsEmptyContent(t *testing.T) {
	reader := &mockReader{values: []string{"", "", "nonempty"}}
	w := NewWatcher(reader, 10*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go w.Start(ctx)

	var got []string
	for clip := range w.Clips() {
		got = append(got, clip.Content)
	}

	if len(got) != 1 || got[0] != "nonempty" {
		t.Fatalf("expected only 'nonempty', got %v", got)
	}
}

func TestWatcher_ClipHasTimestamp(t *testing.T) {
	before := time.Now()
	reader := &mockReader{values: []string{"ts-test"}}
	w := NewWatcher(reader, 10*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go w.Start(ctx)

	clip := <-w.Clips()
	after := time.Now()

	if clip.Timestamp.Before(before) || clip.Timestamp.After(after) {
		t.Errorf("clip timestamp %v is outside [%v, %v]", clip.Timestamp, before, after)
	}
}

func TestWatcher_ShutdownOnCancel(t *testing.T) {
	reader := &mockReader{values: []string{"a"}}
	w := NewWatcher(reader, 50*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		w.Start(ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// ok
	case <-time.After(500*time.Millisecond):
		t.Fatal("watcher did not shut down after context cancellation")
	}
}
