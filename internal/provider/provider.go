package provider

import (
	"context"
	"fmt"

	"mio9/dddns/internal/config"
	"mio9/dddns/internal/provider/cloudflare"
)

type Updater interface {
	Update(ctx context.Context, publicIP string) error
}

func New(cfg config.Provider) (Updater, error) {
	switch cfg.Type {
	case config.ProviderCloudflare:
		return cloudflare.New(cfg)
	default:
		return nil, fmt.Errorf("unsupported provider type %q", cfg.Type)
	}
}
