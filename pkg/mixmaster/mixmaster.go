package mixmaster

import (
	"fmt"
	"log"
)

var deviceData *DeviceData

func NewMixMaster(configFile *string) {

	// Hash of Volume and Button slice to detect change from device
	var volumeHash string
	var buttonHash string

	// Load Config
	cfg := ParseConfig(*configFile)

	// Create Pulse Client
	client, err := CreatePulseClient("MixMaster")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Set up mpris
	mpris, err := MprisInitialize()

	// Create channel to receive data
	dataChan := make(chan *DeviceData)

	// Set up Device
	if cfg.PID != 0 && cfg.VID != 0 && cfg.COMPort == "" && cfg.BaudRate == 0 {
		fmt.Println("HID Mode")
		d, err := InitializeConnectionHID(cfg)
		if err != nil {
			log.Fatalf("Error Initializing Device: %s", err)
		}
		// Start Background program
		go ReadDeviceDataHID(d, cfg, dataChan)
	} else {
		fmt.Println("Serial Mode")
		d, err := InitializeConnection(cfg)
		if err != nil {
			log.Fatalf("Error Initializing Device: %s", err)
		}
		// Start Background program
		go ReadDeviceData(d, cfg, dataChan)
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

		if hash, _ := HashSlice(deviceData.Volume); hash != volumeHash {
			volumeHash, _ = HashSlice(deviceData.Volume)
			sessions.ChangeAppVolume(cfg, deviceData.Volume, client)
			sessions.ChangeMasterVolume(cfg, deviceData.Volume, client)
		}

		if hash, _ := HashSlice(deviceData.Button); hash != buttonHash {
			buttonHash, _ = HashSlice(deviceData.Button)
			players.PausePlay(cfg, deviceData)
		}
	}
}
