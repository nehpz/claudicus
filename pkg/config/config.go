package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	PortRange    *string `yaml:"port_range"`
	StartCommand *string `yaml:"start_command"`
	Agents       []Agent `yaml:"agents"`
}

func DefaultConfig() Config {
	return Config{
		PortRange:    nil,
		StartCommand: nil,
		Agents: []Agent{
			{
				Command: "claude",
				Name:    "bob",
			},
		},
	}
}

type Agent struct {
	Command string `yaml:"command"`
	Name    string `yaml:"name"`
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
