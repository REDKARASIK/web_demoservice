package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

func NewConfig(path string) (*Config, error) {
	_, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if _, err = toml.DecodeFile(path, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
