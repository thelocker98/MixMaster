package mixmaster

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"time"

	hid "github.com/sstallion/go-hid"
	"go.bug.st/serial"
)

type Device struct {
	HidDev    map[int64]string
	SerialDev map[int64]string
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

func ListHIDDevices() []string {
	// Initialize the hid package.
	if err := hid.Init(); err != nil {
	}

	var paths []string

	//hid.Enumerate(0x2341, 0x8037, func(info *hid.DeviceInfo) error {
	hid.Enumerate(hid.VendorIDAny, hid.ProductIDAny, func(info *hid.DeviceInfo) error {
		paths = append(paths, info.Path)
		return nil
	})

	return paths
}

func ListSerialDevices() []string {
	ports, err := serial.GetPortsList()

	if err != nil {
		return nil
	}

	return ports
}

func GetDevice() (*Device, error) {
	deviceList := Device{
		HidDev:    make(map[int64]string),
		SerialDev: make(map[int64]string),
	}

	hidDeviceList := ListHIDDevices()
	for _, device := range hidDeviceList {
		d, err := hid.OpenPath(device)

		if err != nil {
			continue
		}

		// Send a 5 to the HID device to trigger a data push
		data := make([]byte, 64)
		data[0] = 5
		_, err = d.Write(data)
		if err != nil {
			// Close connection and end this current loop
			d.Close()
			continue
		}

		// Read data from the current HID device. If device does not respone it 100 milliseconds then timeout
		_, err = d.ReadWithTimeout(data, 100*time.Millisecond)
		if err != nil {
			// Close connection and end this current loop
			d.Close()
			continue
		}
		// Close any Open Devices
		d.Close()

		// Remove trailing zeros from byte array
		var cleanData []byte
		for _, b := range data {
			if b == 0 {
				break
			}
			cleanData = append(cleanData, b)
		}
		deviceList.HidDev[parseDeviceData(cleanData, false).Id] = device
	}

	serialDeviceList := ListSerialDevices()

	for _, port := range serialDeviceList {
		p, err := serial.Open(port, &serial.Mode{BaudRate: 115200})
		if err != nil {
			fmt.Println("error: ", err)
			continue
		}

		buff := make([]byte, 512)

		p.SetReadTimeout(300 * time.Millisecond)
		n, err := p.Read(buff)
		if err != nil || n == 0 {
			p.Close()
			continue
		}

		var cleanData []byte

		for _, value := range buff[:n] {
			cleanData = append(cleanData, value)
			if value == '}' {
				break
			}
		}
		p.Close()

		deviceList.SerialDev[parseDeviceData(cleanData, false).Id] = port
	}

	// No Device Found
	return &deviceList, nil
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
			continue
		}

		// Try to read a packet
		clean = []byte{}
		deadline := time.Now().Add(1000 * time.Millisecond)
		for time.Now().Before(deadline) {
			// Read requested state.
			if _, err := d.Read(in); err != nil {
				fmt.Println("Read Error:")
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
			continue
		}

		out <- values

		time.Sleep(50 * time.Millisecond)
	}
}

func ReadDeviceDataSerial(p serial.Port, cfg *Config, out chan<- *DeviceData) {
	buff := make([]byte, 256)
	for {
		n, err := p.Read(buff)

		if err != nil {
			continue
		}

		// parse data from the device
		values := parseDeviceData(buff[:n], cfg.SlidderInvert)

		out <- values
	}
}
