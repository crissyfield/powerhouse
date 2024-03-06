package cmd

import (
	"encoding/json"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"github.com/crissyfield/powerhouse/internal/powerhouse"
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
	// Create powerhouse
	ph, err := powerhouse.New()
	if err != nil {
		slog.Error("Unable to create powerhouse", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	// Read list of devices
	devices, err := ph.Devices()
	if err != nil {
		slog.Error("Unable to read list of devices", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	// Dump
	_ = json.NewEncoder(os.Stdout).Encode(devices)

	slog.Info("Done")
}
