package mixmaster

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type AppData struct {
	MpirsAppName string `yaml:"appname"`
	Slidder      int    `yaml:"slidder"`
	Back         int    `yaml:"back"`
	Next         int    `yaml:"next"`
	PlayPause    int    `yaml:"playpause"`
}

type Config struct {
	AppSlidderMapping    map[string]AppData `yaml:"app-slidders"`
	MasterSlidderMapping map[string]int     `yaml:"master-slidders,omitempty"`

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

	for name, data := range cfg.AppSlidderMapping {
		data.Next--
		data.Back--
		data.PlayPause--
		data.Slidder--
		cfg.AppSlidderMapping[name] = data
	}
	for name, data := range cfg.MasterSlidderMapping {
		data--
		cfg.MasterSlidderMapping[name] = data
	}

	return &cfg
}
