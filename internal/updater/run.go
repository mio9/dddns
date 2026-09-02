package updater

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mio9/dddns/internal/config"
)

func Run(ctx context.Context, cfg *config.Config, configPath string) error {
	if cfg.UpdateInterval == 0 {
		return Update(ctx, cfg, configPath)
	}

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	runUpdate := func() {
		if err := Update(ctx, cfg, configPath); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}

	runUpdate()

	ticker := time.NewTicker(cfg.UpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			runUpdate()
		}
	}
}
