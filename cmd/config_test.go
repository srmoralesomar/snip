package cmd

import (
	"bytes"
	"testing"
	"time"
)

func TestRunConfig(t *testing.T) {
	makeTestStore(t) // sets HOME so loadConfig works cleanly

	out := &bytes.Buffer{}
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)
	rootCmd.SetArgs([]string{"config"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("config: %v", err)
	}
	// runConfig prints via fmt.Printf to os.Stdout directly;
	// just verifying no error is sufficient here.
}

func TestFormatUptime(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{30 * time.Second, "30s"},
		{90 * time.Second, "1m 30s"},
		{3600*time.Second + 15*time.Minute + 5*time.Second, "1h 15m 5s"},
		{2*time.Hour + 0*time.Minute + 0*time.Second, "2h 0m 0s"},
	}

	for _, tt := range tests {
		got := formatUptime(tt.d)
		if got != tt.want {
			t.Errorf("formatUptime(%v) = %q; want %q", tt.d, got, tt.want)
		}
	}
}
