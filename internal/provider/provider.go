package provider

import (
	"context"
	"fmt"

	"mio9/dddns/internal/config"
	"mio9/dddns/internal/provider/cloudflare"
	noipprovider "mio9/dddns/internal/provider/noip"
)

type Updater interface {
	Update(ctx context.Context, publicIP string) error
}

func New(cfg config.Provider) (Updater, error) {
	switch cfg.Type {
	case config.ProviderCloudflare:
		return cloudflare.New(cfg)
	case config.ProviderNoIP:
		return noipprovider.New(cfg)
	default:
		return nil, fmt.Errorf("unsupported provider type %q", cfg.Type)
	}
}
