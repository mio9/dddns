package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const defaultIPCheckURL = "https://api.ipify.org"

type Config struct {
	Cloudflare struct {
		APIToken string `yaml:"api_token"`
		ZoneID   string `yaml:"zone_id"`
	} `yaml:"cloudflare"`
	Record struct {
		ID   string `yaml:"id"`
		Name string `yaml:"name"`
		Type string `yaml:"type"`
	} `yaml:"record"`
	IPProvider struct {
		URL string `yaml:"url"`
	} `yaml:"ip-provider"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.Cloudflare.ZoneID == "" {
		return nil, fmt.Errorf("cloudflare.zone_id is required")
	}
	if cfg.Record.ID == "" && cfg.Record.Name == "" {
		return nil, fmt.Errorf("record.id or record.name is required")
	}
	if cfg.Record.Type == "" {
		cfg.Record.Type = "A"
	}
	if cfg.Cloudflare.APIToken == "" {
		cfg.Cloudflare.APIToken = os.Getenv("CLOUDFLARE_API_TOKEN")
	}
	if cfg.Cloudflare.APIToken == "" {
		return nil, fmt.Errorf("cloudflare.api_token or CLOUDFLARE_API_TOKEN env var is required")
	}
	if cfg.IPProvider.URL == "" {
		cfg.IPProvider.URL = defaultIPCheckURL
	}

	return &cfg, nil
}
