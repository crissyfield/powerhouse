package idevice

import (
	"fmt"
	"sync"

	giDevice "github.com/electricbubble/gidevice"
)

// USBMux ...
type USBMux struct {
	// Underlying USBMUX
	usbmux giDevice.Usbmux

	// Function to return devices
	devicesFn func() ([]*Device, error)
}

// NewUSBMux ...
func NewUSBMux() (*USBMux, error) {
	// Create USBMUX connection
	usbmux, err := giDevice.NewUsbmux()
	if err != nil {
		return nil, fmt.Errorf("create USBMux connection: %w", err)
	}

	// Return new client
	c := &USBMux{usbmux: usbmux}

	c.devicesFn = c.devicesFnOnce()

	return c, nil
}

// Devices ...
func (mux *USBMux) Devices() ([]*Device, error) {
	return mux.devicesFn()
}

// devicesFnOnce ...
func (mux *USBMux) devicesFnOnce() func() ([]*Device, error) {
	var devices []*Device
	var err error
	var once sync.Once

	// return function that created devices once
	return func() ([]*Device, error) {
		once.Do(func() {
			// Get all devices
			innerDevices, innerErr := mux.usbmux.Devices()
			if innerErr != nil {
				err = fmt.Errorf("read devices: %w", innerErr)
				return
			}

			// Wrap devices
			devices = make([]*Device, len(innerDevices))

			for i, dev := range innerDevices {
				devices[i] = newDevice(dev)
			}
		})

		return devices, err
	}
}
