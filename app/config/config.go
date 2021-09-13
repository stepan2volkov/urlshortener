package config

import (
	"os"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Addr              string `yaml:"addr"`
	DSN               string `yaml:"dsn" envconfig:"DSN" default:"memory" required:"true"`
	ReadTimeout       int    `yaml:"read_timeout" envconfig:"READ_TIMEOUT" default:"30" required:"true"`
	WriteTimeout      int    `yaml:"write_timeout" envconfig:"WRITE_TIMEOUT" default:"30" required:"true"`
	ReadHeaderTimeout int    `yaml:"read_header_timeout" envconfig:"READ_HEADER_TIMEOUT" default:"30" required:"true"`
}

// GetConfig gets path to yaml-file. If path is an empty string,
// configuration will be obtained from environment variables
func GetConfig(path string) (Config, error) {
	if path == "" {
		return getConfigFromEnv()
	}
	return getConfigFromYaml(path)
}

func getConfigFromYaml(path string) (Config, error) {
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

func getConfigFromEnv() (Config, error) {
	config := Config{}

	err := envconfig.Process("", &config)
	if err != nil {
		return Config{}, err
	}

	config.Addr = ":" + os.Getenv("PORT")
	return config, nil
}
