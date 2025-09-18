package mixmaster

import (
	"fmt"
	"log"

	"gitea.locker98.com/locker98/Mixmaster/pkg/mixmaster/audio"
	"gitea.locker98.com/locker98/Mixmaster/pkg/mixmaster/config"
	"gitea.locker98.com/locker98/Mixmaster/pkg/mixmaster/device"
	"gitea.locker98.com/locker98/Mixmaster/pkg/mixmaster/pulse"
)

var deviceData *device.DeviceData

func NewMixMaster(configFile *string) {

	// Hash of Volume and Button slice to detect change from device
	var volumeHash string
	var buttonHash string

	// Load Config
	cfg := config.ParseConfig(*configFile)

	// Create Pulse Client
	client, err := pulse.CreatePulseClient("MixMaster")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Set up mpris
	mpris, err := audio.MprisInitialize()

	// Create channel to receive data
	dataChan := make(chan *device.DeviceData)

	// Set up Device
	if cfg.PID != 0 && cfg.VID != 0 && cfg.COMPort == "" && cfg.BaudRate == 0 {
		fmt.Println("HID Mode")
		d, err := device.InitializeConnectionHID(cfg)
		if err != nil {
			log.Fatalf("Error Initializing Device: %s", err)
		}
		// Start Background program
		go device.ReadDeviceDataHID(d, cfg, dataChan)
	} else {
		fmt.Println("Serial Mode")
		d, err := device.InitializeConnection(cfg)
		if err != nil {
			log.Fatalf("Error Initializing Device: %s", err)
		}
		// Start Background program
		go device.ReadDeviceData(d, cfg, dataChan)
	}

	for {
		// Read from channel whenever new data arrives
		deviceData = <-dataChan

		// Get pulse audio sessions
		sessions, err := client.GetAudioSessions()

		// Get mpris sessions
		players, err1 := mpris.ConnectToApps(sessions)

		if err != nil || err1 != nil {
			continue
		}

		if hash, _ := device.HashSlice(deviceData.Volume); hash != volumeHash {
			volumeHash, _ = device.HashSlice(deviceData.Volume)
			sessions.ChangeAppVolume(cfg, deviceData.Volume, client)
			sessions.ChangeMasterVolume(cfg, deviceData.Volume, client)
		}

		if hash, _ := device.HashSlice(deviceData.Button); hash != buttonHash {
			buttonHash, _ = device.HashSlice(deviceData.Button)
			players.PausePlay(cfg, deviceData)
		}
	}
}
