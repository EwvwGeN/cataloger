package config

import "time"

type HttpConfig struct {
	Host        string        `yaml:"host"`
	Port        string        `yaml:"port"`
	PingTimeout time.Duration `yaml:"ping_timeout"`
}