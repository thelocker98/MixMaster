package mixmaster

import (
	"fmt"
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

	SlidderInvert bool   `yaml:"invert_slidders,omitempty"`
	SerialNumber  string `yaml:"serialNumber"`
}
type AppConfig struct {
	Theme            string `yaml:"theme"`
	LaunchGUIOnStart bool   `yaml:"launch_gui_on_start"`
	Notifications    bool   `yaml:"notifications"`
}

// This struct stores all the individual devices and groups togeather all the previous structs into a map
type Config struct {
	Devices map[string]DeviceConfig `yaml:"Devices"`
	App     AppConfig
}

// This function parses the yaml config and puts it in to a struct that the application can understand
func ParseConfig(path string) *Config {
	// open the configuration yaml file
	yamlData, err := os.ReadFile(path)
	if err != nil {
		// Check if the file doesn't exist
		if os.IsNotExist(err) {
			// Create an empty file (or with default content)
			defaultContent, err := yaml.Marshal(Config{
				App: AppConfig{
					Notifications:    true,
					LaunchGUIOnStart: true,
				},
			})

			err = os.WriteFile(path, defaultContent, 0644)
			if err != nil {
				log.Fatalf("error creating file: %v", err)
			}
			fmt.Println("Created missing config file:", path)

			// Use the default content as yamlData
			yamlData = defaultContent
		} else {
			log.Fatalf("error reading file: %v", err)
		}
	}

	// create the configuration struct
	var cfg Config

	// parse the yaml file
	err = yaml.Unmarshal(yamlData, &cfg)
	if err != nil {
		log.Fatalf("error parsing yaml: %v", err)
	}
	if cfg.Devices == nil {
		cfg.Devices = make(map[string]DeviceConfig)
	}

	return &cfg
}

func (cfg *Config) SaveConfig(path *string) error {
	// parse config struct into a yaml file
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	// Write to file
	err = os.WriteFile(*path, data, 0644)
	if err != nil {
		return err
	}

	return nil
}
