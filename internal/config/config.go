package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	dirName         = ".wuji"
	driversFileName = "drivers.yaml"
)

// Config holds generic Wuji runtime configuration.
type Config struct {
	Root            string
	DriverEndpoints []string
}

type driversConfig struct {
	Drivers []driverEntry `yaml:"drivers"`
}

type driverEntry struct {
	Endpoint string `yaml:"endpoint"`
}

// Load reads configuration from the project root.
func Load() (*Config, error) {
	root, err := findProjectRoot()
	if err != nil {
		return nil, err
	}

	cfg := &Config{Root: root}

	if v := os.Getenv("WUJI_ROOT"); v != "" {
		cfg.Root = v
	}

	path := filepath.Join(cfg.Root, dirName, driversFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	var df driversConfig
	if err := yaml.Unmarshal(data, &df); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	for _, d := range df.Drivers {
		if d.Endpoint != "" {
			cfg.DriverEndpoints = append(cfg.DriverEndpoints, d.Endpoint)
		}
	}

	return cfg, nil
}

// SaveDriverEndpoint appends an endpoint to the drivers config file.
func SaveDriverEndpoint(root, endpoint string) error {
	dir := filepath.Join(root, dirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	path := filepath.Join(dir, driversFileName)
	var df driversConfig

	if data, err := os.ReadFile(path); err == nil {
		_ = yaml.Unmarshal(data, &df)
	}

	for _, d := range df.Drivers {
		if d.Endpoint == endpoint {
			return nil
		}
	}

	df.Drivers = append(df.Drivers, driverEntry{Endpoint: endpoint})
	data, err := yaml.Marshal(df)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find project root (no go.mod)")
		}
		dir = parent
	}
}
