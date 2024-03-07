package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	LogLevel     string       `yaml:"log_level"`
	HttpConfig HttpConfig `yaml:"http"`
}

func LoadConfig(path string) (*Config, error) {
	var cfg Config
	loadEnv(&cfg)
	if path == "" {
		return &cfg, nil
	}
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(file, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
