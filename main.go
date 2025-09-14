package main

import (
	"flag"
	"fmt"
	"log"

	"gitea.locker98.com/locker98/Mixmaster/audio"
	"gitea.locker98.com/locker98/Mixmaster/config"
	"gitea.locker98.com/locker98/Mixmaster/device"
	"gitea.locker98.com/locker98/Mixmaster/pulse"
)

var deviceData *device.DeviceData

var (
	configFile   = flag.String("config", "config.yaml", "config.yaml file to custom Mix Master")
	showSessions = flag.Bool("show-sessions", false, "Enable debug mode")
)

// Hash of Volume and Button slice to detect change from device
var volumeHash string
var buttonHash string

func main() {
	// Parse Flags
	flag.Parse()

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

	// Check if the show-sessions flag is active
	if *showSessions {
		fmt.Println("Sessions")
		return
	}

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
