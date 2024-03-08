# Powerhouse

Command line tool to measure the power consumption of iDevices.

## Device Setup

To allow for reliable and reproducible power measurements, we recommend setting the device as follows:

#### Update to the latest iOS version

It's best to update to the latest version of iOS or iPadOS to take advantage of Apple's latest power-saving
features.

On your device, go to `Settings` &rarr; `General` &rarr; `Software Update` and install available updates before
proceeding.

#### Disable Auto-Lock

The display has a noticeable effect on the power consumption of the device. This means that measurements may
become unreliable if the display turns off at different times during separate runs. Therefore, it's recommended
that you disable Auto-Lock, which prevents the display from turning off uncontrollably.

On your device, go to `Settings` &rarr; `Display & Brightness` &rarr; `Auto-Lock` and select `Never`.

#### Disable Auto-Brightness

By default, iOS and iPadOS adapt the actual display brightness based on your physical environment. Bright
environments make the display brighter and consume more power, while dark environments make the display dim and
consume less power. It's a good idea to turn off Auto-Brightness so that the display brightness doesn't
fluctuate.

On your device, go to `Settings` &rarr; `Accessibility` &rarr; `Display & Text Size` and turn option
`Auto-Brightness` off.

#### Enable WiFi syncing

While power metrics collection works fine over USB, plugging the device into USB usually starts the charging
process, which can interfere with reliable power metrics. It's recommended to enable WiFi sync, which also
enables wireless power metrics collection.

Connect the device to your Mac via USB and permit bi-directional access. Select the device in Finder and
check the `Show this iDevice when on WiFi` option (where "iDevice" could be "iPhone" or "iPad"). Click the
`Sync` button to activate the new settings.
