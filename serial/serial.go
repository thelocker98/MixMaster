package serial

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"gitea.locker98.com/locker98/Mixmaster/config"
	"github.com/tarm/serial"
)

type DeviceData struct {
	Volume []float32
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
	buf := make([]byte, 32)

	for {
		n, err := s.Read(buf)
		if err != nil {
			continue
		}
		if n > 0 {
			// Remove newline / carriage return
			values := parseVolumes(buf[:n])
			for i, val := range values {
				if cfg.SliderInvert {
					val = 1 - val
				}
				values[i] = roundFloat32(val, 2)
			}
			out <- &DeviceData{Volume: values}
		}
	}
}

func roundFloat32(f float32, decimals int) float32 {
	pow := math.Pow10(decimals)
	return float32(math.Round(float64(f)*pow) / pow)
}

func parseVolumes(buf []byte) []float32 {
	input := strings.TrimSpace(string(buf))
	parts := strings.Split(input, "|")
	// Convert to ints
	values := make([]float32, 0, len(parts))
	for _, p := range parts {
		num, err := strconv.Atoi(p)
		if err != nil {
			fmt.Println("parse error:", err)
			continue
		}
		values = append(values, float32(float32(num)/float32(1023)))
	}
	return values
}
