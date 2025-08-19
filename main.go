package main

import (
	"fmt"

	"gitea.locker98.com/locker98/Mixmaster/config"
	"gitea.locker98.com/locker98/Mixmaster/pulse"
	"gitea.locker98.com/locker98/Mixmaster/serial"
)

var deviceData *serial.DeviceData

func main() {
	cfg := config.ParseConfig("config.yaml")
	fmt.Println(cfg.SliderInvert)

	client, err := pulse.CreatePulseClient("mydeej")

	if err != nil {
		fmt.Println(err)
		return
	}

	// Set up Serial
	p := serial.InitializeConnection(cfg)

	// Create channel to receive data
	dataChan := make(chan *serial.DeviceData)

	go serial.ReadDeviceData(p, cfg, dataChan)

	for {
		// Read from channel whenever new data arrives
		deviceData = <-dataChan
		fmt.Println("Received from task:", deviceData.Volume)

		if len(deviceData.Volume) < 4 {
			fmt.Println("error")
			continue
		}

		for name, num := range cfg.SliderMapping {
			if name == "master" {
				client.ChangeMasterVolume("master", deviceData.Volume[num])
			} else {
				client.ChangeAppVolume(name, deviceData.Volume[num])
			}

		}
	}
}
