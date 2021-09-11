package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Addr              string `yaml:"address"`
	DSN               string `yaml:"dsn"`
	ReadTimeout       int    `yaml:"read_timeout"`
	WriteTimeout      int    `yaml:"write_timeout"`
	ReadHeaderTimeout int    `yaml:"read_header_timeout"`
}

func GetConfig(path string) (Config, error) {
	config := Config{}
	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	err = yaml.NewDecoder(file).Decode(&config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}
