package cmd

import (
	"strings"
	"testing"
)

func TestSingleLine(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello world", "hello world"},
		{"line1\nline2", "line1 line2"},
		{"col1\tcol2", "col1 col2"},
		{"a\r\nb", "a  b"},
		{"no special chars", "no special chars"},
	}

	for _, tt := range tests {
		got := singleLine(tt.input)
		if got != tt.want {
			t.Errorf("singleLine(%q) = %q; want %q", tt.input, got, tt.want)
		}
	}
}

func TestHighlightMatches(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		indexes  []int
		wantHas  []string
		wantNot  []string
	}{
		{
			name:    "no matches — unchanged",
			s:       "hello",
			indexes: nil,
			wantHas: []string{"hello"},
			wantNot: []string{ansiYellow},
		},
		{
			name:    "single char match",
			s:       "hello",
			indexes: []int{0},
			wantHas: []string{ansiBold, ansiYellow, "h", ansiReset, "ello"},
		},
		{
			name:    "all chars matched",
			s:       "hi",
			indexes: []int{0, 1},
			wantHas: []string{ansiBold, ansiYellow, "h", "i", ansiReset},
		},
		{
			name:    "out-of-range index ignored",
			s:       "hi",
			indexes: []int{0, 100},
			wantHas: []string{ansiBold, ansiYellow, "h", ansiReset, "i"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := highlightMatches(tt.s, tt.indexes)
			for _, want := range tt.wantHas {
				if !strings.Contains(got, want) {
					t.Errorf("highlightMatches(%q, %v) = %q; expected to contain %q", tt.s, tt.indexes, got, want)
				}
			}
			for _, notWant := range tt.wantNot {
				if strings.Contains(got, notWant) {
					t.Errorf("highlightMatches(%q, %v) = %q; expected NOT to contain %q", tt.s, tt.indexes, got, notWant)
				}
			}
		})
	}
}
