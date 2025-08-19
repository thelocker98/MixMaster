package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	SliderMapping map[string]int `yaml:"slider_mapping"`

	SliderInvert   bool   `yaml:"invert_sliders"`
	COMPort        string `yaml:"com_port"`
	BaudRate       int    `yaml:"baud_rate"`
	NoiseReduction string `yaml:"noise_reduction"`
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
	return &cfg
}
