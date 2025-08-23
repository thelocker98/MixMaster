package main

import (
	"fmt"

	"gitea.locker98.com/locker98/Mixmaster/audio"
	"gitea.locker98.com/locker98/Mixmaster/config"
	"gitea.locker98.com/locker98/Mixmaster/pulse"
	"gitea.locker98.com/locker98/Mixmaster/serial"
)

var deviceData *serial.DeviceData

func main() {
	cfg := config.ParseConfig("myconfig.yaml")

	client, err := pulse.CreatePulseClient("mydeej")

	if err != nil {
		fmt.Println(err)
		return
	}

	// Set up Serial
	p := serial.InitializeConnection(cfg)

	// Set up mpris
	mpris, err := audio.MprisInitialize()

	// Create channel to receive data
	dataChan := make(chan *serial.DeviceData)

	go serial.ReadDeviceData(p, cfg, dataChan)
	

	// find the number of sliders
	maxSliderNum := 0
	for _, val := range cfg.AppSliderMapping {
		if val > maxSliderNum {
			maxSliderNum = val
		}
	}
	for _, val := range cfg.MasterSliderMapping {
		if val > maxSliderNum {
			maxSliderNum = val
		}
	}
	maxSliderNum += 1

	for {
		// Read from channel whenever new data arrives
		deviceData = <-dataChan
		// Get pulse audio sessions
		sessions, err := client.GetAudioSessions()

		// Get mpris sessions
		players, err1 := mpris.ConnectToApps(sessions)

		fmt.Println(players)

		if err != nil || err1 != nil || len(deviceData.Volume) < maxSliderNum {
			fmt.Print("wrong number of sliders: ")
			continue
		}

		sessions.ChangeAppVolume(cfg, deviceData.Volume, client)
		sessions.ChangeMasterVolume(cfg, deviceData.Volume, client)
		players.PausePlay(cfg, deviceData)
	}
}
