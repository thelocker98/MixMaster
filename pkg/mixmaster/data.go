package mixmaster

type Action int

const (
	PausePlay Action = iota + 1 // starts at 1
	Next
	Previous
)

// type PulseApps struct {
// 	AppName string
// 	Volume  float32
// }

type PulseApps map[string]float32

type MpirsApps struct {
	AppName   string
	PausePlay bool
	Next      bool
	Back      bool
}

type ParsedAudioData struct {
	PulseApps    PulseApps
	MpirsApps    []MpirsApps
	MasterOuputs PulseApps
}

func (data *DeviceData) parseDataConfig(cfg *Config) (ParsedAudioData, error) {
	var parsedData ParsedAudioData
	parsedData.PulseApps = make(map[string]float32)
	parsedData.MasterOuputs = make(map[string]float32)

	for pulseAppName, cfgData := range cfg.AppSlidderMapping {
		if vol, err := GetArrayAt(data.Volume, cfgData.Slidder); err == nil {
			parsedData.PulseApps[pulseAppName] = vol
		}

		if cfgData.MpirsAppName != "" {
			var mpirsData MpirsApps

			mpirsData.AppName = cfgData.MpirsAppName
			mpirsData.PausePlay = data.Button[cfgData.PlayPause]
			mpirsData.Next = data.Button[cfgData.Next]
			mpirsData.Back = data.Button[cfgData.Back]

			parsedData.MpirsApps = append(parsedData.MpirsApps, mpirsData)
		}
	}
	for pulseOutputName, sliderNum := range cfg.MasterSlidderMapping {
		if vol, err := GetArrayAt(data.Volume, sliderNum); err == nil {
			parsedData.MasterOuputs[pulseOutputName] = vol
		}
	}
	return parsedData, nil
}
