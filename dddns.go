package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"mio9/dddns/internal/config"
	"mio9/dddns/internal/fetch"
	"mio9/dddns/internal/updater"

	"github.com/urfave/cli/v3"
)

var defaultConfigNames = []string{"dddns.yaml", "dddns.yml"}

func resolveConfigPath(flagPath string) (string, error) {
	if flagPath != "" {
		return flagPath, nil
	}

	workDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve config path: %w", err)
	}

	for _, name := range defaultConfigNames {
		candidate := filepath.Join(workDir, name)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("config file not found: pass --config or create dddns.yaml or dddns.yml in %s", workDir)
}

func resolveConfigWritePath(flagPath string) (string, error) {
	if flagPath != "" {
		return flagPath, nil
	}

	workDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve config path: %w", err)
	}

	for _, name := range defaultConfigNames {
		candidate := filepath.Join(workDir, name)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return filepath.Join(workDir, defaultConfigNames[0]), nil
}

var Version string = "indev"

func main() {
	command := &cli.Command{
		Name:  "dddns",
		Usage: "Dynamic DNS updater",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "Show version",
			},
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Path to YAML config file (default: dddns.yaml or dddns.yml in current directory)",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Start DDNS updates (one-shot or timer mode from config)",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					configPath, err := resolveConfigPath(cmd.String("config"))
					if err != nil {
						return err
					}
					cfg, err := config.Load(configPath)
					if err != nil {
						return err
					}
					return updater.Run(ctx, cfg, configPath)
				},
			},
			{
				Name:      "fetch",
				Usage:     "Fetch A records matching current public IP from a provider into config",
				Arguments: []cli.Argument{
					&cli.StringArg{
						Name:      "provider",
						UsageText: "Provider name (cloudflare or no-ip)",
					},
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "zone-id",
						Usage: "Cloudflare zone ID (overrides config)",
					},
					&cli.StringFlag{
						Name:  "zone-name",
						Usage: "No-IP zone name (overrides config)",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					providerType := cmd.Args().First()
					if providerType == "" {
						return fmt.Errorf("provider is required")
					}
					if providerType != config.ProviderCloudflare && providerType != config.ProviderNoIP {
						return fmt.Errorf("unsupported provider %q", providerType)
					}

					configPath, err := resolveConfigWritePath(cmd.String("config"))
					if err != nil {
						return err
					}

					return fetch.Run(ctx, providerType, configPath, config.FetchOverrides{
						ZoneID:   cmd.String("zone-id"),
						ZoneName: cmd.String("zone-name"),
					})
				},
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Bool("version") {
				fmt.Printf("%s %s/%s\n", Version, runtime.GOOS, runtime.GOARCH)
				return nil
			}
			return cli.ShowAppHelp(cmd)
		},
	}

	if err := command.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
