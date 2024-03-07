package idevice

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	giDevice "github.com/electricbubble/gidevice"
	"github.com/electricbubble/gidevice/pkg/libimobiledevice"
)

// Device ...
type Device struct {
	// Underlying device
	dev giDevice.Device

	// Function to return internal lockdown client
	internalLockdownClientFn func() (*LockdownClient, error)

	// Function to read pair record
	readPairRecordFn func() (*giDevice.PairRecord, error)

	// Function to return iOS version
	iOSVersionFn func() ([]int, error)
}

// newDevice ...
func newDevice(device giDevice.Device) *Device {
	// Return new device
	dev := &Device{dev: device}

	dev.internalLockdownClientFn = dev.internalLockdownClientFnOnce()
	dev.readPairRecordFn = dev.readPairRecordFnOnce()
	dev.iOSVersionFn = dev.iOSVersionFnOnce()

	return dev
}

// ConnectionType ...
func (dev *Device) ConnectionType() string {
	// ...
	return dev.dev.Properties().ConnectionType
}

// Info ...
func (dev *Device) Info() (any, error) {
	// Get internal lockdown client
	ldc, innerErr := dev.internalLockdownClientFn()
	if innerErr != nil {
		return nil, fmt.Errorf("get internal lockdown client: %w", innerErr)
	}

	// ...
	return ldc.Info()
}

// IOSVersion ...
func (dev *Device) IOSVersion() ([]int, error) {
	return dev.iOSVersionFn()
}

// internalLockdownClientFnOnce ...
func (dev *Device) internalLockdownClientFnOnce() func() (*LockdownClient, error) {
	var ldc *LockdownClient
	var err error
	var once sync.Once

	// Return function that creates internal lockdown client once
	return func() (*LockdownClient, error) {
		once.Do(func() {
			// Create lockdown client
			innerLDC, innerErr := NewLockdownClient(dev)
			if innerErr != nil {
				err = fmt.Errorf("create lockdown client: %w", innerErr)
				return
			}

			ldc = innerLDC
		})

		return ldc, err
	}
}

// pairRecordFnOnce ...
func (dev *Device) readPairRecordFnOnce() func() (*giDevice.PairRecord, error) {
	var pairRecord *libimobiledevice.PairRecord
	var err error
	var once sync.Once

	// Return function that reads pair record once
	return func() (*giDevice.PairRecord, error) {
		once.Do(func() {
			// Read pair record
			innerPairRecord, innerErr := dev.dev.ReadPairRecord()
			if innerErr != nil {
				err = fmt.Errorf("read pair record: %w", innerErr)
				return
			}

			pairRecord = innerPairRecord
		})

		return pairRecord, err
	}
}

// iOSVersionFnOnce ...
func (dev *Device) iOSVersionFnOnce() func() ([]int, error) {
	var iOSVersion []int
	var err error
	var once sync.Once

	// Return function that extracts the iOS version once
	return func() ([]int, error) {
		once.Do(func() {
			// Get internal lockdown client
			ldc, innerErr := dev.internalLockdownClientFn()
			if innerErr != nil {
				err = fmt.Errorf("get internal lockdown client: %w", innerErr)
				return
			}

			// Get lockdown product version
			var value libimobiledevice.LockdownValueResponse

			innerErr = ldc.send(
				&libimobiledevice.LockdownValueRequest{
					LockdownBasicRequest: libimobiledevice.LockdownBasicRequest{
						Label:           libimobiledevice.BundleID,
						ProtocolVersion: libimobiledevice.ProtocolVersion,
						Request:         libimobiledevice.RequestTypeGetValue,
					},
					Domain: "",
					Key:    "ProductVersion",
				},
				&value,
			)

			if innerErr != nil {
				err = fmt.Errorf("get lockdown product version: %w", innerErr)
				return
			}

			productVersion, ok := value.Value.(string)
			if !ok {
				err = fmt.Errorf("unexpected product version: %v", value.Value)
				return
			}

			// Extract product version
			pv := strings.Split(productVersion, ".")

			iOSVersion = make([]int, len(pv))
			for i, v := range pv {
				iOSVersion[i], _ = strconv.Atoi(v)
			}
		})

		return iOSVersion, err
	}
}
