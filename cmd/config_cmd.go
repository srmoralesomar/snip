package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show current effective configuration",
	Long: `Show the current effective configuration for snip.

Values reflect the merged result of defaults, ~/.snip/config.yaml (if present),
and any CLI flags passed to this invocation.`,
	RunE: runConfig,
}

func init() {
	rootCmd.AddCommand(configCmd)
}

func runConfig(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	fmt.Printf("max_history:      %d\n", cfg.MaxHistory)
	fmt.Printf("poll_interval_ms: %d\n", cfg.PollIntervalMs)
	fmt.Printf("storage_path:     %s\n", cfg.StoragePath)
	return nil
}
