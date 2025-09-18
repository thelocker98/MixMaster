package device

import (
	"errors"
	"fmt"
	"time"

	"gitea.locker98.com/locker98/Mixmaster/pkg/mixmaster/config"
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
		buf := make([]byte, 64) // automatically filled with zeros
		for i := range buf {
			buf[i] = 0
		}
		buf[0] = 5
		if _, err := d.Write(buf); err != nil {
			fmt.Println("Write error:")
		}

		// Try to read a packet
		clean = []byte{}
		deadline := time.Now().Add(1000 * time.Millisecond)
		for time.Now().Before(deadline) {
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

		values := parseDeviceData(clean, cfg.SlidderInvert)
		if values.err != nil {
			var err error
			err = errors.New("Trying to Initialize")
			for err != nil {
				fmt.Println("searching for new device")
				time.Sleep(1 * time.Second)
				d, err = InitializeConnectionHID(cfg)
			}

			continue
		}

		out <- values

		time.Sleep(50 * time.Millisecond)
	}
}
