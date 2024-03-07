package cmd

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/crissyfield/powerhouse/internal/powerhouse"
)

// CmdMeasure defines the CLI sub-command 'list'.
var CmdMeasure = &cobra.Command{
	Use:   "measure [flags]",
	Short: "...",
	Args:  cobra.NoArgs,
	Run:   runMeasure,
}

// Initialize CLI options.
func init() {
	// Measure
	CmdMeasure.Flags().DurationP("duration", "d", 10*time.Minute, "max duration of the measurement")
}

// runMeasure is called when the "test" command is used.
func runMeasure(_ *cobra.Command, _ []string) {
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

	if len(devices) == 0 {
		slog.Warn("No device connected. Exiting")
		os.Exit(1) //nolint
	}

	// Start reporting metrics
	ctx, cancel := context.WithCancel(context.Background())

	metrics, err := devices[0].ReportMetrics(ctx)
	if err != nil {
		slog.Error("Unable to start metrics", slog.Any("error", err))
		os.Exit(1) //nolint
	}

	// Create signal that fires on interrupt
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Create timer that fires an interrupt
	expired := time.NewTimer(viper.GetDuration("duration"))

	// Event loop
loop:
	for {
		select {
		case <-stop:
			// Stop
			slog.Info("Stop requested")
			break loop

		case <-expired.C:
			// Stop
			slog.Info("Time is up")
			break loop

		case m := <-metrics:
			// Handling of potential errors
			if m.Err != nil {
				slog.Error("Unable to report error", slog.Any("error", m.Err))
				os.Exit(1) //nolint
			}

			// Report
			_ = json.NewEncoder(os.Stdout).Encode(m)
		}
	}

	// Canceling the context stops reporting
	cancel()

	for range metrics { //nolint
		// Do something with old metrics
	}
}
