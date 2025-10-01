package mixmaster

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type mpirsData struct {
	Back      int `yaml:"back"`
	PlayPause int `yaml:"playpause"`
	Next      int `yaml:"next"`
}

type Config struct {
	AppVolumeControls    map[string]int       `yaml:"pulse-app-slidders,omitempty"`
	MasterVolumeControls map[string]int       `yaml:"pulse-master-slidders,omitempty"`
	AppMediaControls     map[string]mpirsData `yaml:"app-media-control,omitempty"`

	SlidderInvert  bool   `yaml:"invert_slidders,omitempty"`
	COMPort        string `yaml:"com_port,omitempty"`
	BaudRate       int    `yaml:"baud_rate,omitempty"`
	NoiseReduction string `yaml:"noise_reduction,omitempty"`
	VID            uint16 `yaml:"usb_vid,omitempty"`
	PID            uint16 `yaml:"usb_pid,omitempty"`
}

func ParseConfig(path string) *Config {
	yamlData, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("error reading file: %v", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(yamlData, &cfg); err != nil {
		log.Fatalf("error parsing yaml: %v", err)
	}

	for name, data := range cfg.AppVolumeControls {
		cfg.AppVolumeControls[name] = data - 1
	}
	for name, data := range cfg.AppMediaControls {
		data.Next--
		data.Back--
		data.PlayPause--
		cfg.AppMediaControls[name] = data
	}
	for name, data := range cfg.MasterVolumeControls {
		data--
		cfg.MasterVolumeControls[name] = data
	}

	return &cfg
}
