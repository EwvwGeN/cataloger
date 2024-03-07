package main

import (
	"flag"
	"fmt"
	"log/slog"

	c "github.com/EwvwGeN/InHouseAd_assignment/internal/config"
	l "github.com/EwvwGeN/InHouseAd_assignment/internal/logger"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config", "", "path to config file")
}
func main() {
	flag.Parse()
	cfg, err := c.LoadConfig(configPath)
	if err != nil {
		panic(fmt.Sprintf("cant load config from path %s: %s", configPath, err.Error()))
	}
	logger := l.SetupLogger(cfg.LogLevel)
	logger.Info("logger is initiated")
	logger.Debug("config data", slog.Any("config", cfg))
}