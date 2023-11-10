package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"
)

// CmdList defines the CLI sub-command 'list'.
var CmdList = &cobra.Command{
	Use:   "list [flags]",
	Short: "...",
	Args:  cobra.NoArgs,
	Run:   runList,
}

// Initialize CLI options.
func init() {
}

// runList is called when the "test" command is used.
func runList(_ *cobra.Command, _ []string) {
	slog.Info("Done")
}
