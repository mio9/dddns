package ipcache

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Path(configPath, configuredPath string) string {
	if configuredPath != "" {
		return configuredPath
	}

	configDir := filepath.Dir(configPath)
	configBase := strings.TrimSuffix(filepath.Base(configPath), filepath.Ext(configPath))
	if configBase == "" {
		configBase = "dddns"
	}
	return filepath.Join(configDir, "."+configBase+".ip")
}

func Load(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read IP cache: %w", err)
	}

	ip := strings.TrimSpace(string(data))
	if ip == "" {
		return "", nil
	}

	return ip, nil
}

func Save(path, ip string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create IP cache directory: %w", err)
	}

	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, []byte(ip), 0o644); err != nil {
		return fmt.Errorf("write IP cache: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("save IP cache: %w", err)
	}

	return nil
}
