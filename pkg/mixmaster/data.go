package mixmaster

import "fmt"

type MpirsApp struct {
	Back      bool
	PausePlay bool
	Next      bool
}

type PulseApps map[string]float32
type MpirsApps map[string]MpirsApp

type ParsedAudioData struct {
	PulseApps    PulseApps
	MasterOuputs PulseApps
	MpirsApps    MpirsApps
}

func (data *DeviceData) parseDataConfig(cfg *Config) (ParsedAudioData, error) {
	var parsedData ParsedAudioData
	parsedData.PulseApps = make(map[string]float32)
	parsedData.MpirsApps = make(map[string]MpirsApp)
	parsedData.MasterOuputs = make(map[string]float32)

	for pulseAppName, sliderNum := range cfg.AppVolumeControls {
		if vol, err := GetArrayAt(data.Volume, sliderNum); err == nil {
			parsedData.PulseApps[pulseAppName] = vol
		}
	}
	for pulseOutputName, sliderNum := range cfg.MasterVolumeControls {
		if vol, err := GetArrayAt(data.Volume, sliderNum); err == nil {
			parsedData.MasterOuputs[pulseOutputName] = vol
		}
	}
	for mpirsAppName, mpirsAction := range cfg.AppMediaControls {
		if mpirsAppName != "" {
			var tempAppData MpirsApp
			if playpause, err := GetArrayAt(data.Button, mpirsAction.PlayPause); err == nil {
				tempAppData.PausePlay = playpause
			}
			if next, err := GetArrayAt(data.Button, mpirsAction.Next); err == nil {
				tempAppData.Next = next
			}
			if back, err := GetArrayAt(data.Button, mpirsAction.Back); err == nil {
				tempAppData.Back = back
			}

			parsedData.MpirsApps[mpirsAppName] = tempAppData
		}
	}

	fmt.Println(parsedData)
	return parsedData, nil
}
