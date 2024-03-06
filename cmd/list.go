package cmd

import (
	"encoding/json"
	"log/slog"
	"os"

	"github.com/crissyfield/powerhouse/internal/powerhouse"
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
	// Create client
	c, err := powerhouse.New()
	if err != nil {
		slog.Error("Unable to create client", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	// Read list of devices
	devs, err := c.Devices()
	if err != nil {
		slog.Error("Unable to read list of devices", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	// Dump
	_ = json.NewEncoder(os.Stdout).Encode(devs)

	slog.Info("Done")
}
