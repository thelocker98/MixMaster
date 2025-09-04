package main

import (
	"flag"
	"fmt"

	"gitea.locker98.com/locker98/Mixmaster/audio"
	"gitea.locker98.com/locker98/Mixmaster/config"
	"gitea.locker98.com/locker98/Mixmaster/pulse"
	"gitea.locker98.com/locker98/Mixmaster/serial"
)

var deviceData *serial.DeviceData

var (
	configFile   = flag.String("config", "config.yaml", "config.yaml file to custom Mix Master")
	showSessions = flag.Bool("show-sessions", false, "Enable debug mode")
)

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
	dataChan := make(chan *serial.DeviceData)

	// Set up Serial
	p := serial.InitializeConnection(cfg)

	// Start Background program
	go serial.ReadDeviceData(p, cfg, dataChan)
	// Get COunt form config to make sure that their is enough slider and button valuse from the hardware
	sliderCount, buttonCount := findNumberChannelCount(cfg)

	for {
		// Read from channel whenever new data arrives
		deviceData = <-dataChan
		// Get pulse audio sessions
		sessions, err := client.GetAudioSessions()

		// Get mpris sessions
		players, err1 := mpris.ConnectToApps(sessions)

		if err != nil || err1 != nil || len(deviceData.Volume) < sliderCount || len(deviceData.Button) < buttonCount {
			fmt.Print("wrong number of sliders: ")
			continue
		}

		sessions.ChangeAppVolume(cfg, deviceData.Volume, client)
		sessions.ChangeMasterVolume(cfg, deviceData.Volume, client)
		players.PausePlay(cfg, deviceData)
	}
}

func findNumberChannelCount(cfg *config.Config) (int, int) {
	// Count Number of Sliders in Config
	totalSliderCount := 0
	for _, val := range cfg.AppSliderMapping {
		if val > totalSliderCount {
			totalSliderCount = val
		}
	}
	for _, val := range cfg.MasterSliderMapping {
		if val > totalSliderCount {
			totalSliderCount = val
		}
	}

	// Count Number of Buttons in Config
	totalButtonCount := 0
	for _, val := range cfg.AppSliderMapping {
		if val > totalSliderCount {
			totalSliderCount = val
		}
	}
	return totalSliderCount + 1, totalButtonCount + 1
}
