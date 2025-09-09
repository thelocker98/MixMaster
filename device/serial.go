package device

import (
	"bufio"
	"time"

	"gitea.locker98.com/locker98/Mixmaster/config"
	"github.com/tarm/serial"
)

func InitializeConnection(cfg *config.Config) (*serial.Port, error) {
	// Configure the serial port
	c := &serial.Config{
		Name:        cfg.COMPort,  // your serial port
		Baud:        cfg.BaudRate, // baud rate, adjust as needed
		ReadTimeout: time.Second * 5,
	}

	// Open the port
	s, err := serial.OpenPort(c)
	if err != nil {
		return nil, err
	}

	return s, err
}

func ReadDeviceData(s *serial.Port, cfg *config.Config, out chan<- *DeviceData) {
	r := bufio.NewReader(s)

	for {
		data, err := r.ReadBytes('\n')
		if err != nil {
			continue
		}

		// parse data from the device
		values := parseDeviceData(data, cfg.SliderInvert)

		out <- values
	}
}
