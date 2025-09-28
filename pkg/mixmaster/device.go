package mixmaster

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"time"

	hid "github.com/sstallion/go-hid"
	"github.com/tarm/serial"
)

type Device struct {
	HidDev    *string
	SerialDev *serial.Config
}

type DeviceData struct {
	Id     int64
	Volume []float32
	Button []bool
	err    error
}
type arduinoMsg struct {
	Id       int64 `json:"id"`
	Slidders []int `json:"s"`
	Buttons  []int `json:"b"`
}

func roundFloat32(f float32, decimals int) float32 {
	pow := math.Pow10(decimals)
	return float32(math.Round(float64(f)*pow) / pow)
}
func strPtr(s string) *string { return &s }

func parseDeviceData(buf []byte, invertSliders bool) *DeviceData {

	var msg arduinoMsg
	err := json.Unmarshal(buf, &msg)
	if err != nil {
		return &DeviceData{
			Id:     0,
			Volume: nil,
			Button: nil,
			err:    err,
		}
	}

	// Convert to normalized float values (0–1)
	volumes := make([]float32, len(msg.Slidders))
	for i, v := range msg.Slidders {
		if invertSliders {
			v = 255 - v
		}
		volumes[i] = roundFloat32(float32(v)/255.0, 2) // since Arduino sends 0–255
	}

	buttons := make([]bool, len(msg.Buttons))
	for i, btn := range msg.Buttons {
		buttons[i] = (btn == 1)
	}

	return &DeviceData{
		Id:     msg.Id,
		Volume: volumes,
		Button: buttons,
		err:    nil,
	}
}

func GetArrayAt[T any](array []T, index int) (T, error) {
	var zero T
	if index < 0 || index >= len(array) {
		return zero, fmt.Errorf("index %d out of range [0,%d)", index, len(array))
	}
	return array[index], nil
}

func HashSlice[T any](slice []T) (string, error) {
	b, err := json.Marshal(slice)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}

func ListDevices() []string {
	// Initialize the hid package.
	if err := hid.Init(); err != nil {
	}

	var paths []string

	hid.Enumerate(hid.VendorIDAny, hid.ProductIDAny, func(info *hid.DeviceInfo) error {
		fmt.Printf("%s:\n", info.Path)
		paths = append(paths, info.Path)
		return nil
	})

	return paths
}

func GetDevice(id int64) (*Device, error) {
	if err := hid.Init(); err != nil {
	}

	// hid.Enumerate(hid.VendorIDAny, hid.ProductIDAny, func(info *hid.DeviceInfo) error {
	// 	fmt.Printf("%s:\n", info.Path)
	// 	d, _ := hid.OpenPath(info.Path)

	// 	d.Close()
	// 	return nil
	// })

	// return nil, errors.New("No Device Found")
	return &Device{HidDev: strPtr("/dev/hidraw5"), SerialDev: nil}, nil
}

func ReadDeviceDataHID(d *hid.Device, cfg *Config, out chan<- *DeviceData) {
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
		fmt.Println(values)

		if values.err != nil {
			// var err error
			// err = errors.New("Trying to Initialize")
			// for err != nil {
			// 	fmt.Println("searching for new device")
			// 	time.Sleep(1 * time.Second)
			// 	d, err = InitializeConnectionHID(cfg)
			// }

			continue
		}

		out <- values

		time.Sleep(50 * time.Millisecond)
	}
}

func ReadDeviceDataSerial(s *serial.Port, cfg *Config, out chan<- *DeviceData) {
	r := bufio.NewReader(s)

	for {
		data, errReading := r.ReadBytes('\n')
		if errReading != nil {
			// var err error
			// err = errors.New("Trying to Initialize")

			// for err != nil {
			// 	s, err = InitializeConnection(cfg)
			// }
			// r = bufio.NewReader(s)

			continue
		}

		// parse data from the device
		values := parseDeviceData(data, cfg.SlidderInvert)

		out <- values
	}
}
