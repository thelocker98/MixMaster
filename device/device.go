package device

import (
	"encoding/json"
	"math"
)

type DeviceData struct {
	Volume []float32
	Button []bool
	err    error
}
type arduinoMsg struct {
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
		Volume: volumes,
		Button: buttons,
		err:    nil,
	}
}
