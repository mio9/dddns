package updater

import (
	"context"
	"errors"
	"fmt"

	"mio9/dddns/internal/config"
	"mio9/dddns/internal/ipprovider"
	"mio9/dddns/internal/provider"
)

func Update(ctx context.Context, cfg *config.Config) error {
	publicIP, err := ipprovider.FetchPublicIP(ctx, cfg.IPProvider.URL)
	if err != nil {
		return err
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

	return errors.Join(updateErrors...)
}
