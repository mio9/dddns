package main

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
	IP struct {
		URL string `yaml:"url"`
	} `yaml:"ip"`
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	// Config validation

	if config.Cloudflare.ZoneID == "" {
		return nil, fmt.Errorf("cloudflare.zone_id is required")
	}
	if config.Record.ID == "" && config.Record.Name == "" {
		return nil, fmt.Errorf("record.id or record.name is required")
	}
	if config.Record.Type == "" {
		config.Record.Type = "A"
	}
	if config.Cloudflare.APIToken == "" {
		config.Cloudflare.APIToken = os.Getenv("CLOUDFLARE_API_TOKEN")
	}
	if config.Cloudflare.APIToken == "" {
		return nil, fmt.Errorf("cloudflare.api_token or CLOUDFLARE_API_TOKEN env var is required")
	}
	if config.IP.URL == "" {
		config.IP.URL = defaultIPCheckURL
	}

	return &config, nil
}
