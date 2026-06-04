package config

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

var (
	ErrInternal   = errors.New("can not load config")
	ErrMarshaling = errors.New("can not marshal config path")
)

func MustLoad() *Config {
	cfg, err := loadConfig()
	if err != nil {
		panic(err)
	}

	return cfg
}

func loadConfig() (*Config, error) {
	var configPath string

	flag.StringVar(&configPath, "config_path", configPath, "Path to config file")
	flag.Parse()

	if configPath == "" {
		if path := os.Getenv("CONFIG_PATH"); path != "" {
			configPath = path
		} else {
			configPath = ConfigBasePath
		}
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no config file found")
	}

	var localConfig LocalDeploy

	if err := cleanenv.ReadConfig(configPath, &localConfig); err != nil {
		return nil, fmt.Errorf("error reading local config: %w", err)
	}

	for _, env := range localConfig.Deploy.Env {
		if err := os.Setenv(env.Name, env.Value); err != nil {
			return nil, fmt.Errorf("error setting env %s: %w", env.Name, err)
		}
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
