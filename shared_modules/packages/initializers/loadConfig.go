package initializers

import (
	"fmt"
	"my_app/shared_modules/interfaces"
	"os"

	"gopkg.in/yaml.v3"
)

func innerLoadConfig(filename string) (*interfaces.Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config interfaces.Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return &config, nil
}

func LoadConfig(filename string) *interfaces.Config {
	config, err := innerLoadConfig("config.yaml")
	if err != nil {
		panic(fmt.Sprintf("Error loading config: %v", err))
	}

	return config
}
