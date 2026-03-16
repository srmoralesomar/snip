package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/omarmorales/snip/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var noColor bool

// Version is the current release version of snip.
const Version = "1.0.0"

var rootCmd = &cobra.Command{
	Use:     "snip",
	Short:   "snip — clipboard history manager",
	Version: Version,
	Long: `snip is a lightweight clipboard history manager.

It runs a background daemon that watches your clipboard and stores
every copied item in a local database (~/.snip/history.db).
Use the CLI commands to search, recall, and paste any entry.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if noColor {
			color.NoColor = true
		}
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable color output")
}

// Execute is the entry point called from main.go.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// loadConfig returns the effective configuration, merging defaults,
// ~/.snip/config.yaml, and any viper values already bound to flags.
func loadConfig() (*config.Config, error) {
	return config.Load(viper.GetViper())
}
