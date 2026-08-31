package updater

import (
	"context"
	"errors"
	"fmt"

	"mio9/dddns/internal/config"
	"mio9/dddns/internal/ipcache"
	"mio9/dddns/internal/ipprovider"
	"mio9/dddns/internal/provider"
)

func Update(ctx context.Context, cfg *config.Config, configPath string) error {
	publicIP, err := ipprovider.FetchPublicIP(ctx, cfg.IPProvider.URL)
	if err != nil {
		return err
	}

	cachePath := ipcache.Path(configPath, cfg.IPCache.Path)
	cachedIP, err := ipcache.Load(cachePath)
	if err != nil {
		return err
	}
	if cachedIP == publicIP {
		fmt.Printf("unchanged: public IP still %s\n", publicIP)
		return nil
	}

	var updateErrors []error
	for index, providerConfig := range cfg.Providers {
		updater, err := provider.New(providerConfig)
		if err != nil {
			updateErrors = append(updateErrors, fmt.Errorf("providers[%d]: %w", index, err))
			continue
		}
		if err := updater.Update(ctx, publicIP); err != nil {
			updateErrors = append(updateErrors, fmt.Errorf("providers[%d] (%s): %w", index, providerConfig.Type, err))
		}
	}
	if err := errors.Join(updateErrors...); err != nil {
		return err
	}

	return ipcache.Save(cachePath, publicIP)
}
