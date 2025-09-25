package mixmaster

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"

	hid "github.com/sstallion/go-hid"
)

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

func GetAt[T any](array []T, index int) (T, error) {
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
