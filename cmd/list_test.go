package cmd

import (
	"testing"
	"time"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		max      int
		expected string
	}{
		{"short string unchanged", "hello", 80, "hello"},
		{"exact length unchanged", "hello", 5, "hello"},
		{"long string truncated", "hello world", 7, "hello …"},
		{"newlines replaced", "line1\nline2", 80, "line1 line2"},
		{"tabs replaced", "col1\tcol2", 80, "col1 col2"},
		{"carriage return replaced", "a\rb", 80, "a b"},
		{"unicode truncation", "héllo wörld", 7, "héllo …"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.max)
			if got != tt.expected {
				t.Errorf("truncate(%q, %d) = %q; want %q", tt.input, tt.max, got, tt.expected)
			}
		})
	}
}

func TestRelativeTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		t        time.Time
		expected string
	}{
		{"just now", now.Add(-500 * time.Millisecond), "just now"},
		{"seconds ago", now.Add(-30 * time.Second), "30s ago"},
		{"1 minute ago", now.Add(-90 * time.Second), "1m ago"},
		{"minutes ago", now.Add(-5 * time.Minute), "5m ago"},
		{"1 hour ago", now.Add(-90 * time.Minute), "1h ago"},
		{"hours ago", now.Add(-3 * time.Hour), "3h ago"},
		{"1 day ago", now.Add(-36 * time.Hour), "1d ago"},
		{"days ago", now.Add(-72 * time.Hour), "3d ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := relativeTime(tt.t)
			if got != tt.expected {
				t.Errorf("relativeTime(%v) = %q; want %q", tt.t, got, tt.expected)
			}
		})
	}
}
