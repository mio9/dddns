package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"mio9/dddns/internal/config"
)

const minimalProviderConfig = `
providers:
  - type: cloudflare
    zone_id: "zone-id"
    api_token: "token"
    records:
      - name: "home.example.com"
`

func TestLoadUpdateIntervalValid(t *testing.T) {
	tests := []struct {
		value string
		want  time.Duration
	}{
		{value: "30s", want: 30 * time.Second},
		{value: "5m", want: 5 * time.Minute},
		{value: "1h", want: time.Hour},
		{value: "1h30m", want: 90 * time.Minute},
	}

	for _, test := range tests {
		configPath := writeConfig(t, "update-interval: "+test.value+"\n"+minimalProviderConfig)
		cfg, err := config.Load(configPath)
		if err != nil {
			t.Fatalf("Load(%q): %v", test.value, err)
		}
		if cfg.UpdateInterval != test.want {
			t.Fatalf("Load(%q).UpdateInterval = %v, want %v", test.value, cfg.UpdateInterval, test.want)
		}
	}
}

func TestLoadUpdateIntervalOmitted(t *testing.T) {
	configPath := writeConfig(t, minimalProviderConfig)
	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.UpdateInterval != 0 {
		t.Fatalf("UpdateInterval = %v, want zero for one-shot mode", cfg.UpdateInterval)
	}
}

func TestLoadUpdateIntervalInvalid(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{name: "zero", value: "0s"},
		{name: "negative", value: "-5m"},
		{name: "bad format", value: "not-a-duration"},
	}

	for _, test := range tests {
		configPath := writeConfig(t, "update-interval: "+test.value+"\n"+minimalProviderConfig)
		_, err := config.Load(configPath)
		if err == nil {
			t.Fatalf("%s: expected error for update-interval %q", test.name, test.value)
		}
	}
}

func writeConfig(t *testing.T, contents string) string {
	t.Helper()

	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(configPath, []byte(contents), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return configPath
}
