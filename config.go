package main

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	BfdHost   string            `yaml:"bfd-host"`
	GobgpHost string            `yaml:"gobgp-host"`
	Peers     map[string]string `yaml:"peers"`
	Logging   LoggingConfig     `yaml:"logging"`
}

type LoggingConfig struct {
	Logfile     string `yaml:"logfile"`
	LogToStdout bool   `yaml:"log-also-to-stdout"`
}

func LoadConfig(path string) (*Config, error) {
	// Default config file path to $CWD/config.yml
	if len(path) == 0 {
		path = "config.yml"
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	config := &Config{}
	if err := yaml.Unmarshal([]byte(data), config); err != nil {
		return nil, fmt.Errorf("Error parsing config file: %v", err)
	}

	return config, nil
}
