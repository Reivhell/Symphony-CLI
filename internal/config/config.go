package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	LocalTemplatePaths []string `yaml:"local_template_paths"`
	CacheTTLHours      int      `yaml:"cache_ttl_hours"`
	HTTPProxy          string   `yaml:"http_proxy"`
	HTTPSProxy         string   `yaml:"https_proxy"`
	GitHubToken        string   `yaml:"github_token"`
	DefaultFormat      string   `yaml:"default_format"`
}

// Default membangun konfigurasi default yang aman
func Default() *Config {
	return &Config{
		LocalTemplatePaths: []string{},
		CacheTTLHours:      24,
		DefaultFormat:      "human",
	}
}

// Load membaca file konfigurasi global di ~/.symphony/config.yaml.
// Bila belum diciptakan, ia kembali membawa konfigurasi default.
func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Default(), fmt.Errorf("gagal mendapatkan user home dir: %w", err)
	}

	configPath := filepath.Join(home, ".symphony", "config.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil // File default safe
		}
		return Default(), fmt.Errorf("gagal membaca file config %s: %w", configPath, err)
	}

	cfg := Default()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return cfg, fmt.Errorf("sintaks YAML dari file config %s keliru: %w", configPath, err)
	}

	return cfg, nil
}
