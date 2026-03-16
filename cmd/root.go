package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "snip",
	Short: "snip — clipboard history manager",
	Long: `snip is a lightweight clipboard history manager.

It runs a background daemon that watches your clipboard and stores
every copied item in a local database (~/.snip/history.db).
Use the CLI commands to search, recall, and paste any entry.`,
}

// Execute is the entry point called from main.go.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
