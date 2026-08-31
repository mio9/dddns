package ipcache_test

import (
	"os"
	"path/filepath"
	"testing"

	"mio9/dddns/internal/ipcache"
)

func TestLoadSaveRoundTrip(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "last.ip")

	ip, err := ipcache.Load(cachePath)
	if err != nil {
		t.Fatalf("Load empty cache: %v", err)
	}
	if ip != "" {
		t.Fatalf("Load empty cache = %q, want empty", ip)
	}

	if err := ipcache.Save(cachePath, "203.0.113.42"); err != nil {
		t.Fatalf("Save: %v", err)
	}

	ip, err = ipcache.Load(cachePath)
	if err != nil {
		t.Fatalf("Load saved cache: %v", err)
	}
	if ip != "203.0.113.42" {
		t.Fatalf("Load saved cache = %q, want 203.0.113.42", ip)
	}
}

func TestPathUsesConfiguredPath(t *testing.T) {
	got := ipcache.Path("/etc/dddns/config.yaml", "/var/lib/dddns/last.ip")
	if got != "/var/lib/dddns/last.ip" {
		t.Fatalf("Path configured = %q", got)
	}
}

func TestPathDefaultsFromConfigPath(t *testing.T) {
	got := ipcache.Path("/etc/dddns/config.yaml", "")
	want := filepath.Join("/etc/dddns", ".config.ip")
	if got != want {
		t.Fatalf("Path default = %q, want %q", got, want)
	}
}

func TestSaveCreatesParentDirectory(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "nested", "last.ip")
	if err := ipcache.Save(cachePath, "203.0.113.10"); err != nil {
		t.Fatalf("Save nested path: %v", err)
	}
	if _, err := os.Stat(cachePath); err != nil {
		t.Fatalf("cache file missing: %v", err)
	}
}
