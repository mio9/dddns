package fetch

import (
	"context"
	"fmt"

	"mio9/dddns/internal/config"
	"mio9/dddns/internal/ipprovider"
	"mio9/dddns/internal/provider"
)

func Run(ctx context.Context, providerType, configPath string, overrides config.FetchOverrides) error {
	cfg, fileCfg, err := config.LoadOrEmpty(configPath)
	if err != nil {
		return err
	}

	providerConfig, err := config.ResolveProviderCredentials(providerType, fileCfg, overrides)
	if err != nil {
		return err
	}

	publicIP, err := ipprovider.FetchPublicIP(ctx, cfg.IPProvider.URL)
	if err != nil {
		return err
	}

	fetcher, err := provider.NewFetcher(providerConfig)
	if err != nil {
		return err
	}

	records, err := fetcher.FetchMatchingARecords(ctx, publicIP)
	if err != nil {
		return err
	}
	if len(records) == 0 {
		return fmt.Errorf("no A records matching current IP %s found in %s", publicIP, providerType)
	}

	if err := config.MergeProviderRecords(configPath, providerType, providerConfig, records); err != nil {
		return err
	}

	fmt.Printf("fetched %d record(s) for %s into %s\n", len(records), providerType, configPath)
	return nil
}
