package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	defaultInferenceHost = "127.0.0.1"
	defaultInferencePort = 8080
	defaultGRPCAddr      = "127.0.0.1:50052"
)

// Config holds driver-local configuration.
type Config struct {
	Root          string `yaml:"-"`
	ServerBin     string `yaml:"server_bin"`
	ModelsDir     string `yaml:"models_dir"`
	DefaultModel  string `yaml:"default_model"`
	InferenceHost string `yaml:"inference_host"`
	InferencePort int    `yaml:"inference_port"`
	GRPCAddr      string `yaml:"grpc_addr"`
	OllamaAPI     string `yaml:"ollama_api"`
	OllamaThink   bool   `yaml:"ollama_think"`
}

// Load reads config from the driver root directory.
func Load() (*Config, error) {
	root, err := findDriverRoot()
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Root:          root,
		ServerBin:     "vendor/llama/llama-server",
		ModelsDir:     "models",
		InferenceHost: defaultInferenceHost,
		InferencePort: defaultInferencePort,
		GRPCAddr:      defaultGRPCAddr,
		OllamaAPI:     "http://127.0.0.1:11434",
		OllamaThink:   false,
	}

	path := filepath.Join(root, "config.yaml")
	if data, err := os.ReadFile(path); err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
	}

	cfg.Root = root
	cfg.ServerBin = resolvePath(root, cfg.ServerBin)
	cfg.ModelsDir = resolvePath(root, cfg.ModelsDir)

	if v := os.Getenv("WUJI_DRIVER_ROOT"); v != "" {
		cfg.Root = v
		cfg.ServerBin = resolvePath(v, "vendor/llama/llama-server")
		cfg.ModelsDir = resolvePath(v, "models")
	}

	return cfg, nil
}

func (c *Config) ServerAvailable() bool {
	info, err := os.Stat(c.ServerBin)
	return err == nil && !info.IsDir()
}

func (c *Config) ListModels() ([]string, error) {
	var models []string
	err := filepath.WalkDir(c.ModelsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if d.IsDir() {
			return nil
		}
		name := d.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".gguf") {
			return nil
		}
		rel, err := filepath.Rel(c.ModelsDir, path)
		if err != nil {
			return nil
		}
		models = append(models, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return models, nil
}

func (c *Config) DownloadURL() string {
	switch runtime.GOOS + "/" + runtime.GOARCH {
	case "linux/arm64":
		return "https://github.com/ggml-org/llama.cpp/releases/download/b10502/llama-b10502-bin-ubuntu-arm64.tar.gz"
	default:
		return "https://github.com/ggml-org/llama.cpp/releases/download/b10502/llama-b10502-bin-ubuntu-x64.tar.gz"
	}
}

func findDriverRoot() (string, error) {
	if v := os.Getenv("WUJI_DRIVER_ROOT"); v != "" {
		return v, nil
	}

	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "drivers", "llama", "config.yaml")); err == nil {
			return filepath.Join(dir, "drivers", "llama"), nil
		}
		if _, err := os.Stat(filepath.Join(dir, "config.yaml")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, "vendor", "llama")); err == nil {
				return dir, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find drivers/llama directory")
		}
		dir = parent
	}
}

func resolvePath(root, p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(root, p)
}
