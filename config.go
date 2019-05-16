package main

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Bfd     ServerConfig      `yaml:"bfd"`
	Gobgp   ServerConfig      `yaml:"gobgp"`
	Peers   map[string]string `yaml:"peers"`
	Logging LoggingConfig     `yaml:"logging"`
}

type ServerConfig struct {
	Host string    `yaml:"host"`
	Tls  TlsConfig `yaml:"tls,omitempty"`
}

type TlsConfig struct {
	Enable   bool   `yaml:"enable"`
	CertFile string `yaml:"cert"`
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
