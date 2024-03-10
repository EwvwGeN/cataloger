package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	LogLevel     string     `yaml:"log_level"`
	HttpConfig   HttpConfig `yaml:"http"`
	Validator    Validator  `yaml:"validator"`
}

func LoadConfig(path string) (*Config, error) {
	var (
		cfg Config
		err error
		fileData []byte
	)
	if path != "" {
		fileData, err = os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		err = yaml.Unmarshal(fileData, &cfg)
	} else {
		err = loadEnv(&cfg)
	}
	if err != nil {
		return nil, err
	}
	err = cfg.Validator.mustBeRegex()
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
