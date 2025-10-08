package mixmaster

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// this struct stores the indexs of the buttons that control mpirs app features
type mpirsData struct {
	Back      int `yaml:"back"`
	PlayPause int `yaml:"playpause"`
	Next      int `yaml:"next"`
}

// This struct stores the main configuration data for each device that is connected to mixmaster
type DeviceConfig struct {
	AppVolumeControls    map[string]int       `yaml:"pulse-app-slidders,omitempty"`
	MasterVolumeControls map[string]int       `yaml:"pulse-master-slidders,omitempty"`
	AppMediaControls     map[string]mpirsData `yaml:"app-media-control,omitempty"`

	SlidderInvert bool  `yaml:"invert_slidders,omitempty"`
	SerialNumber  int64 `yaml:"serialNumber"`
}

// This struct stores all the individual devices and groups togeather all the previous structs into a map
type Config struct {
	Devices map[string]DeviceConfig `yaml:"Devices"`
}

// This function parses the yaml config and puts it in to a struct that the application can understand
func ParseConfig(path string) *Config {
	// open the configuration yaml file
	yamlData, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("error reading file: %v", err)
	}

	// create the configuration struct
	var cfg Config
	// parse the yaml file
	if err := yaml.Unmarshal(yamlData, &cfg); err != nil {
		log.Fatalf("error parsing yaml: %v", err)
	}

	// go through and minus 1 from all the values. this allows the software to ignore any filds that are not populated
	// because a zero is equivilate to a nil when the parse is look through the yaml and it gives it a value of -1 which is ignored
	for deviceName, val := range cfg.Devices {
		for name, data := range val.AppVolumeControls {
			val.AppVolumeControls[name] = data - 1
		}
		for name, data := range val.AppMediaControls {
			data.Next--
			data.Back--
			data.PlayPause--
			val.AppMediaControls[name] = data
		}
		for name, data := range val.MasterVolumeControls {
			data--
			val.MasterVolumeControls[name] = data
		}
		cfg.Devices[deviceName] = val
	}

	return &cfg
}
