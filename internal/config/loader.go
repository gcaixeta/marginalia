package config

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

func Default() *Config {
	return &Config{}
}

func Save(cfg *Config) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(configDir, "marginalia")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "config.toml"), data, 0644)
}

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
