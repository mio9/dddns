package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type FetchOverrides struct {
	ZoneID   string
	ZoneName string
}

func LoadOrEmpty(path string) (*Config, *fileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := &Config{}
			cfg.IPProvider.URL = defaultIPCheckURL
			return cfg, &fileConfig{}, nil
		}
		return nil, nil, fmt.Errorf("read config: %w", err)
	}

	fileCfg, cfg, err := parseFileConfig(data)
	if err != nil {
		return nil, nil, err
	}
	if cfg.IPProvider.URL == "" {
		cfg.IPProvider.URL = defaultIPCheckURL
	}
	return cfg, fileCfg, nil
}

func LoadFileConfig(path string) (*fileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &fileConfig{}, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	fileCfg, _, err := parseFileConfig(data)
	if err != nil {
		return nil, err
	}
	return fileCfg, nil
}

func ResolveProviderCredentials(providerType string, fileCfg *fileConfig, overrides FetchOverrides) (Provider, error) {
	provider := findProviderByType(fileCfg.Providers, providerType)
	provider.Type = providerType

	switch providerType {
	case ProviderCloudflare:
		if overrides.ZoneID != "" {
			provider.ZoneID = overrides.ZoneID
		}
		if provider.APIToken == "" {
			provider.APIToken = os.Getenv("CLOUDFLARE_API_TOKEN")
		}
		if provider.ZoneID == "" {
			return Provider{}, fmt.Errorf("zone_id is required: set in config, --zone-id, or provider block")
		}
		if provider.APIToken == "" {
			return Provider{}, fmt.Errorf("api_token or CLOUDFLARE_API_TOKEN env var is required")
		}
	case ProviderNoIP:
		if overrides.ZoneName != "" {
			provider.ZoneName = overrides.ZoneName
		}
		if provider.APIKey == "" {
			provider.APIKey = os.Getenv("NOIP_API_KEY")
		}
		if provider.ZoneName == "" {
			return Provider{}, fmt.Errorf("zone_name is required: set in config, --zone-name, or provider block")
		}
		if provider.APIKey == "" {
			return Provider{}, fmt.Errorf("api_key or NOIP_API_KEY env var is required")
		}
	default:
		return Provider{}, fmt.Errorf("unsupported provider type %q", providerType)
	}

	return provider, nil
}

func MergeProviderRecords(path, providerType string, providerFields Provider, records []Record) error {
	fileCfg, err := LoadFileConfig(path)
	if err != nil {
		return err
	}

	if fileCfg.IPProvider.URL == "" {
		fileCfg.IPProvider.URL = defaultIPCheckURL
	}

	providerFields.Type = providerType
	providerFields.Records = records
	original := findProviderByType(fileCfg.Providers, providerType)

	index := findProviderIndexByType(fileCfg.Providers, providerType)
	if index >= 0 {
		existing := fileCfg.Providers[index]
		providerFields.ZoneID = firstNonEmpty(providerFields.ZoneID, existing.ZoneID)
		providerFields.ZoneName = firstNonEmpty(providerFields.ZoneName, existing.ZoneName)
		providerFields.APIToken = firstNonEmpty(providerFields.APIToken, existing.APIToken)
		providerFields.APIKey = firstNonEmpty(providerFields.APIKey, existing.APIKey)
		fileCfg.Providers[index] = providerFields
	} else {
		fileCfg.Providers = append(fileCfg.Providers, providerFields)
	}

	writtenIndex := index
	if writtenIndex < 0 {
		writtenIndex = len(fileCfg.Providers) - 1
	}
	if original.APIToken == "" {
		fileCfg.Providers[writtenIndex].APIToken = ""
	}
	if original.APIKey == "" {
		fileCfg.Providers[writtenIndex].APIKey = ""
	}

	return Save(path, fileCfg)
}

func Save(path string, fileCfg *fileConfig) error {
	data, err := yaml.Marshal(fileCfg)
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}
