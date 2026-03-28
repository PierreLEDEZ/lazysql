package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Connections []Connection `yaml:"connections"`
}

type Connection struct {
	Name     string     `yaml:"name"`
	Driver   string     `yaml:"driver"`
	Host     string     `yaml:"host"`
	Port     int        `yaml:"port"`
	User     string     `yaml:"user"`
	Password string     `yaml:"password"`
	Database string     `yaml:"database"`
	Path     string     `yaml:"path"`
	SSH      *SSHConfig `yaml:"ssh,omitempty"`
}

type SSHConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	KeyPath  string `yaml:"key_path"`
	Password string `yaml:"password"`
}

func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "lazysql", "config.yaml")
}

func Load() (*Config, error) {
	return LoadFrom(DefaultPath())
}

func LoadFrom(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func Save(cfg *Config) error {
	return SaveTo(cfg, DefaultPath())
}

func SaveTo(cfg *Config, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (c *Config) AddConnection(conn Connection) {
	c.Connections = append(c.Connections, conn)
}

func (c *Config) UpdateConnection(index int, conn Connection) error {
	if index < 0 || index >= len(c.Connections) {
		return errors.New("invalid connection index")
	}
	c.Connections[index] = conn
	return nil
}

func (c *Config) RemoveConnection(index int) error {
	if index < 0 || index >= len(c.Connections) {
		return errors.New("invalid connection index")
	}
	c.Connections = append(c.Connections[:index], c.Connections[index+1:]...)
	return nil
}

func (c *Connection) Validate() error {
	switch c.Driver {
	case "postgres", "mysql", "sqlite":
	default:
		return errors.New("driver must be one of: postgres, mysql, sqlite")
	}
	if c.Driver == "sqlite" {
		if c.Path == "" {
			return errors.New("path is required for sqlite driver")
		}
	} else {
		if c.Host == "" {
			return errors.New("host is required for " + c.Driver + " driver")
		}
	}
	return nil
}
