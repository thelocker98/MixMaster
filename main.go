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

		sessions, err := client.GetAudioSessions()

		if err != nil || len(deviceData.Volume) < 4 {
			continue
		}

		sessions.ChangeAppVolume(cfg, deviceData.Volume, client)
		sessions.ChangeMasterVolume(cfg, deviceData.Volume, client)
	}
}
