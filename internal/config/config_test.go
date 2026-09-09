package config_test

import (
	"os"
	"path/filepath"
	"strings"
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

func TestLoadOrEmptyMissingFile(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "missing.yaml")

	cfg, fileCfg, err := config.LoadOrEmpty(configPath)
	if err != nil {
		t.Fatalf("LoadOrEmpty: %v", err)
	}
	if cfg.IPProvider.URL != "https://api.ipify.org" {
		t.Fatalf("IPProvider.URL = %q, want default", cfg.IPProvider.URL)
	}
	if len(fileCfg.Providers) != 0 {
		t.Fatalf("fileCfg.Providers = %v, want empty", fileCfg.Providers)
	}
}

func TestLoadOrEmptyExistingFile(t *testing.T) {
	configPath := writeConfig(t, `
update-interval: "10m"
ip-provider:
  url: "https://example.com/ip"
providers:
  - type: cloudflare
    zone_id: "zone"
    api_token: "token"
    records:
      - name: "home.example.com"
`)

	cfg, fileCfg, err := config.LoadOrEmpty(configPath)
	if err != nil {
		t.Fatalf("LoadOrEmpty: %v", err)
	}
	if cfg.IPProvider.URL != "https://example.com/ip" {
		t.Fatalf("IPProvider.URL = %q, want custom URL", cfg.IPProvider.URL)
	}
	if fileCfg.UpdateInterval != "10m" {
		t.Fatalf("UpdateInterval = %q, want 10m", fileCfg.UpdateInterval)
	}
}

func TestMergeProviderRecordsCreateNewFile(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "dddns.yaml")
	records := []config.Record{
		{Name: "home.example.com", Type: "A", ID: "rec-1"},
	}

	err := config.MergeProviderRecords(configPath, config.ProviderCloudflare, config.Provider{
		ZoneID:   "zone-id",
		APIToken: "token",
	}, records)
	if err != nil {
		t.Fatalf("MergeProviderRecords: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	contents := string(data)
	for _, want := range []string{
		"zone_id: zone-id",
		"home.example.com",
		"rec-1",
		"https://api.ipify.org",
	} {
		if !strings.Contains(contents, want) {
			t.Fatalf("config missing %q:\n%s", want, contents)
		}
	}
}

func TestMergeProviderRecordsReplaceRecordsPreserveFields(t *testing.T) {
	configPath := writeConfig(t, `
update-interval: "5m"
ip-provider:
  url: "https://example.com/ip"
providers:
  - type: cloudflare
    zone_id: "zone-id"
    api_token: "token"
    records:
      - name: "old.example.com"
        type: "A"
  - type: no-ip
    zone_name: "example.com"
    api_key: "noip-key"
    records:
      - name: "vpn.example.com"
        type: "A"
`)

	newRecords := []config.Record{
		{Name: "new.example.com", Type: "A", ID: "new-id"},
	}
	err := config.MergeProviderRecords(configPath, config.ProviderCloudflare, config.Provider{
		ZoneID: "zone-id",
	}, newRecords)
	if err != nil {
		t.Fatalf("MergeProviderRecords: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	contents := string(data)
	for _, want := range []string{
		`update-interval: 5m`,
		"https://example.com/ip",
		"new.example.com",
		"new-id",
		"type: no-ip",
		"vpn.example.com",
	} {
		if !strings.Contains(contents, want) {
			t.Fatalf("config missing %q:\n%s", want, contents)
		}
	}
	if strings.Contains(contents, "old.example.com") {
		t.Fatalf("old records should be replaced:\n%s", contents)
	}
}

func TestMergeProviderRecordsAppendProvider(t *testing.T) {
	configPath := writeConfig(t, `
providers:
  - type: cloudflare
    zone_id: "zone-id"
    api_token: "token"
    records:
      - name: "home.example.com"
        type: "A"
`)

	err := config.MergeProviderRecords(configPath, config.ProviderNoIP, config.Provider{
		ZoneName: "example.com",
		APIKey:   "noip-key",
	}, []config.Record{
		{Name: "vpn.example.com", Type: "A"},
	})
	if err != nil {
		t.Fatalf("MergeProviderRecords: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	contents := string(data)
	for _, want := range []string{
		"type: cloudflare",
		"home.example.com",
		"type: no-ip",
		"vpn.example.com",
	} {
		if !strings.Contains(contents, want) {
			t.Fatalf("config missing %q:\n%s", want, contents)
		}
	}
}
