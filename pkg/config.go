package pkg

import (
	"io"
	"log/slog"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Database DB     `yaml:"database"`
	Server   Server `yaml:"server"`
}

type DB struct {
	Protocol string `yaml:"protocol"`
	Hostname string `yaml:"hostname"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	SSLMode  string `yaml:"sslmode"`
}

type Server struct {
	Port int `yaml:"port"`
}

func (c *Config) LoadFile(file io.Reader) error {
	data, err := io.ReadAll(file)
	if err != nil {
		slog.Error("Failed to read file", "error", err)
		return err
	}

	err = yaml.Unmarshal(data, c)
	if err != nil {
		slog.Error("Failed to unmarshal data", "error", err)
		return err

	}

	return nil
}
