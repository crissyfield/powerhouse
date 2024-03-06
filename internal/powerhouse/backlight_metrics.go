package powerhouse

import (
	"fmt"

	"github.com/mitchellh/mapstructure"

	"github.com/crissyfield/powerhouse/internal/idevice"
)

// BacklightMetrics ...
type BacklightMetrics struct {
	RawBrightnessMin   uint64
	RawBrightnessMax   uint64
	RawBrightnessValue uint64
	BrightnessMin      uint64
	BrightnessMax      uint64
	BrightnessValue    uint64
}

// backlightMetricsFromDiagnosticRelayClient reads a BacklightMetrics object from the device.
func backlightMetricsFromDiagnosticRelayClient(drc *idevice.DiagnosticRelayClient) (*BacklightMetrics, error) {
	// Read info from device
	res, err := drc.ReadIORegistry("AppleARMBacklight", "")
	if err != nil {
		return nil, fmt.Errorf("read info from device: %w", err)
	}

	// Parse info
	var backlight struct {
		IODisplayParameters struct {
			RawBrightness struct {
				Min   uint64 `mapstructure:"min"`
				Max   uint64 `mapstructure:"max"`
				Value uint64 `mapstructure:"value"`
			} `mapstructure:"rawBrightness"`
			Brightness struct {
				Min   uint64 `mapstructure:"min"`
				Max   uint64 `mapstructure:"max"`
				Value uint64 `mapstructure:"value"`
			} `mapstructure:"brightness"`
		} `mapstructure:"IODisplayParameters"`
	}

	err = mapstructure.Decode(res, &backlight)
	if err != nil {
		return nil, fmt.Errorf("parse info: %w", err)
	}

	// Send
	return &BacklightMetrics{
		RawBrightnessMin:   backlight.IODisplayParameters.RawBrightness.Min,
		RawBrightnessMax:   backlight.IODisplayParameters.RawBrightness.Max,
		RawBrightnessValue: backlight.IODisplayParameters.RawBrightness.Value,
		BrightnessMin:      backlight.IODisplayParameters.Brightness.Min,
		BrightnessMax:      backlight.IODisplayParameters.Brightness.Max,
		BrightnessValue:    backlight.IODisplayParameters.Brightness.Value,
	}, nil
}
