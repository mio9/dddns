package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

const defaultIPCheckURL = "https://api.ipify.org"

const ProviderCloudflare = "cloudflare"
const ProviderNoIP = "no-ip"

type Record struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

type Provider struct {
	Type     string   `yaml:"type"`
	ZoneID   string   `yaml:"zone_id"`
	ZoneName string   `yaml:"zone_name"`
	APIToken string   `yaml:"api_token"`
	APIKey   string   `yaml:"api_key"`
	Records  []Record `yaml:"records"`
}

type Config struct {
	UpdateInterval time.Duration
	IPProvider     struct {
		URL string `yaml:"url"`
	} `yaml:"ip-provider"`
	IPCache struct {
		Path string `yaml:"path"`
	} `yaml:"ip-cache"`
	Providers []Provider `yaml:"providers"`
}

type fileConfig struct {
	UpdateInterval string `yaml:"update-interval"`
	IPProvider     struct {
		URL string `yaml:"url"`
	} `yaml:"ip-provider"`
	IPCache struct {
		Path string `yaml:"path"`
	} `yaml:"ip-cache"`
	Providers []Provider `yaml:"providers"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var fileCfg fileConfig
	if err := yaml.Unmarshal(data, &fileCfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	cfg := Config{
		IPProvider: fileCfg.IPProvider,
		IPCache:    fileCfg.IPCache,
		Providers:  fileCfg.Providers,
	}
	if fileCfg.UpdateInterval != "" {
		interval, err := time.ParseDuration(fileCfg.UpdateInterval)
		if err != nil {
			return nil, fmt.Errorf("parse update-interval: %w", err)
		}
		if interval <= 0 {
			return nil, fmt.Errorf("update-interval must be greater than zero")
		}
		cfg.UpdateInterval = interval
	}

	if len(cfg.Providers) == 0 {
		return nil, fmt.Errorf("at least one provider is required")
	}
	for index := range cfg.Providers {
		if err := validateProvider(&cfg.Providers[index], index); err != nil {
			return nil, err
		}
	}
	if cfg.IPProvider.URL == "" {
		cfg.IPProvider.URL = defaultIPCheckURL
	}

	return &cfg, nil
}

func validateProvider(provider *Provider, index int) error {
	if provider.Type == "" {
		return fmt.Errorf("providers[%d]: type is required", index)
	}

	switch provider.Type {
	case ProviderCloudflare:
		return validateCloudflareProvider(provider, index)
	case ProviderNoIP:
		return validateNoIPProvider(provider, index)
	default:
		return fmt.Errorf("providers[%d]: unsupported type %q", index, provider.Type)
	}
}

func validateCloudflareProvider(provider *Provider, index int) error {
	if provider.ZoneID == "" {
		return fmt.Errorf("providers[%d]: zone_id is required", index)
	}
	if provider.APIToken == "" {
		provider.APIToken = os.Getenv("CLOUDFLARE_API_TOKEN")
	}
	if provider.APIToken == "" {
		return fmt.Errorf("providers[%d]: api_token or CLOUDFLARE_API_TOKEN env var is required", index)
	}
	return validateRecords(provider.Records, fmt.Sprintf("providers[%d]", index))
}

func validateNoIPProvider(provider *Provider, index int) error {
	if provider.ZoneName == "" {
		return fmt.Errorf("providers[%d]: zone_name is required", index)
	}
	if provider.APIKey == "" {
		provider.APIKey = os.Getenv("NOIP_API_KEY")
	}
	if provider.APIKey == "" {
		return fmt.Errorf("providers[%d]: api_key or NOIP_API_KEY env var is required", index)
	}
	return validateRecords(provider.Records, fmt.Sprintf("providers[%d]", index))
}

func validateRecords(records []Record, prefix string) error {
	if len(records) == 0 {
		return fmt.Errorf("%s: at least one record is required", prefix)
	}
	for index := range records {
		if records[index].ID == "" && records[index].Name == "" {
			return fmt.Errorf("%s.records[%d]: id or name is required", prefix, index)
		}
		if records[index].Type == "" {
			records[index].Type = "A"
		}
	}
	return nil
}
