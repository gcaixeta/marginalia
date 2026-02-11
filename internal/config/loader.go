package config

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

func Load() (*Config, error) {
	configPath, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	configFile := filepath.Join(configPath, "marginalia", "config.toml")
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
