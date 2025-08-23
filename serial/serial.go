package serial

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"gitea.locker98.com/locker98/Mixmaster/config"
	"github.com/tarm/serial"
)

type DeviceData struct {
	Volume []float32
	Button []bool
	err    error
}
type arduinoMsg struct {
	Slidders []int `json:"slidders"`
	Buttons  []int `json:"buttons"`
}

func InitializeConnection(cfg *config.Config) *serial.Port {
	// Configure the serial port
	c := &serial.Config{
		Name:        cfg.COMPort,  // your serial port
		Baud:        cfg.BaudRate, // baud rate, adjust as needed
		ReadTimeout: time.Second * 5,
	}

	// Open the port
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatalf("failed to open port: %v", err)
	}
	//defer s.Close()

	return s
}

func ReadDeviceData(s *serial.Port, cfg *config.Config, out chan<- *DeviceData) {
	//buf := make([]byte, 128)
	r := bufio.NewReader(s)

	for {
		data, err := r.ReadBytes('\n')
		if err != nil {
			continue
		}

		// Remove newline / carriage return
		values := parseVolumes(data)
		for i, val := range values.Volume {
			if cfg.SliderInvert {
				val = 1 - val
			}
			values.Volume[i] = roundFloat32(val, 2)
		}
		out <- values
	}
}

func roundFloat32(f float32, decimals int) float32 {
	pow := math.Pow10(decimals)
	return float32(math.Round(float64(f)*pow) / pow)
}

func parseVolumes(buf []byte) *DeviceData {
	// Trim whitespace
	input := string(buf)
	fmt.Println(input)

	var msg arduinoMsg
	err := json.Unmarshal([]byte(input), &msg)
	if err != nil {
		fmt.Println("parse error:", err)
		return &DeviceData{
			Volume: nil,
			Button: nil,
			err:    err,
		}
	}

	// Convert to normalized float values (0–1)
	volumes := make([]float32, len(msg.Slidders))
	for i, v := range msg.Slidders {
		volumes[i] = float32(v) / 255.0 // since Arduino sends 0–255
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
