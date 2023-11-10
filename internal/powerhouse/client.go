package powerhouse

import (
	"fmt"
	"sync"

	giDevice "github.com/electricbubble/gidevice"
)

// Client ...
type Client struct {
	usbmux      giDevice.Usbmux
	devices     []*Device
	devicesErr  error
	devicesOnce sync.Once
}

// NewClient ...
func NewClient() (*Client, error) {
	// Create USBMUX connection
	usbmux, err := giDevice.NewUsbmux()
	if err != nil {
		return nil, fmt.Errorf("create USBMUX connection: %w", err)
	}

	// ...
	return &Client{
		usbmux: usbmux,
	}, nil
}

// Devices ...
func (c *Client) Devices() ([]*Device, error) {
	// ...
	c.devicesOnce.Do(func() {
		// ...
		ds, err := c.usbmux.Devices()
		if err != nil {
			c.devicesErr = fmt.Errorf("read devices: %w", err)
			return
		}

		// ...
		devices := make([]*Device, len(ds))

		for i, d := range ds {
			devices[i], err = newDevice(d)
			if err != nil {
				c.devicesErr = fmt.Errorf("initialize device: %w", err)
				return
			}
		}

		c.devices = devices
	})

	return c.devices, c.devicesErr
}