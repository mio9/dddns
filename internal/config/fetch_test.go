package config

import (
	"testing"
)

func TestResolveProviderCredentialsPrecedence(t *testing.T) {
	t.Setenv("CLOUDFLARE_API_TOKEN", "env-token")
	t.Setenv("NOIP_API_KEY", "env-key")

	fileCfg := &fileConfig{
		Providers: []Provider{
			{
				Type:     ProviderCloudflare,
				ZoneID:   "config-zone",
				APIToken: "config-token",
			},
			{
				Type:     ProviderNoIP,
				ZoneName: "config.example.com",
				APIKey:   "config-key",
			},
		},
	}

	cloudflareProvider, err := ResolveProviderCredentials(
		ProviderCloudflare,
		fileCfg,
		FetchOverrides{ZoneID: "flag-zone"},
	)
	if err != nil {
		t.Fatalf("ResolveProviderCredentials cloudflare: %v", err)
	}
	if cloudflareProvider.ZoneID != "flag-zone" {
		t.Fatalf("ZoneID = %q, want flag override", cloudflareProvider.ZoneID)
	}
	if cloudflareProvider.APIToken != "config-token" {
		t.Fatalf("APIToken = %q, want config value", cloudflareProvider.APIToken)
	}

	noipProvider, err := ResolveProviderCredentials(
		ProviderNoIP,
		fileCfg,
		FetchOverrides{ZoneName: "flag.example.com"},
	)
	if err != nil {
		t.Fatalf("ResolveProviderCredentials no-ip: %v", err)
	}
	if noipProvider.ZoneName != "flag.example.com" {
		t.Fatalf("ZoneName = %q, want flag override", noipProvider.ZoneName)
	}
	if noipProvider.APIKey != "config-key" {
		t.Fatalf("APIKey = %q, want config value", noipProvider.APIKey)
	}

	fromEnv, err := ResolveProviderCredentials(
		ProviderCloudflare,
		&fileConfig{},
		FetchOverrides{ZoneID: "flag-zone"},
	)
	if err != nil {
		t.Fatalf("ResolveProviderCredentials from env: %v", err)
	}
	if fromEnv.APIToken != "env-token" {
		t.Fatalf("APIToken = %q, want env value", fromEnv.APIToken)
	}
}
