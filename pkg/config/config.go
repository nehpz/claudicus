package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DevCommand *string `yaml:"devCommand"`
	PortRange  *string `yaml:"portRange"`
}

func DefaultConfig() Config {
	return Config{
		DevCommand: nil,
		PortRange:  nil,
	}
}

// LoadConfig loads the configuration from the specified path
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// GetDefaultConfigPath returns the default path for the config file
func GetDefaultConfigPath() string {
	return "uzi.yaml"
}
