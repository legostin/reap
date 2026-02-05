package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	RefreshInterval int               `toml:"refresh_interval"`
	ShowSystem      bool              `toml:"show_system"`
	PortColors      map[string]string `toml:"port_colors"`
	PortLabels      map[string]string `toml:"port_labels"`
}

func Default() Config {
	return Config{
		RefreshInterval: 2,
		ShowSystem:      false,
		PortColors:      map[string]string{},
		PortLabels:      map[string]string{},
	}
}

func Load() Config {
	cfg := Default()

	home, err := os.UserHomeDir()
	if err != nil {
		return cfg
	}

	path := filepath.Join(home, ".config", "reap", "config.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}

	_ = toml.Unmarshal(data, &cfg)

	if cfg.RefreshInterval < 1 {
		cfg.RefreshInterval = 2
	}

	return cfg
}
