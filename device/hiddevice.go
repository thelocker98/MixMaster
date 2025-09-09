package device

import (
	"gitea.locker98.com/locker98/Mixmaster/config"
	hid "github.com/sstallion/go-hid"
)

func InitializeConnectionHID(cfg *config.Config) (*hid.Device, error) {
	// Initialize the hid package.
	if err := hid.Init(); err != nil {
		return nil, err
	}

	// Open the device using the VID and PID.
	d, err := hid.OpenFirst(cfg.VID, cfg.PID)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func ReadDeviceDataHID(d *hid.Device, cfg *config.Config, out chan<- *DeviceData) {
	// Buffers for read/write
	in := make([]byte, 64)
	var clean []byte

	for {
		// Try to read a packet
		clean = []byte{}
		for {
			// Read requested state.
			if _, err := d.Read(in); err != nil {
				break
			}
			// Remove trailing zeros
			for _, b := range in {
				if b == 0 {
					break
				}
				clean = append(clean, b)
			}
			// Convert to string
			if clean[0] != '{' {
				clean = []byte{}
				break
			} else if clean[len(clean)-1] == '}' {
				break
			}
		}

		values := parseDeviceData(clean, cfg.SliderInvert)

		out <- values
	}
}
