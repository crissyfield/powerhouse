package idevice

import (
	"fmt"
	"sync"

	giDevice "github.com/electricbubble/gidevice"
)

// Client ...
type Client struct {
	// Underlying USBMUX
	usbmux giDevice.Usbmux

	// Function to return devices
	devicesFn func() ([]*Device, error)
}

// NewClient ...
func NewClient() (*Client, error) {
	// Create USBMUX connection
	usbmux, err := giDevice.NewUsbmux()
	if err != nil {
		return nil, fmt.Errorf("create USBMUX connection: %w", err)
	}

	// Return new client
	c := &Client{usbmux: usbmux}

	c.devicesFn = c.devicesFnOnce()

	return c, nil
}

// Devices ...
func (c *Client) Devices() ([]*Device, error) {
	return c.devicesFn()
}

// devicesFnOnce ...
func (c *Client) devicesFnOnce() func() ([]*Device, error) {
	var devices []*Device
	var err error
	var once sync.Once

	// return function that created devices once
	return func() ([]*Device, error) {
		once.Do(func() {
			// Get all devices
			innerDevices, innerErr := c.usbmux.Devices()
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
