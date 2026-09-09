package config

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

func parseFileConfig(data []byte) (*fileConfig, *Config, error) {
	var fileCfg fileConfig
	if err := yaml.Unmarshal(data, &fileCfg); err != nil {
		return nil, nil, fmt.Errorf("parse config: %w", err)
	}

	cfg := &Config{
		IPProvider: fileCfg.IPProvider,
		IPCache:    fileCfg.IPCache,
		Providers:  fileCfg.Providers,
	}
	if fileCfg.UpdateInterval != "" {
		interval, err := time.ParseDuration(fileCfg.UpdateInterval)
		if err != nil {
			return nil, nil, fmt.Errorf("parse update-interval: %w", err)
		}
		if interval <= 0 {
			return nil, nil, fmt.Errorf("update-interval must be greater than zero")
		}
		cfg.UpdateInterval = interval
	}

	return &fileCfg, cfg, nil
}
